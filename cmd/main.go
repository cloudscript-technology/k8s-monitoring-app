package main

import (
	"os"
	"os/signal"
	"syscall"

	"k8s-monitoring-app/internal/core"
	"k8s-monitoring-app/internal/env"
	"k8s-monitoring-app/internal/monitoring"
	"k8s-monitoring-app/internal/server"

	"github.com/rs/zerolog/log"
)

func main() {
	err := env.GetEnv()
	if err != nil {
		log.Warn().Msg("Aviso: .env não encontrado, usando variáveis de ambiente do sistema")
	}

	httpServer, err := server.NewHTTPServer(&core.ApiServiceConfiguration{})
	if err != nil {
		log.Error().Msg("Erro ao criar servidor")
		os.Exit(1)
	}
	defer httpServer.SQLite.Close()

	// Initialize and start monitoring service
	monitoringSvc, err := monitoring.NewMonitoringService(httpServer.SQLite)
	if err != nil {
		log.Error().Msg("Erro ao criar serviço de monitoramento")
		os.Exit(1)
	}

	if err := monitoringSvc.Start(); err != nil {
		log.Error().Msg("Erro ao iniciar serviço de monitoramento")
		os.Exit(1)
	}
	defer monitoringSvc.Stop()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Error().Msg("Erro ao iniciar servidor HTTP")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Info().Msg("Recebido sinal de interrupção, encerrando aplicação...")

	// Graceful shutdown is handled by deferred functions
}
