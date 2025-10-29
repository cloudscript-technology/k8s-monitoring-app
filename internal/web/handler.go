package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"k8s-monitoring-app/internal/core"
	"k8s-monitoring-app/internal/security"
	serverModel "k8s-monitoring-app/internal/server/model"
	applicationModel "k8s-monitoring-app/pkg/application/model"
	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"
	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"
	projectModel "k8s-monitoring-app/pkg/project/model"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	templates *template.Template
}

func NewHandler() (*Handler, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"div": func(a, b interface{}) float64 {
			aFloat, _ := toFloat64(a)
			bFloat, _ := toFloat64(b)
			if bFloat == 0 {
				return 0
			}
			return aFloat / bFloat
		},
		"add": func(a, b interface{}) float64 {
			aFloat, _ := toFloat64(a)
			bFloat, _ := toFloat64(b)
			return aFloat + bFloat
		},
		"sub": func(a, b interface{}) float64 {
			aFloat, _ := toFloat64(a)
			bFloat, _ := toFloat64(b)
			return aFloat - bFloat
		},
		"firstChar": func(s string) string {
			if len(s) == 0 {
				return "?"
			}
			return string(s[0])
		},
	}

	templates, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		templates: templates,
	}, nil
}

// Helper function to convert interface{} to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	default:
		return 0, nil
	}
}

// Dashboard renders the main dashboard page
func (h *Handler) Dashboard(sc *core.HTTPServerContext) error {
	// Get user info from context (set by auth middleware)
	userEmail := sc.Get("user_email")
	userName := sc.Get("user_name")
	userPicture := sc.Get("user_picture")

	data := map[string]interface{}{
		"Title":       "Dashboard",
		"UserEmail":   userEmail,
		"UserName":    userName,
		"UserPicture": userPicture,
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "layout.html", data); err != nil {
		log.Error().Err(err).Msg("error executing template")
		return err
	}

	return nil
}

// DeleteMetric deletes a metric and returns success response for HTMX
func (h *Handler) DeleteMetric(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if id == "" {
		return sc.String(http.StatusBadRequest, "ID is required")
	}

	err := serverModel.ServerRepos.ApplicationMetric.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sc.String(http.StatusNotFound, "Metric not found")
		}
		log.Error().Err(err).Str("id", id).Msg("error deleting metric")
		return sc.String(http.StatusInternalServerError, "Error deleting metric")
	}

	log.Info().Str("id", id).Msg("metric deleted successfully")
	return sc.String(http.StatusOK, "Metric deleted successfully")
}

// DeleteApplication deletes an application and returns success response for HTMX
func (h *Handler) DeleteApplication(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if id == "" {
		return sc.String(http.StatusBadRequest, "ID is required")
	}

	err := serverModel.ServerRepos.Application.Delete(ctx, id)
	if err != nil {
		log.Error().Str("id", id).Msg("error deleting application")
		return sc.String(http.StatusInternalServerError, "Error deleting application")
	}

	log.Info().Str("id", id).Msg("application deleted successfully")
	return sc.String(http.StatusOK, "Application deleted successfully")
}

// DeleteProject deletes a project and returns success response for HTMX
func (h *Handler) DeleteProject(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if id == "" {
		return sc.String(http.StatusBadRequest, "ID is required")
	}

	err := serverModel.ServerRepos.Project.Delete(ctx, id)
	if err != nil {
		log.Error().Str("id", id).Msg("error deleting project")
		return sc.String(http.StatusInternalServerError, "Error deleting project")
	}

	log.Info().Str("id", id).Msg("project deleted successfully")
	return sc.String(http.StatusOK, "Project deleted successfully")
}

// RenderCadastroProjects renders the project registration page
func (h *Handler) RenderCadastroProjects(sc *core.HTTPServerContext) error {
	// Get user info from context (set by auth middleware)
	userEmail := sc.Get("user_email")
	userName := sc.Get("user_name")
	userPicture := sc.Get("user_picture")

	data := map[string]interface{}{
		"Title":       "Cadastro de Projetos",
		"UserEmail":   userEmail,
		"UserName":    userName,
		"UserPicture": userPicture,
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "cadastro-projetos.html", data); err != nil {
		log.Error().Err(err).Msg("error executing cadastro-projetos template")
		return err
	}

	return nil
}

// RenderCadastroApplications renders the application registration page
func (h *Handler) RenderCadastroApplications(sc *core.HTTPServerContext) error {
	// Get user info from context (set by auth middleware)
	userEmail := sc.Get("user_email")
	userName := sc.Get("user_name")
	userPicture := sc.Get("user_picture")

	data := map[string]interface{}{
		"Title":       "Cadastro de Aplicações",
		"UserEmail":   userEmail,
		"UserName":    userName,
		"UserPicture": userPicture,
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "cadastro-aplicacoes.html", data); err != nil {
		log.Error().Err(err).Msg("error executing cadastro-aplicacoes template")
		return err
	}

	return nil
}

// RenderCadastroMetrics renders the metrics registration page
func (h *Handler) RenderCadastroMetrics(sc *core.HTTPServerContext) error {
	// Get user info from context (set by auth middleware)
	userEmail := sc.Get("user_email")
	userName := sc.Get("user_name")
	userPicture := sc.Get("user_picture")

	data := map[string]interface{}{
		"Title":       "Cadastro de Métricas",
		"UserEmail":   userEmail,
		"UserName":    userName,
		"UserPicture": userPicture,
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "cadastro-metricas.html", data); err != nil {
		log.Error().Err(err).Msg("error executing cadastro-metricas template")
		return err
	}

	return nil
}

// GetProjectsOptions returns projects as select options for forms
func (h *Handler) GetProjectsOptions(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	log.Info().Msg("GetProjectsOptions called")

	// Get all projects
	projects, err := serverModel.ServerRepos.Project.List(ctx)
	if err != nil {
		log.Error().Msg("error listing projects")
		return sc.String(http.StatusInternalServerError, "Error loading projects")
	}

	log.Info().Int("count", len(projects)).Msg("Projects retrieved for options")

	// Generate HTML options
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	// Write default option
	sc.Response().Writer.Write([]byte(`<option value="">Selecione um projeto</option>`))

	// Write project options
	for _, project := range projects {
		optionHTML := fmt.Sprintf(`<option value="%s">%s</option>`, project.ID, project.Name)
		sc.Response().Writer.Write([]byte(optionHTML))
	}

	return nil
}

// GetApplicationsOptions returns applications as select options for forms
func (h *Handler) GetApplicationsOptions(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	// Get query parameter for filtering by project
	projectID := sc.QueryParam("project_id")

	log.Info().
		Str("project_id", projectID).
		Msg("GetApplicationsOptions called")

	var applications []applicationModel.Application
	var err error

	// Get applications filtered by project if specified
	if projectID != "" {
		applications, err = serverModel.ServerRepos.Application.ListByProject(ctx, projectID)
	} else {
		applications, err = serverModel.ServerRepos.Application.List(ctx)
	}

	if err != nil {
		log.Error().Msg("error listing applications")
		return sc.String(http.StatusInternalServerError, "Error loading applications")
	}

	log.Info().
		Int("count", len(applications)).
		Str("project_filter", projectID).
		Msg("Applications retrieved for options")

	// Generate HTML options
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	// Write default option
	sc.Response().Writer.Write([]byte(`<option value="">Selecione uma aplicação</option>`))

	// Write application options with project name
	for _, app := range applications {
		// Get project details
		project, err := serverModel.ServerRepos.Project.Get(ctx, app.ProjectID)
		projectName := "N/A"
		if err == nil {
			projectName = project.Name
		}

		optionHTML := fmt.Sprintf(`<option value="%s">%s - %s</option>`, app.ID, projectName, app.Name)
		sc.Response().Writer.Write([]byte(optionHTML))
	}

	return nil
}

// GetMetricTypesOptions returns metric types as select options for forms
func (h *Handler) GetMetricTypesOptions(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	log.Info().Msg("GetMetricTypesOptions called")

	// Get all metric types
	metricTypes, err := serverModel.ServerRepos.MetricType.List(ctx)
	if err != nil {
		log.Error().Msg("error listing metric types")
		return sc.String(http.StatusInternalServerError, "Error loading metric types")
	}

	log.Info().Int("count", len(metricTypes)).Msg("Metric types retrieved for options")

	// Generate HTML options
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	// Write default option
	sc.Response().Writer.Write([]byte(`<option value="">Selecione um tipo de métrica</option>`))

	// Write metric type options
	for _, metricType := range metricTypes {
		optionHTML := fmt.Sprintf(`<option value="%s">%s</option>`, metricType.ID, metricType.Name)
		sc.Response().Writer.Write([]byte(optionHTML))
	}

	return nil
}

// GetMetricConfigurationFields returns configuration fields for a specific metric type
func (h *Handler) GetMetricConfigurationFields(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	metricTypeID := sc.Param("id")

	if metricTypeID == "" {
		return sc.String(http.StatusBadRequest, "Metric type ID is required")
	}

	log.Info().Str("metric_type_id", metricTypeID).Msg("GetMetricConfigurationFields called")

	// Get metric type details
	metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metricTypeID)
	if err != nil {
		log.Error().Msg("error getting metric type")
		return sc.String(http.StatusNotFound, "Metric type not found")
	}

	// Set content type to HTML
	sc.Response().Header().Set("Content-Type", "text/html")

	// Generate configuration fields based on metric type
	fieldsHTML := h.generateConfigurationFields(metricType.Name)
	sc.Response().Writer.Write([]byte(fieldsHTML))

	return nil
}

// generateConfigurationFields generates HTML form fields based on metric type
func (h *Handler) generateConfigurationFields(metricTypeName string) string {
	var fieldsHTML string

	switch metricTypeName {
	case "HealthCheck":
		fieldsHTML = `
			<div class="form-group">
				<label for="health_check_url">URL do Health Check:</label>
				<input type="url" id="health_check_url" name="health_check_url" required 
					   placeholder="http://app.namespace.svc.cluster.local:8080/health">
			</div>
			<div class="form-group">
				<label for="method">Método HTTP:</label>
				<select id="method" name="method" required>
					<option value="GET">GET</option>
					<option value="POST">POST</option>
					<option value="HEAD">HEAD</option>
				</select>
			</div>
			<div class="form-group">
				<label for="expected_status">Status HTTP Esperado:</label>
				<input type="number" id="expected_status" name="expected_status" value="200" required min="100" max="599">
			</div>
			<div class="form-group">
				<label for="timeout_seconds">Timeout (segundos):</label>
				<input type="number" id="timeout_seconds" name="timeout_seconds" value="10" required min="1" max="300">
			</div>`

	case "PodStatus":
		fieldsHTML = `
			<div class="form-group">
				<label for="pod_label_selector">Seletor de Labels do Pod:</label>
				<input type="text" id="pod_label_selector" name="pod_label_selector" required 
					   placeholder="app=minha-aplicacao">
			</div>
			<div class="form-group">
				<label for="container_name">Nome do Container:</label>
				<input type="text" id="container_name" name="container_name" required 
					   placeholder="web">
			</div>`

	case "PodMemoryUsage", "PodCpuUsage":
		fieldsHTML = `
			<div class="form-group">
				<label for="pod_label_selector">Seletor de Labels do Pod:</label>
				<input type="text" id="pod_label_selector" name="pod_label_selector" required 
					   placeholder="app=minha-aplicacao">
			</div>
			<div class="form-group">
				<label for="container_name">Nome do Container:</label>
				<input type="text" id="container_name" name="container_name" required 
					   placeholder="web">
			</div>`

	case "PvcUsage":
		fieldsHTML = `
			<div class="form-group">
				<label for="pvc_name">Nome do PVC:</label>
				<input type="text" id="pvc_name" name="pvc_name" required 
					   placeholder="minha-aplicacao-data">
			</div>
			<div class="form-group">
				<label for="pod_label_selector">Seletor de Labels do Pod:</label>
				<input type="text" id="pod_label_selector" name="pod_label_selector" required 
					   placeholder="app=minha-aplicacao">
			</div>
			<div class="form-group">
				<label for="container_name">Nome do Container:</label>
				<input type="text" id="container_name" name="container_name" required 
					   placeholder="web">
			</div>
			<div class="form-group">
				<label for="pvc_mount_path">Caminho de Montagem do PVC:</label>
				<input type="text" id="pvc_mount_path" name="pvc_mount_path" required 
					   placeholder="/data">
			</div>`

	case "PodActiveNodes":
		fieldsHTML = `
			<div class="form-group">
				<label for="pod_label_selector">Seletor de Labels do Pod:</label>
				<input type="text" id="pod_label_selector" name="pod_label_selector" required 
					   placeholder="app=minha-aplicacao">
			</div>`

	case "RedisConnection":
		fieldsHTML = `
			<div class="form-group">
				<label for="connection_host">Host do Redis:</label>
				<input type="text" id="connection_host" name="connection_host" required 
					   placeholder="redis.default.svc.cluster.local">
			</div>
			<div class="form-group">
				<label for="connection_port">Porta:</label>
				<input type="number" id="connection_port" name="connection_port" value="6379" required min="1" max="65535">
			</div>
			<div class="form-group">
				<label for="connection_password">Senha (opcional):</label>
				<input type="password" id="connection_password" name="connection_password" 
					   placeholder="senha-do-redis">
			</div>
			<div class="form-group">
				<label for="connection_db">Database:</label>
				<input type="number" id="connection_db" name="connection_db" value="0" required min="0" max="15">
			</div>
			<div class="form-group">
				<label for="connection_ssl">SSL:</label>
				<select id="connection_ssl" name="connection_ssl" required>
					<option value="false">Não</option>
					<option value="true">Sim</option>
				</select>
			</div>
			<div class="form-group">
				<label for="connection_timeout">Timeout (segundos):</label>
				<input type="number" id="connection_timeout" name="connection_timeout" value="5" required min="1" max="300">
			</div>`

	case "PostgreSQLConnection":
		fieldsHTML = `
			<div class="form-group">
				<label for="connection_host">Host do PostgreSQL:</label>
				<input type="text" id="connection_host" name="connection_host" required 
					   placeholder="postgres.default.svc.cluster.local">
			</div>
			<div class="form-group">
				<label for="connection_port">Porta:</label>
				<input type="number" id="connection_port" name="connection_port" value="5432" required min="1" max="65535">
			</div>
			<div class="form-group">
				<label for="connection_username">Usuário:</label>
				<input type="text" id="connection_username" name="connection_username" required 
					   placeholder="usuario">
			</div>
			<div class="form-group">
				<label for="connection_password">Senha:</label>
				<input type="password" id="connection_password" name="connection_password" required 
					   placeholder="senha">
			</div>
			<div class="form-group">
				<label for="connection_database">Database:</label>
				<input type="text" id="connection_database" name="connection_database" required 
					   placeholder="minha_base">
			</div>
			<div class="form-group">
				<label for="connection_ssl">SSL:</label>
				<select id="connection_ssl" name="connection_ssl" required>
					<option value="false">Não</option>
					<option value="true">Sim</option>
				</select>
			</div>
			<div class="form-group">
				<label for="connection_timeout">Timeout (segundos):</label>
				<input type="number" id="connection_timeout" name="connection_timeout" value="10" required min="1" max="300">
			</div>`

	case "MongoDBConnection":
		fieldsHTML = `
			<div class="form-group">
				<label for="connection_host">Host do MongoDB:</label>
				<input type="text" id="connection_host" name="connection_host" required 
					   placeholder="mongodb.default.svc.cluster.local">
			</div>
			<div class="form-group">
				<label for="connection_port">Porta:</label>
				<input type="number" id="connection_port" name="connection_port" value="27017" required min="1" max="65535">
			</div>
			<div class="form-group">
				<label for="connection_username">Usuário:</label>
				<input type="text" id="connection_username" name="connection_username" required 
					   placeholder="admin">
			</div>
			<div class="form-group">
				<label for="connection_password">Senha:</label>
				<input type="password" id="connection_password" name="connection_password" required 
					   placeholder="senha">
			</div>
			<div class="form-group">
				<label for="connection_database">Database:</label>
				<input type="text" id="connection_database" name="connection_database" required 
					   placeholder="minha_base">
			</div>
			<div class="form-group">
				<label for="connection_auth_source">Auth Source:</label>
				<input type="text" id="connection_auth_source" name="connection_auth_source" value="admin" required 
					   placeholder="admin">
			</div>
			<div class="form-group">
				<label for="connection_ssl">SSL:</label>
				<select id="connection_ssl" name="connection_ssl" required>
					<option value="false">Não</option>
					<option value="true">Sim</option>
				</select>
			</div>
			<div class="form-group">
				<label for="connection_timeout">Timeout (segundos):</label>
				<input type="number" id="connection_timeout" name="connection_timeout" value="5" required min="1" max="300">
			</div>`

	case "MySQLConnection":
		fieldsHTML = `
			<div class="form-group">
				<label for="connection_host">Host do MySQL:</label>
				<input type="text" id="connection_host" name="connection_host" required 
					   placeholder="mysql.default.svc.cluster.local">
			</div>
			<div class="form-group">
				<label for="connection_port">Porta:</label>
				<input type="number" id="connection_port" name="connection_port" value="3306" required min="1" max="65535">
			</div>
			<div class="form-group">
				<label for="connection_username">Usuário:</label>
				<input type="text" id="connection_username" name="connection_username" required 
					   placeholder="root">
			</div>
			<div class="form-group">
				<label for="connection_password">Senha:</label>
				<input type="password" id="connection_password" name="connection_password" required 
					   placeholder="senha">
			</div>
			<div class="form-group">
				<label for="connection_database">Database:</label>
				<input type="text" id="connection_database" name="connection_database" required 
					   placeholder="minha_base">
			</div>
			<div class="form-group">
				<label for="connection_ssl">SSL:</label>
				<select id="connection_ssl" name="connection_ssl" required>
					<option value="false">Não</option>
					<option value="true">Sim</option>
				</select>
			</div>
			<div class="form-group">
				<label for="connection_timeout">Timeout (segundos):</label>
				<input type="number" id="connection_timeout" name="connection_timeout" value="5" required min="1" max="300">
			</div>`

	case "KongConnection":
		fieldsHTML = `
			<div class="form-group">
				<label for="connection_host">Host do Kong Admin:</label>
				<input type="text" id="connection_host" name="connection_host" required 
					   placeholder="kong-admin.default.svc.cluster.local">
			</div>
			<div class="form-group">
				<label for="connection_port">Porta:</label>
				<input type="number" id="connection_port" name="connection_port" value="8001" required min="1" max="65535">
			</div>
			<div class="form-group">
				<label for="kong_admin_url">URL Admin do Kong:</label>
				<input type="url" id="kong_admin_url" name="kong_admin_url" required 
					   placeholder="http://kong-admin.default.svc.cluster.local:8001">
			</div>
			<div class="form-group">
				<label for="connection_username">Usuário (opcional):</label>
				<input type="text" id="connection_username" name="connection_username" 
					   placeholder="admin">
			</div>
			<div class="form-group">
				<label for="connection_password">Senha (opcional):</label>
				<input type="password" id="connection_password" name="connection_password" 
					   placeholder="senha">
			</div>
			<div class="form-group">
				<label for="connection_ssl">SSL:</label>
				<select id="connection_ssl" name="connection_ssl" required>
					<option value="false">Não</option>
					<option value="true">Sim</option>
				</select>
			</div>
			<div class="form-group">
				<label for="connection_timeout">Timeout (segundos):</label>
				<input type="number" id="connection_timeout" name="connection_timeout" value="5" required min="1" max="300">
			</div>`

	case "IngressCertificate":
		fieldsHTML = `
			<div class="form-group">
				<label for="ingress_name">Nome do Ingress:</label>
				<input type="text" id="ingress_name" name="ingress_name" required 
					   placeholder="minha-aplicacao-ingress">
			</div>
			<div class="form-group">
				<label for="ingress_namespace">Namespace (opcional):</label>
				<input type="text" id="ingress_namespace" name="ingress_namespace" 
					   placeholder="default">
			</div>
			<div class="form-group">
				<label for="tls_secret_name">Nome do Secret TLS (opcional):</label>
				<input type="text" id="tls_secret_name" name="tls_secret_name" 
					   placeholder="minha-aplicacao-tls">
			</div>
			<div class="form-group">
				<label for="warning_days">Dias de Aviso:</label>
				<input type="number" id="warning_days" name="warning_days" value="30" required min="1" max="365">
			</div>`

	case "KafkaConsumerLag":
		fieldsHTML = `
			<div class="form-group">
				<label for="kafka_bootstrap_servers">Servidores Bootstrap do Kafka:</label>
				<input type="text" id="kafka_bootstrap_servers" name="kafka_bootstrap_servers" required 
					   placeholder="kafka:9092">
			</div>
			<div class="form-group">
				<label for="kafka_consumer_group">Grupo de Consumidores:</label>
				<input type="text" id="kafka_consumer_group" name="kafka_consumer_group" required 
					   placeholder="meu-grupo-consumidor">
			</div>
			<div class="form-group">
				<label for="kafka_topic">Tópico (opcional):</label>
				<input type="text" id="kafka_topic" name="kafka_topic" 
					   placeholder="meu-topico">
			</div>
			<div class="form-group">
				<label for="kafka_lag_threshold">Limite de Lag:</label>
				<input type="number" id="kafka_lag_threshold" name="kafka_lag_threshold" value="1000" required min="0">
			</div>
			<div class="form-group">
				<label for="kafka_security_protocol">Protocolo de Segurança (opcional):</label>
				<select id="kafka_security_protocol" name="kafka_security_protocol">
					<option value="">Nenhum</option>
					<option value="PLAINTEXT">PLAINTEXT</option>
					<option value="SASL_PLAINTEXT">SASL_PLAINTEXT</option>
					<option value="SASL_SSL">SASL_SSL</option>
					<option value="SSL">SSL</option>
				</select>
			</div>
			<div class="form-group">
				<label for="kafka_sasl_mechanism">Mecanismo SASL (opcional):</label>
				<select id="kafka_sasl_mechanism" name="kafka_sasl_mechanism">
					<option value="">Nenhum</option>
					<option value="PLAIN">PLAIN</option>
					<option value="SCRAM-SHA-256">SCRAM-SHA-256</option>
					<option value="SCRAM-SHA-512">SCRAM-SHA-512</option>
				</select>
			</div>
			<div class="form-group">
				<label for="kafka_sasl_username">Usuário SASL (opcional):</label>
				<input type="text" id="kafka_sasl_username" name="kafka_sasl_username" 
					   placeholder="usuario-kafka">
			</div>
			<div class="form-group">
				<label for="kafka_sasl_password">Senha SASL (opcional):</label>
				<input type="password" id="kafka_sasl_password" name="kafka_sasl_password" 
					   placeholder="senha-kafka">
			</div>`

	default:
		fieldsHTML = `<p>Configuração não disponível para este tipo de métrica.</p>`
	}

	return fieldsHTML
}

func (h *Handler) GetProjectsList(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	projects, err := serverModel.ServerRepos.Project.List(ctx)
	if err != nil {
		log.Error().Msg("error listing projects")
		return sc.String(http.StatusInternalServerError, "Error loading projects")
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	for _, project := range projects {
		if err := h.templates.ExecuteTemplate(sc.Response().Writer, "project-list-item", project); err != nil {
			log.Error().Msg("error executing template")
			return err
		}
	}

	return nil
}

// GetApplicationsList returns all applications for HTMX partial
func (h *Handler) GetApplicationsList(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	// Get query parameters for filtering
	projectID := sc.QueryParam("project_id")
	nameFilter := sc.QueryParam("name")

	log.Info().
		Str("project_id", projectID).
		Str("name_filter", nameFilter).
		Msg("GetApplicationsList called with filters")

	applications, err := serverModel.ServerRepos.Application.List(ctx)
	if err != nil {
		log.Error().Msg("error listing applications")
		return sc.String(http.StatusInternalServerError, "Error loading applications")
	}

	// Apply filters
	var filteredApplications []applicationModel.Application
	for _, app := range applications {
		// Filter by project ID if specified
		if projectID != "" && app.ProjectID != projectID {
			continue
		}

		// Filter by name if specified (case-insensitive partial match)
		if nameFilter != "" {
			nameFilterLower := strings.ToLower(nameFilter)
			appNameLower := strings.ToLower(app.Name)
			if !strings.Contains(appNameLower, nameFilterLower) {
				continue
			}
		}

		filteredApplications = append(filteredApplications, app)
	}

	log.Info().
		Int("total_applications", len(applications)).
		Int("filtered_applications", len(filteredApplications)).
		Msg("Applications filtered")

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	// Create a struct to pass formatted data to template
	type ApplicationDisplay struct {
		ID          string
		ProjectID   string
		Name        string
		Description string
		Namespace   string
		ProjectName string
	}

	for _, app := range filteredApplications {
		display := ApplicationDisplay{
			ID:          app.ID,
			ProjectID:   app.ProjectID,
			Name:        app.Name,
			Description: app.Description,
			Namespace:   app.Namespace,
		}

		// Get project details
		project, err := serverModel.ServerRepos.Project.Get(ctx, app.ProjectID)
		if err != nil {
			log.Error().Str("project_id", app.ProjectID).Msg("error getting project")
			display.ProjectName = "N/A"
		} else {
			display.ProjectName = project.Name
		}

		if err := h.templates.ExecuteTemplate(sc.Response().Writer, "application-list-item", display); err != nil {
			log.Error().Msg("error executing template")
			return err
		}
	}

	return nil
}

// GetMetricsList returns all metrics for HTMX partial
func (h *Handler) GetMetricsList(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	// Get query parameters for filtering
	projectID := sc.QueryParam("project_id")
	applicationID := sc.QueryParam("application_id")
	metricTypeID := sc.QueryParam("metric_type_id")

	log.Info().
		Str("project_id", projectID).
		Str("application_id", applicationID).
		Str("metric_type_id", metricTypeID).
		Msg("GetMetricsList called with filters")

	metrics, err := serverModel.ServerRepos.ApplicationMetric.List(ctx)
	if err != nil {
		log.Error().Msg("error listing metrics")
		return sc.String(http.StatusInternalServerError, "Error loading metrics")
	}

	// Apply filters
	var filteredMetrics []applicationMetricModel.ApplicationMetric
	for _, metric := range metrics {
		// Filter by metric type ID if specified
		if metricTypeID != "" && metric.TypeID != metricTypeID {
			continue
		}

		// Filter by application ID if specified
		if applicationID != "" && metric.ApplicationID != applicationID {
			continue
		}

		// Filter by project ID if specified (need to get application first)
		if projectID != "" {
			application, err := serverModel.ServerRepos.Application.Get(ctx, metric.ApplicationID)
			if err != nil || application.ProjectID != projectID {
				continue
			}
		}

		filteredMetrics = append(filteredMetrics, metric)
	}

	log.Info().
		Int("total_metrics", len(metrics)).
		Int("filtered_metrics", len(filteredMetrics)).
		Msg("Metrics filtered")

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	// Create a struct to pass formatted data to template
	type MetricDisplay struct {
		ID              string
		TypeID          string
		ApplicationID   string
		ApplicationName string
		ProjectName     string
		MetricTypeName  string
		Configuration   string
	}

	for _, metric := range filteredMetrics {
		display := MetricDisplay{
			ID:            metric.ID,
			TypeID:        metric.TypeID,
			ApplicationID: metric.ApplicationID,
			Configuration: string(security.RedactSensitiveFieldsRaw(metric.Configuration)),
		}

		// Get metric type details
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metric.TypeID)
		if err != nil {
			log.Error().Str("metric_type_id", metric.TypeID).Msg("error getting metric type")
			display.MetricTypeName = "N/A"
		} else {
			display.MetricTypeName = metricType.Name
		}

		// Get application details
		application, err := serverModel.ServerRepos.Application.Get(ctx, metric.ApplicationID)
		if err != nil {
			log.Error().Str("application_id", metric.ApplicationID).Msg("error getting application")
			display.ApplicationName = "N/A"
			display.ProjectName = "N/A"
		} else {
			display.ApplicationName = application.Name

			// Get project details
			project, err := serverModel.ServerRepos.Project.Get(ctx, application.ProjectID)
			if err != nil {
				log.Error().Str("project_id", application.ProjectID).Msg("error getting project")
				display.ProjectName = "N/A"
			} else {
				display.ProjectName = project.Name
			}
		}

		if err := h.templates.ExecuteTemplate(sc.Response().Writer, "metric-list-item", display); err != nil {
			log.Error().Msg("error executing template")
			return err
		}
	}

	return nil
}

// RenderLogin renders the login page
func (h *Handler) RenderLogin(sc *core.HTTPServerContext) error {
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "login.html", nil); err != nil {
		log.Error().Err(err).Msg("error executing login template")
		return err
	}

	return nil
}

// RenderAuthError renders the authentication error page
func (h *Handler) RenderAuthError(sc *core.HTTPServerContext) error {
	reason := sc.QueryParam("reason")

	messages := map[string]string{
		"invalid_state":      "Invalid authentication state. Please try again.",
		"no_code":            "No authorization code received. Please try again.",
		"exchange_failed":    "Failed to exchange authorization code. Please try again.",
		"user_info_failed":   "Failed to retrieve user information. Please try again.",
		"email_not_verified": "Your email address is not verified with Google.",
		"domain_not_allowed": "Your email domain is not authorized to access this application.",
		"session_failed":     "Failed to create session. Please try again.",
		"unauthorized":       "You are not authorized to access this application.",
	}

	message, ok := messages[reason]
	if !ok {
		message = "An unknown error occurred during authentication."
	}

	data := map[string]interface{}{
		"Message": message,
		"Reason":  reason,
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusUnauthorized)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "auth-error.html", data); err != nil {
		log.Error().Err(err).Msg("error executing auth-error template")
		return err
	}

	return nil
}

type ProjectWithApplications struct {
	Project      projectModel.Project
	Applications []applicationModel.Application
}

// GetProjects returns all projects with their applications
func (h *Handler) GetProjects(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	log.Info().Msg("GetProjects called")

	// Get all projects
	projects, err := serverModel.ServerRepos.Project.List(ctx)
	if err != nil {
		log.Error().Msg("error listing projects")
		return sc.String(http.StatusInternalServerError, "Error loading projects")
	}

	log.Info().Int("count", len(projects)).Msg("Projects retrieved")

	// Get applications for each project
	var projectsWithApps []ProjectWithApplications
	for _, project := range projects {
		apps, err := serverModel.ServerRepos.Application.ListByProject(ctx, project.ID)
		if err != nil {
			log.Error().Str("project_id", project.ID).Msg("error listing applications")
			continue
		}

		projectsWithApps = append(projectsWithApps, ProjectWithApplications{
			Project:      project,
			Applications: apps,
		})
	}

	// Render the projects template using pre-loaded templates
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	for _, projectWithApps := range projectsWithApps {
		if err := h.templates.ExecuteTemplate(sc.Response().Writer, "project-card", projectWithApps); err != nil {
			log.Error().Msg("error executing template")
			return err
		}
	}

	return nil
}

type ApplicationMetricsView struct {
	ApplicationID          string
	ApplicationName        string
	ApplicationDescription string
	ApplicationNamespace   string
	MetricsByType          map[string]*MetricWithValue
}

type MetricWithValue struct {
	MetricID      string
	MetricTypeID  string
	Configuration map[string]interface{}
	LatestValue   *MetricValueParsed
}

type MetricValueParsed struct {
	ID                  string
	ApplicationMetricID string
	Value               map[string]interface{}
	CreatedAt           string
	UpdatedAt           string
}

// GetApplicationMetrics returns metrics for a specific application
func (h *Handler) GetApplicationMetrics(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationID := sc.Param("id")
	if applicationID == "" {
		return sc.String(http.StatusBadRequest, "Application ID required")
	}

	// Get application details
	application, err := serverModel.ServerRepos.Application.Get(ctx, applicationID)
	if err != nil {
		log.Error().Str("application_id", applicationID).Msg("error getting application")
		return sc.String(http.StatusNotFound, "Application not found")
	}

	// Get all metrics for this application
	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		log.Error().Str("application_id", applicationID).Msg("error listing application metrics")
		return sc.String(http.StatusInternalServerError, "Error loading metrics")
	}

	// Build metrics map organized by type
	metricsByType := make(map[string]*MetricWithValue)

	for _, metric := range applicationMetrics {
		// Get metric type details
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metric.TypeID)
		if err != nil {
			log.Error().Str("metric_type_id", metric.TypeID).Msg("error getting metric type")
			continue
		}

		metricWithValue := &MetricWithValue{
			MetricID:     metric.ID,
			MetricTypeID: metric.TypeID,
		}

		// Parse Configuration from JSON to map (with redaction applied first)
		var config map[string]interface{}
		redacted := security.RedactSensitiveFieldsRaw(metric.Configuration)
		if err = json.Unmarshal(redacted, &config); err == nil {
			metricWithValue.Configuration = config
		}

		// Get the latest value (limit 1)
		values, err := serverModel.ServerRepos.ApplicationMetricValue.ListByApplicationMetric(ctx, metric.ID, 1)
		if err == nil && len(values) > 0 {
			// Parse the Value from JSON to map
			var valueMap map[string]interface{}
			if err := json.Unmarshal(values[0].Value, &valueMap); err == nil {
				metricWithValue.LatestValue = &MetricValueParsed{
					ID:                  values[0].ID,
					ApplicationMetricID: values[0].ApplicationMetricID,
					Value:               valueMap,
					CreatedAt:           values[0].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt:           values[0].UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				}
			}
		}

		metricsByType[metricType.Name] = metricWithValue
	}

	data := ApplicationMetricsView{
		ApplicationID:          application.ID,
		ApplicationName:        application.Name,
		ApplicationDescription: application.Description,
		ApplicationNamespace:   application.Namespace,
		MetricsByType:          metricsByType,
	}

	log.Info().
		Str("app_name", data.ApplicationName).
		Int("metrics_count", len(metricsByType)).
		Msg("Rendering application metrics")

	// Debug: log metric types
	for metricTypeName, metric := range metricsByType {
		hasValue := metric.LatestValue != nil
		log.Info().
			Str("metric_type", metricTypeName).
			Bool("has_value", hasValue).
			Msg("Metric in map")
	}

	// Render the application metrics template using pre-loaded templates with custom functions
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "application-metrics", data); err != nil {
		log.Error().
			Str("app_id", applicationID).
			Str("app_name", data.ApplicationName).
			Msg("error executing template - check template syntax")
		return err
	}

	return nil
}

// Helper function to get metric value by type
func getMetricValueByType(ctx context.Context, applicationID, metricTypeName string) *applicationMetricValueModel.ApplicationMetricValue {
	// Get all metrics for the application
	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		return nil
	}

	// Find the metric with the specified type
	for _, metric := range applicationMetrics {
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metric.TypeID)
		if err != nil {
			continue
		}

		if metricType.Name == metricTypeName {
			// Get the latest value
			values, err := serverModel.ServerRepos.ApplicationMetricValue.ListByApplicationMetric(ctx, metric.ID, 1)
			if err == nil && len(values) > 0 {
				return &values[0]
			}
		}
	}

	return nil
}
