package database

import (
	"database/sql"
	"fmt"
)

func createTables(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS files (
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
}
