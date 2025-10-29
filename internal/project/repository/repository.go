package project

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	projectModel "k8s-monitoring-app/pkg/project/model"

	"github.com/rs/zerolog/log"
)

// generateUUID generates a simple UUID v4
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type Repository interface {
	Get(ctx context.Context, id string, customFieldName ...string) (projectModel.Project, error)
	List(ctx context.Context) ([]projectModel.Project, error)
	Add(ctx context.Context, project *projectModel.Project) error
	Update(ctx context.Context, project *projectModel.Project) error
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

func (repo *repository) Get(ctx context.Context, id string, customFieldName ...string) (projectModel.Project, error) {
	project := projectModel.Project{}

	sqlString := `
	SELECT
		p.id, p.name, p.description
	FROM 
		projects p
	WHERE`

	if len(customFieldName) > 0 {
		sqlString = fmt.Sprintf("%s p.%s = ?", sqlString, customFieldName[0])
	} else {
		sqlString = fmt.Sprintf("%s p.id = ?", sqlString)
	}

	err := repo.db.QueryRowContext(ctx, sqlString, id).Scan(
		&project.ID, &project.Name, &project.Description)

	if err != nil {
		return project, err
	}

	return project, nil
}

func (repo *repository) List(ctx context.Context) ([]projectModel.Project, error) {
	projects := []projectModel.Project{}

	sqlString := `
	SELECT
		id, name, description
	FROM
		projects
	ORDER BY name`

	rows, err := repo.db.QueryContext(ctx, sqlString)
	if err != nil {
		return projects, err
	}
	defer rows.Close()

	for rows.Next() {
		project := projectModel.Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description)
		if err != nil {
			return projects, err
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (repo *repository) Add(ctx context.Context, project *projectModel.Project) error {
	// Generate UUID for SQLite
	project.ID = generateUUID()

	sqlString := `INSERT INTO projects(
		id, name, description
		) VALUES (?, ?, ?)`

	_, err := repo.db.ExecContext(ctx, sqlString,
		project.ID, project.Name, project.Description,
	)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) Update(ctx context.Context, project *projectModel.Project) error {
	var params []interface{}

	sqlString := `UPDATE projects SET `

	if project.Name != "" {
		sqlString = fmt.Sprintf("%s name = ?, ", sqlString)
		params = append(params, project.Name)
	}
	if project.Description != "" {
		sqlString = fmt.Sprintf("%s description = ?, ", sqlString)
		params = append(params, project.Description)
	}
	if len(params) == 0 {
		log.Warn().Msg("no fields to update")
		return nil
	}

	sqlString = fmt.Sprintf("%s WHERE id = ?", sqlString[:len(sqlString)-2])
	params = append(params, project.ID)

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
	sqlString := `DELETE FROM projects WHERE id = ?`

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
