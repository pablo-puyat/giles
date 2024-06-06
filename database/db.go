package database

import (
	"database/sql"
	"fmt"
	"giles/models"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbInstance *DB
	dbOnce     sync.Once
)

type DB struct {
	*sql.DB // Embed the *sql.DB to get its methods
}

func GetInstance() (*DB, error) {
	dbOnce.Do(func() {
		var err error
		sqlDB, err := sql.Open("sqlite3", "./giles.sqlite3")
		if err != nil {
			panic(fmt.Errorf("error opening database: %v", err))
		}

		dbInstance = &DB{sqlDB} // Correctly assign the *sql.DB to the embedded field

		_, err = dbInstance.Exec(`CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY,
            hash TEXT,
            name TEXT,
            path TEXT,
            size INTEGER,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`)
		if err != nil {
			panic(fmt.Errorf("error creating table: %v", err))
		}
	})
	return dbInstance, nil
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
