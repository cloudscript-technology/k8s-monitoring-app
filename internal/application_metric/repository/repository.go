package application_metric

import (
	"context"
	"database/sql"
	"fmt"

	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"

	"github.com/rs/zerolog/log"
)

type Repository interface {
	Get(ctx context.Context, id string, customFieldName ...string) (applicationMetricModel.ApplicationMetric, error)
	List(ctx context.Context) ([]applicationMetricModel.ApplicationMetric, error)
	ListByApplication(ctx context.Context, applicationID string) ([]applicationMetricModel.ApplicationMetric, error)
	Add(ctx context.Context, applicationMetric *applicationMetricModel.ApplicationMetric) error
	Update(ctx context.Context, applicationMetric *applicationMetricModel.ApplicationMetric) error
	Delete(ctx context.Context, id string) error
	GetDB() *sql.DB
}

type repository struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) GetDB() *sql.DB {
	return repo.db
}

func (repo *repository) Get(ctx context.Context, id string, customFieldName ...string) (applicationMetricModel.ApplicationMetric, error) {
	applicationMetric := applicationMetricModel.ApplicationMetric{}

	sqlString := `
	SELECT
		am.id, am.application_id, am.type_id, am.configuration, am.created_at, am.updated_at
	FROM 
		application_metrics am
	WHERE`

	if len(customFieldName) > 0 {
		sqlString = fmt.Sprintf("%s am.%s = $1", sqlString, customFieldName[0])
	} else {
		sqlString = fmt.Sprintf("%s am.id = $1", sqlString)
	}

	err := repo.db.QueryRowContext(ctx, sqlString, id).Scan(
		&applicationMetric.ID, &applicationMetric.ApplicationID, &applicationMetric.TypeID,
		&applicationMetric.Configuration, &applicationMetric.CreatedAt, &applicationMetric.UpdatedAt)

	if err != nil {
		return applicationMetric, err
	}

	return applicationMetric, nil
}

func (repo *repository) List(ctx context.Context) ([]applicationMetricModel.ApplicationMetric, error) {
	applicationMetrics := []applicationMetricModel.ApplicationMetric{}

	sqlString := `
	SELECT
		id, application_id, type_id, configuration, created_at, updated_at
	FROM
		application_metrics
	ORDER BY created_at DESC`

	rows, err := repo.db.QueryContext(ctx, sqlString)
	if err != nil {
		return applicationMetrics, err
	}
	defer rows.Close()

	for rows.Next() {
		applicationMetric := applicationMetricModel.ApplicationMetric{}
		err := rows.Scan(
			&applicationMetric.ID, &applicationMetric.ApplicationID, &applicationMetric.TypeID,
			&applicationMetric.Configuration, &applicationMetric.CreatedAt, &applicationMetric.UpdatedAt)
		if err != nil {
			return applicationMetrics, err
		}

		applicationMetrics = append(applicationMetrics, applicationMetric)
	}

	return applicationMetrics, nil
}

func (repo *repository) ListByApplication(ctx context.Context, applicationID string) ([]applicationMetricModel.ApplicationMetric, error) {
	applicationMetrics := []applicationMetricModel.ApplicationMetric{}

	sqlString := `
	SELECT
		id, application_id, type_id, configuration, created_at, updated_at
	FROM
		application_metrics
	WHERE application_id = $1
	ORDER BY created_at DESC`

	rows, err := repo.db.QueryContext(ctx, sqlString, applicationID)
	if err != nil {
		return applicationMetrics, err
	}
	defer rows.Close()

	for rows.Next() {
		applicationMetric := applicationMetricModel.ApplicationMetric{}
		err := rows.Scan(
			&applicationMetric.ID, &applicationMetric.ApplicationID, &applicationMetric.TypeID,
			&applicationMetric.Configuration, &applicationMetric.CreatedAt, &applicationMetric.UpdatedAt)
		if err != nil {
			return applicationMetrics, err
		}

		applicationMetrics = append(applicationMetrics, applicationMetric)
	}

	return applicationMetrics, nil
}

func (repo *repository) Add(ctx context.Context, applicationMetric *applicationMetricModel.ApplicationMetric) error {
	sqlString := `INSERT INTO application_metrics(
		application_id, type_id, configuration
		) VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := repo.db.QueryRowContext(ctx, sqlString,
		applicationMetric.ApplicationID, applicationMetric.TypeID, applicationMetric.Configuration,
	).Scan(&applicationMetric.ID, &applicationMetric.CreatedAt, &applicationMetric.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) Update(ctx context.Context, applicationMetric *applicationMetricModel.ApplicationMetric) error {
	var params []interface{}

	sqlString := `UPDATE application_metrics SET `

	if applicationMetric.TypeID != "" {
		sqlString = fmt.Sprintf("%s type_id = $%d, ", sqlString, len(params)+1)
		params = append(params, applicationMetric.TypeID)
	}
	if len(applicationMetric.Configuration) > 0 {
		sqlString = fmt.Sprintf("%s configuration = $%d, ", sqlString, len(params)+1)
		params = append(params, applicationMetric.Configuration)
	}
	if len(params) == 0 {
		log.Warn().Msg("no fields to update")
		return nil
	}

	// Always update the updated_at field
	sqlString = fmt.Sprintf("%s updated_at = now(), ", sqlString)

	sqlString = fmt.Sprintf("%s WHERE id = $%d", sqlString[:len(sqlString)-2], len(params)+1)
	params = append(params, applicationMetric.ID)

	result, err := repo.db.ExecContext(ctx, sqlString, params...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (repo *repository) Delete(ctx context.Context, id string) error {
	// Start a transaction to ensure both deletes happen atomically
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be no-op if tx.Commit() is called

	// First, delete all metric values associated with this metric
	deleteValuesSQL := `DELETE FROM application_metric_values WHERE application_metric_id = $1`
	valuesResult, err := tx.ExecContext(ctx, deleteValuesSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete metric values: %w", err)
	}

	// Log how many values were deleted
	valuesDeleted, _ := valuesResult.RowsAffected()
	if valuesDeleted > 0 {
		log.Info().
			Str("metric_id", id).
			Int64("values_deleted", valuesDeleted).
			Msg("deleted metric values before deleting metric")
	}

	// Then, delete the metric itself
	deleteMetricSQL := `DELETE FROM application_metrics WHERE id = $1`
	result, err := tx.ExecContext(ctx, deleteMetricSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete metric: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
