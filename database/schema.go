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

func createViews(db *sql.DB) {
	_, err := db.Exec(`CREATE VIEW IF NOT EXISTS comic_files AS
SELECT f.id as file_id
FROM files f
         JOIN files_file_types fft ON f.id = fft.file_id
WHERE fft.file_type_id = 1;`)
	if err != nil {
		panic(fmt.Errorf("error creating view: %v", err))
	}

	_, err = db.Exec(`CREATE VIEW IF NOT EXISTS comic_files_hashes AS
SELECT file_id, hash_id
FROM files_hashes
WHERE file_id IN (SELECT file_id FROM comic_files);`)
	if err != nil {
		panic(fmt.Errorf("error creating view: %v", err))
	}
	_, err = db.Exec(`CREATE VIEW IF NOT EXISTS comic_files_duplicates AS
SELECT count(cfh.file_id), cfh.hash_id
FROM comic_files_hashes cfh
GROUP BY cfh.hash_id
HAVING count(cfh.file_id) > 1;`)
	if err != nil {
		panic(fmt.Errorf("error creating view: %v", err))
	}
}
