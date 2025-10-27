package application_metric_value

import (
	"context"
	"database/sql"
	"fmt"

	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"
)

type Repository interface {
	Get(ctx context.Context, id string) (applicationMetricValueModel.ApplicationMetricValue, error)
	ListByApplicationMetric(ctx context.Context, applicationMetricID string, limit int) ([]applicationMetricValueModel.ApplicationMetricValue, error)
	Add(ctx context.Context, applicationMetricValue *applicationMetricValueModel.ApplicationMetricValue) error
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

func (repo *repository) Get(ctx context.Context, id string) (applicationMetricValueModel.ApplicationMetricValue, error) {
	applicationMetricValue := applicationMetricValueModel.ApplicationMetricValue{}

	sqlString := `
	SELECT
		amv.id, amv.application_metric_id, amv.value, amv.created_at, amv.updated_at
	FROM 
		application_metric_values amv
	WHERE amv.id = $1`

	err := repo.db.QueryRowContext(ctx, sqlString, id).Scan(
		&applicationMetricValue.ID, &applicationMetricValue.ApplicationMetricID,
		&applicationMetricValue.Value, &applicationMetricValue.CreatedAt, &applicationMetricValue.UpdatedAt)

	if err != nil {
		return applicationMetricValue, err
	}

	return applicationMetricValue, nil
}

func (repo *repository) ListByApplicationMetric(ctx context.Context, applicationMetricID string, limit int) ([]applicationMetricValueModel.ApplicationMetricValue, error) {
	applicationMetricValues := []applicationMetricValueModel.ApplicationMetricValue{}

	sqlString := `
	SELECT
		id, application_metric_id, value, created_at, updated_at
	FROM
		application_metric_values
	WHERE application_metric_id = $1
	ORDER BY created_at DESC`

	if limit > 0 {
		sqlString = fmt.Sprintf("%s LIMIT %d", sqlString, limit)
	}

	rows, err := repo.db.QueryContext(ctx, sqlString, applicationMetricID)
	if err != nil {
		return applicationMetricValues, err
	}
	defer rows.Close()

	for rows.Next() {
		applicationMetricValue := applicationMetricValueModel.ApplicationMetricValue{}
		err := rows.Scan(
			&applicationMetricValue.ID, &applicationMetricValue.ApplicationMetricID,
			&applicationMetricValue.Value, &applicationMetricValue.CreatedAt, &applicationMetricValue.UpdatedAt)
		if err != nil {
			return applicationMetricValues, err
		}

		applicationMetricValues = append(applicationMetricValues, applicationMetricValue)
	}

	return applicationMetricValues, nil
}

func (repo *repository) Add(ctx context.Context, applicationMetricValue *applicationMetricValueModel.ApplicationMetricValue) error {
	sqlString := `INSERT INTO application_metric_values(
		application_metric_id, value
		) VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := repo.db.QueryRowContext(ctx, sqlString,
		applicationMetricValue.ApplicationMetricID, applicationMetricValue.Value,
	).Scan(&applicationMetricValue.ID, &applicationMetricValue.CreatedAt, &applicationMetricValue.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}
