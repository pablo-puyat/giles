package database

import (
	"database/sql"
	"fmt"
	"giles/models"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewConnection() (*DB, error) {
	db, err := sql.Open("sqlite3", "./giles.sqlite3")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY,
        hash TEXT,
        name TEXT,
        path TEXT,
        size INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}

	return &DB{db}, nil
}

func (db *DB) InsertFiles(files []models.FileData) error {
	if len(files) == 0 {
		return nil
	}

	// Start building the SQL query
	query := "INSERT OR IGNORE INTO files(name, path, size) VALUES "

	// Create a slice to hold the values for the placeholders
	values := make([]interface{}, 0, len(files)*3)

	// Add a placeholder for each file data
	for _, file := range files {
		query += "(?, ?, ?),"
		values = append(values, file.Name, file.Path, file.Size)
	}

	// Remove the trailing comma
	query = query[:len(query)-1]

	// Execute the query
	_, err := db.Exec(query, values...)
	return err
}
