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
			path TEXT UNIQUE,
			size INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		panic(fmt.Errorf("error creating table: %v", err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS hashes (
			id INTEGER PRIMARY KEY,
			hash TEXT UNIQUE
		)`)
	if err != nil {
		panic(fmt.Errorf("error creating table: %v", err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files_hashes (
			file_id INTEGER,
			hash_id INTEGER,
			PRIMARY KEY (file_id, hash_id),
			FOREIGN KEY (file_id) REFERENCES files(id),
			FOREIGN KEY (hash_id) REFERENCES hashes(id)
		)`)
	if err != nil {
		panic(fmt.Errorf("error creating table: %v", err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS file_types (
			id INTEGER PRIMARY KEY,
			type TEXT UNIQUE
		)`)
	if err != nil {
		panic(fmt.Errorf("error creating table: %v", err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS file_extensions_file_types (
			file_extension TEXT,
			file_type_id TEXT UNIQUE,
			PRIMARY KEY (file_extension, file_type_id),
			FOREIGN KEY (file_type_id) REFERENCES file_types(id)
		)`)
	if err != nil {
		panic(fmt.Errorf("error creating table: %v", err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files_file_types (
			file_id INTEGER,
			file_type_id INTEGER,
			PRIMARY KEY (file_id, file_type_id),
			FOREIGN KEY (file_id) REFERENCES files(id),
			FOREIGN KEY (file_type_id) REFERENCES file_types(id)
		)`)
	if err != nil {
		panic(fmt.Errorf("error creating table: %v", err))
	}
}
