package application_metric

import (
	"context"
	"database/sql"
	"fmt"

	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
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
		log.Warn(ctx).Msg("no fields to update")
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
	sqlString := `DELETE FROM application_metrics WHERE id = $1`

	result, err := repo.db.ExecContext(ctx, sqlString, id)
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
