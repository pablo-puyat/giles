package database

import (
	"database/sql"
	"fmt"
)

type FileStore struct {
	db        *sql.DB
	BatchSize int
}

func New() (*FileStore, error) {
	db, err := sql.Open("sqlite3", "./giles.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTables(db)

	return &FileStore{
		db:        db,
		BatchSize: 100,
	}, nil
}
