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

	if err := createTables(db); err != nil {
		return nil, err
	}

	fmt.Println("Using database: ", databasePath)

	return &FileStore{
		db:        db,
		BatchSize: 100,
	}, nil
}

func (fs *FileStore) Close() error {
	if fs.db != nil {
		return fs.db.Close()
	}
	return nil
}
