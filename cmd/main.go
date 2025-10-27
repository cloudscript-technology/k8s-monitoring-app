package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"k8s-monitoring-app/internal/core"
	"k8s-monitoring-app/internal/env"
	"k8s-monitoring-app/internal/monitoring"
	"k8s-monitoring-app/internal/server"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/apmtracer"
	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
)

func main() {
	log.Init(context.Background())
	apmtracer.Init(context.Background())

	err := env.GetEnv()
	if err != nil {
		log.Error(context.Background(), err).Msg("Erro ao carregar variáveis de ambiente")
		os.Exit(1)
	}

	httpServer, err := server.NewHTTPServer(&core.ApiServiceConfiguration{})
	if err != nil {
		log.Error(context.Background(), err).Msg("Erro ao criar servidor")
		os.Exit(1)
	}
	defer httpServer.Postgres.Close()

	// Initialize and start monitoring service
	monitoringSvc, err := monitoring.NewMonitoringService(httpServer.Postgres)
	if err != nil {
		log.Error(context.Background(), err).Msg("Erro ao criar serviço de monitoramento")
		os.Exit(1)
	}

	if err := monitoringSvc.Start(); err != nil {
		log.Error(context.Background(), err).Msg("Erro ao iniciar serviço de monitoramento")
		os.Exit(1)
	}
	defer monitoringSvc.Stop()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Error(context.Background(), err).Msg("Erro ao iniciar servidor HTTP")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Info(context.Background()).Msg("Recebido sinal de interrupção, encerrando aplicação...")

	// Graceful shutdown is handled by deferred functions
}
