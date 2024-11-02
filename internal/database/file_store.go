package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	db        *sql.DB
	BatchSize int
	Path      string
}

func New(dbPath string) (*FileStore, error) {
	path := getPath(dbPath)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Println("using database: ", path)

	return &FileStore{
		db:        db,
		BatchSize: 100,
	}, nil
}

func getPath(dbPath string) string {
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			pwd, err := os.Getwd()
			if err != nil {
				pwd = "."
			}
			return filepath.Join(pwd, "giles.db")
		}
		return filepath.Join(homeDir, ".local", "share", "giles", "db.sqlite")
	}
	return dbPath
}

func (fs *FileStore) Close() error {
	if fs.db != nil {
		return fs.db.Close()
	}
	return nil
}
