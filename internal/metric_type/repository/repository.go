package metric_type

import (
	"context"
	"database/sql"
	"fmt"

	metricTypeModel "k8s-monitoring-app/pkg/metric_type/model"
)

type Repository interface {
	Get(ctx context.Context, id string, customFieldName ...string) (metricTypeModel.MetricType, error)
	List(ctx context.Context) ([]metricTypeModel.MetricType, error)
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

func (repo *repository) Get(ctx context.Context, id string, customFieldName ...string) (metricTypeModel.MetricType, error) {
	metricType := metricTypeModel.MetricType{}

	sqlString := `
	SELECT
		mt.id, mt.name, mt.description, mt.created_at, mt.updated_at
	FROM 
		metric_types mt
	WHERE`

	if len(customFieldName) > 0 {
		sqlString = fmt.Sprintf("%s mt.%s = $1", sqlString, customFieldName[0])
	} else {
		sqlString = fmt.Sprintf("%s mt.id = $1", sqlString)
	}

	err := repo.db.QueryRowContext(ctx, sqlString, id).Scan(
		&metricType.ID, &metricType.Name, &metricType.Description,
		&metricType.CreatedAt, &metricType.UpdatedAt)

	if err != nil {
		return metricType, err
	}

	return metricType, nil
}

func (repo *repository) List(ctx context.Context) ([]metricTypeModel.MetricType, error) {
	metricTypes := []metricTypeModel.MetricType{}

	sqlString := `
	SELECT
		id, name, description, created_at, updated_at
	FROM
		metric_types
	ORDER BY name`

	rows, err := repo.db.QueryContext(ctx, sqlString)
	if err != nil {
		return metricTypes, err
	}
	defer rows.Close()

	for rows.Next() {
		metricType := metricTypeModel.MetricType{}
		err := rows.Scan(
			&metricType.ID, &metricType.Name, &metricType.Description,
			&metricType.CreatedAt, &metricType.UpdatedAt)
		if err != nil {
			return metricTypes, err
		}

		metricTypes = append(metricTypes, metricType)
	}

	return metricTypes, nil
}
