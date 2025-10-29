package core

import (
	"database/sql"
	"errors"
	"fmt"
	"k8s-monitoring-app/internal/env"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"go.elastic.co/apm/module/apmsql"
)

func ConnectDatabase() (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(env.DB_PATH)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %s", err.Error())
	}

	db, err := apmsql.Open("sqlite3", env.DB_PATH)
	if err != nil {
		return db, fmt.Errorf("failed to connect to database: %s | %s", err.Error(), env.DB_PATH)
	}

	// Enable foreign keys for SQLite
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return db, fmt.Errorf("failed to enable foreign keys: %s", err.Error())
	}

	var count int
	err = db.QueryRow("SELECT 1 AS count;").Scan(&count)
	if err != nil {
		return db, fmt.Errorf("failed to test database: %s | %s", err.Error(), env.DB_PATH)
	}
	if count != 1 {
		return nil, errors.New("failed to return test database")
	}
	return db, err
}
