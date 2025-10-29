package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	applicationModel "k8s-monitoring-app/pkg/application/model"

	"github.com/rs/zerolog/log"
)

// generateUUID generates a simple UUID v4
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type Repository interface {
	Get(ctx context.Context, id string, customFieldName ...string) (applicationModel.Application, error)
	List(ctx context.Context) ([]applicationModel.Application, error)
	ListByProject(ctx context.Context, projectID string) ([]applicationModel.Application, error)
	Add(ctx context.Context, application *applicationModel.Application) error
	Update(ctx context.Context, application *applicationModel.Application) error
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

func (repo *repository) Get(ctx context.Context, id string, customFieldName ...string) (applicationModel.Application, error) {
	application := applicationModel.Application{}

	sqlString := `
	SELECT
		a.id, a.project_id, a.name, a.description, a.namespace, a.created_at, a.updated_at
	FROM 
		applications a
	WHERE`

	if len(customFieldName) > 0 {
		sqlString = fmt.Sprintf("%s a.%s = ?", sqlString, customFieldName[0])
	} else {
		sqlString = fmt.Sprintf("%s a.id = ?", sqlString)
	}

	err := repo.db.QueryRowContext(ctx, sqlString, id).Scan(
		&application.ID, &application.ProjectID, &application.Name, &application.Description,
		&application.Namespace, &application.CreatedAt, &application.UpdatedAt)

	if err != nil {
		return application, err
	}

	return application, nil
}

func (repo *repository) List(ctx context.Context) ([]applicationModel.Application, error) {
	applications := []applicationModel.Application{}

	sqlString := `
	SELECT
		id, project_id, name, description, namespace, created_at, updated_at
	FROM
		applications
	ORDER BY name`

	rows, err := repo.db.QueryContext(ctx, sqlString)
	if err != nil {
		return applications, err
	}
	defer rows.Close()

	for rows.Next() {
		application := applicationModel.Application{}
		err := rows.Scan(
			&application.ID, &application.ProjectID, &application.Name, &application.Description,
			&application.Namespace, &application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return applications, err
		}

		applications = append(applications, application)
	}

	return applications, nil
}

func (repo *repository) ListByProject(ctx context.Context, projectID string) ([]applicationModel.Application, error) {
	applications := []applicationModel.Application{}

	sqlString := `
	SELECT
		id, project_id, name, description, namespace, created_at, updated_at
	FROM
		applications
	WHERE project_id = ?
	ORDER BY name`

	rows, err := repo.db.QueryContext(ctx, sqlString, projectID)
	if err != nil {
		return applications, err
	}
	defer rows.Close()

	for rows.Next() {
		application := applicationModel.Application{}
		err := rows.Scan(
			&application.ID, &application.ProjectID, &application.Name, &application.Description,
			&application.Namespace, &application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return applications, err
		}

		applications = append(applications, application)
	}

	return applications, nil
}

func (repo *repository) Add(ctx context.Context, application *applicationModel.Application) error {
	// Generate UUID and timestamps for SQLite
	application.ID = generateUUID()
	now := time.Now()
	application.CreatedAt = now
	application.UpdatedAt = now
	
	sqlString := `INSERT INTO applications(
		id, project_id, name, description, namespace, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := repo.db.ExecContext(ctx, sqlString,
		application.ID, application.ProjectID, application.Name, application.Description, 
		application.Namespace, application.CreatedAt, application.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) Update(ctx context.Context, application *applicationModel.Application) error {
	var params []interface{}

	sqlString := `UPDATE applications SET `

	if application.Name != "" {
		sqlString = fmt.Sprintf("%s name = ?, ", sqlString)
		params = append(params, application.Name)
	}
	if application.Description != "" {
		sqlString = fmt.Sprintf("%s description = ?, ", sqlString)
		params = append(params, application.Description)
	}
	if application.Namespace != "" {
		sqlString = fmt.Sprintf("%s namespace = ?, ", sqlString)
		params = append(params, application.Namespace)
	}
	if len(params) == 0 {
		log.Warn().Msg("no fields to update")
		return nil
	}

	// Add updated_at timestamp
	sqlString = fmt.Sprintf("%s updated_at = ?, ", sqlString)
	params = append(params, time.Now())

	sqlString = fmt.Sprintf("%s WHERE id = ?", sqlString[:len(sqlString)-2])
	params = append(params, application.ID)

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
	sqlString := `DELETE FROM applications WHERE id = ?`

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