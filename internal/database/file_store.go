package database

import (
	"database/sql"
	"fmt"
)

type FileStore struct {
	db        *sql.DB
	BatchSize int
}

func New(databasePath string) (*FileStore, error) {
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTables(db)

	fmt.Println("Using database: ", databasePath)
	return &FileStore{
		db:        db,
		BatchSize: 100,
	}, nil
}
