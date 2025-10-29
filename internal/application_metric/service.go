package application_metric

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s-monitoring-app/internal/core"
    "k8s-monitoring-app/internal/security"
	serverModel "k8s-monitoring-app/internal/server/model"
	model "k8s-monitoring-app/pkg/application_metric/model"

	"github.com/rs/zerolog/log"
)

type service struct{}

func NewService() model.Service {
	return &service{}
}

func (s *service) Get(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	id := sc.Param("id")

	if len(id) == 0 {
		log.Error().Msg("id is empty")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	applicationMetric, err := serverModel.ServerRepos.ApplicationMetric.Get(ctx, id)
	if err != nil {
		log.Error().Msg("error getting application metric")
		return sc.String(http.StatusNotFound, "application metric not found")
	}

    // Redact sensitive configuration fields before returning
    applicationMetric.Configuration = security.RedactSensitiveFieldsRaw(applicationMetric.Configuration)
    return sc.JSON(http.StatusOK, applicationMetric)
}

func (s *service) List(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.List(ctx)
	if err != nil {
		log.Error().Msg("error listing application metrics")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

    // Redact sensitive configuration fields in each item before returning
    for i := range applicationMetrics {
        applicationMetrics[i].Configuration = security.RedactSensitiveFieldsRaw(applicationMetrics[i].Configuration)
    }
    return sc.JSON(http.StatusOK, applicationMetrics)
}

func (s *service) ListByApplication(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationID := sc.Param("application_id")

	if len(applicationID) == 0 {
		log.Error().Msg("application_id is empty")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		log.Error().Msg("error listing application metrics by application")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

    // Redact sensitive configuration fields before returning
    for i := range applicationMetrics {
        applicationMetrics[i].Configuration = security.RedactSensitiveFieldsRaw(applicationMetrics[i].Configuration)
    }
    return sc.JSON(http.StatusOK, applicationMetrics)
}

func (s *service) Add(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationMetric := model.ApplicationMetric{}
	if err := sc.Bind(&applicationMetric); err != nil {
		log.Error().Msg("error binding application metric")
		return sc.String(http.StatusBadRequest, "invalid request body")
	}

	// Validate that the application exists
	_, err := serverModel.ServerRepos.Application.Get(ctx, applicationMetric.ApplicationID)
	if err != nil {
		log.Error().Msg("error getting application")
		return sc.String(http.StatusBadRequest, "application not found")
	}

	// Validate that the metric type exists
	metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, applicationMetric.TypeID)
	if err != nil {
		log.Error().Msg("error getting metric type")
		return sc.String(http.StatusBadRequest, "metric type not found")
	}

	// Check if a metric of this type already exists for this application
	existingMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationMetric.ApplicationID)
	if err != nil {
		log.Error().Msg("error listing application metrics")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	for _, existing := range existingMetrics {
		if existing.TypeID == applicationMetric.TypeID {
			log.Warn().
				Str("application_id", applicationMetric.ApplicationID).
				Str("metric_type", metricType.Name).
				Str("existing_metric_id", existing.ID).
				Msg("metric of this type already exists for this application")
			return sc.JSON(http.StatusConflict, map[string]interface{}{
				"error":              "metric already exists",
				"message":            "A metric of type '" + metricType.Name + "' already exists for this application",
				"existing_metric_id": existing.ID,
				"metric_type":        metricType.Name,
			})
		}
	}

	// Validate configuration schema early to avoid runtime collector failures
	var cfg model.Configuration
	if err := json.Unmarshal(applicationMetric.Configuration, &cfg); err != nil {
		log.Warn().Err(err).
			Str("application_id", applicationMetric.ApplicationID).
			Str("metric_type", metricType.Name).
			Msg("invalid metric configuration payload")
		return sc.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid configuration",
			"message": "Configuration JSON does not match expected schema for '" + metricType.Name + "'",
			"details": err.Error(),
		})
	}

	// Additional validation for connection-type metrics to avoid silent misconfigurations
	if err := validateConfigByType(metricType.Name, cfg); err != nil {
		log.Warn().Err(err).
			Str("application_id", applicationMetric.ApplicationID).
			Str("metric_type", metricType.Name).
			Msg("invalid connection configuration")
		return sc.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid configuration",
			"message": err.Error(),
		})
	}

	if err := serverModel.ServerRepos.ApplicationMetric.Add(ctx, &applicationMetric); err != nil {
		log.Error().Msg("error add application metric")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

    // Redact sensitive configuration fields before returning
    applicationMetric.Configuration = security.RedactSensitiveFieldsRaw(applicationMetric.Configuration)
    return sc.JSON(http.StatusCreated, applicationMetric)
}

func (s *service) Update(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	// First get the existing application metric to check it exists
	existingMetric, err := serverModel.ServerRepos.ApplicationMetric.Get(ctx, id)
	if err != nil {
		log.Error().Msg("error getting application metric")
		return sc.String(http.StatusNotFound, "application metric not found")
	}

	applicationMetric := model.ApplicationMetric{}
	if err = sc.Bind(&applicationMetric); err != nil {
		log.Error().Msg("error binding application metric")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}
	applicationMetric.ID = id

	// Validate that the metric type exists if it's being changed
	if applicationMetric.TypeID != "" {
		metricType, err2 := serverModel.ServerRepos.MetricType.Get(ctx, applicationMetric.TypeID)
		if err2 != nil {
			log.Error().Msgf("error getting metric type: %s", err2.Error())
			return sc.String(http.StatusBadRequest, "metric type not found")
		}

		// If the type is being changed, check if another metric of the new type already exists
		if applicationMetric.TypeID != existingMetric.TypeID {
			existingMetrics, err3 := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, existingMetric.ApplicationID)
			if err3 != nil {
				log.Error().Msgf("error listing application metrics: %s", err3.Error())
				return sc.String(http.StatusInternalServerError, "internal server error")
			}

			for _, existing := range existingMetrics {
				// Skip the metric being updated
				if existing.ID == id {
					continue
				}

				if existing.TypeID == applicationMetric.TypeID {
					log.Warn().
						Str("application_id", existingMetric.ApplicationID).
						Str("metric_type", metricType.Name).
						Str("existing_metric_id", existing.ID).
						Msg("metric of this type already exists for this application")
					return sc.JSON(http.StatusConflict, map[string]interface{}{
						"error":              "metric already exists",
						"message":            "A metric of type '" + metricType.Name + "' already exists for this application",
						"existing_metric_id": existing.ID,
						"metric_type":        metricType.Name,
					})
				}
			}
		}
	}

	// Determine metric type for configuration validation
	metricTypeForValidationID := existingMetric.TypeID
	if applicationMetric.TypeID != "" {
		metricTypeForValidationID = applicationMetric.TypeID
	}
	metricTypeForValidation, err := serverModel.ServerRepos.MetricType.Get(ctx, metricTypeForValidationID)
	if err != nil {
		log.Error().Msg("error getting metric type for validation")
		return sc.String(http.StatusBadRequest, "metric type not found")
	}

	// If configuration is provided, validate schema early
	if len(applicationMetric.Configuration) > 0 {
		var cfg model.Configuration
		if err := json.Unmarshal(applicationMetric.Configuration, &cfg); err != nil {
			log.Warn().Err(err).
				Str("application_id", existingMetric.ApplicationID).
				Str("metric_type", metricTypeForValidation.Name).
				Str("application_metric_id", id).
				Msg("invalid metric configuration payload on update")
			return sc.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "invalid configuration",
				"message": "Configuration JSON does not match expected schema for '" + metricTypeForValidation.Name + "'",
				"details": err.Error(),
			})
		}

		// Additional validation for connection-type metrics
		if err := validateConfigByType(metricTypeForValidation.Name, cfg); err != nil {
			log.Warn().Err(err).
				Str("application_id", existingMetric.ApplicationID).
				Str("metric_type", metricTypeForValidation.Name).
				Str("application_metric_id", id).
				Msg("invalid connection configuration on update")
			return sc.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "invalid configuration",
				"message": err.Error(),
			})
		}
	}

	if err := serverModel.ServerRepos.ApplicationMetric.Update(ctx, &applicationMetric); err != nil {
		log.Error().Msg("error updating application metric")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

    // Redact sensitive configuration fields before returning
    applicationMetric.Configuration = security.RedactSensitiveFieldsRaw(applicationMetric.Configuration)
    return sc.JSON(http.StatusOK, applicationMetric)
}

func (s *service) Delete(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if len(id) == 0 {
		log.Error().Msg("id is empty")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}

	err := serverModel.ServerRepos.ApplicationMetric.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sc.String(http.StatusNotFound, "application metric not found")
		}
		log.Error().Err(err).Str("id", id).Msg("error deleting application metric")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, map[string]bool{"success": true})
}

// validateConfigByType performs required-field checks for connection-type metrics.
// It prevents silent zero-values (e.g., port=0, timeout=0) from reaching collectors.
func validateConfigByType(metricTypeName string, cfg model.Configuration) error {
	switch metricTypeName {
	case "PostgreSQLConnection":
		if cfg.ConnectionHost == "" {
			return fmt.Errorf("connection_host is required for %s", metricTypeName)
		}
		if cfg.ConnectionPort <= 0 {
			return fmt.Errorf("connection_port must be a positive integer for %s", metricTypeName)
		}
		if cfg.ConnectionUsername == "" {
			return fmt.Errorf("connection_username is required for %s", metricTypeName)
		}
		if cfg.ConnectionDatabase == "" {
			return fmt.Errorf("connection_database is required for %s", metricTypeName)
		}
		if cfg.ConnectionTimeout <= 0 {
			return fmt.Errorf("connection_timeout must be a positive integer for %s", metricTypeName)
		}
	case "MySQLConnection":
		if cfg.ConnectionHost == "" {
			return fmt.Errorf("connection_host is required for %s", metricTypeName)
		}
		if cfg.ConnectionPort <= 0 {
			return fmt.Errorf("connection_port must be a positive integer for %s", metricTypeName)
		}
		if cfg.ConnectionUsername == "" {
			return fmt.Errorf("connection_username is required for %s", metricTypeName)
		}
		if cfg.ConnectionDatabase == "" {
			return fmt.Errorf("connection_database is required for %s", metricTypeName)
		}
		if cfg.ConnectionTimeout <= 0 {
			return fmt.Errorf("connection_timeout must be a positive integer for %s", metricTypeName)
		}
	case "MongoDBConnection":
		if cfg.ConnectionHost == "" {
			return fmt.Errorf("connection_host is required for %s", metricTypeName)
		}
		if cfg.ConnectionPort <= 0 {
			return fmt.Errorf("connection_port must be a positive integer for %s", metricTypeName)
		}
		if cfg.ConnectionUsername == "" {
			return fmt.Errorf("connection_username is required for %s", metricTypeName)
		}
		if cfg.ConnectionPassword == "" {
			return fmt.Errorf("connection_password is required for %s", metricTypeName)
		}
		if cfg.ConnectionDatabase == "" {
			return fmt.Errorf("connection_database is required for %s", metricTypeName)
		}
		if cfg.ConnectionTimeout <= 0 {
			return fmt.Errorf("connection_timeout must be a positive integer for %s", metricTypeName)
		}
	case "RedisConnection":
		if cfg.ConnectionHost == "" {
			return fmt.Errorf("connection_host is required for %s", metricTypeName)
		}
		if cfg.ConnectionPort <= 0 {
			return fmt.Errorf("connection_port must be a positive integer for %s", metricTypeName)
		}
		if cfg.ConnectionTimeout <= 0 {
			return fmt.Errorf("connection_timeout must be a positive integer for %s", metricTypeName)
		}
	case "KongConnection":
 		// Either explicit admin URL or host+port
 		if cfg.KongAdminURL == "" {
 			if cfg.ConnectionHost == "" || cfg.ConnectionPort <= 0 {
 				return fmt.Errorf("kong_admin_url or connection_host+connection_port are required for %s", metricTypeName)
 			}
 		}
 		if cfg.ConnectionTimeout <= 0 {
 			return fmt.Errorf("connection_timeout must be a positive integer for %s", metricTypeName)
 		}
	default:
 		// Non-connection metric types: no additional checks here
 		return nil
 	}
 	return nil
}
