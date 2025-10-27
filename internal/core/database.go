package core

import (
	"database/sql"
	"errors"
	"fmt"

	"k8s-monitoring-app/internal/env"

	"go.elastic.co/apm/module/apmsql"
)

func ConnectDatabase() (*sql.DB, error) {
	db, err := apmsql.Open("postgres", env.DB_CONNECTION_STRING)
	if err != nil {
		return db, fmt.Errorf("failed to connect to database: %s | %s", err.Error(), env.DB_CONNECTION_STRING)
	}

	var count int
	err = db.QueryRow("SELECT 1 AS count;").Scan(&count)
	if err != nil {
		return db, fmt.Errorf("failed to test database: %s | %s", err.Error(), env.DB_CONNECTION_STRING)
	}
	if count != 1 {
		return nil, errors.New("failed to return test database")
	}
	return db, err
}
