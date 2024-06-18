package database

import (
	"database/sql"
	"fmt"
	"giles/models"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbInstance *DB
	dbOnce     sync.Once
)

type DB struct {
	*sql.DB
}

func GetInstance() (*DB, error) {
	dbOnce.Do(func() {
		var err error
		sqlDB, err := sql.Open("sqlite3", "./giles.db")
		if err != nil {
			panic(fmt.Errorf("error opening database: %v", err))
		}

		dbInstance = &DB{sqlDB}

		createTables(sqlDB)
	})
	return dbInstance, nil
}

func (db *DB) InsertFiles(files []models.FileData) error {
	if len(files) == 0 {
		return nil
	}

	stmt, err := db.Prepare("INSERT OR IGNORE INTO files(name, path, size) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("Error inserting preparing statement")
	}
	defer stmt.Close()

	for _, file := range files {
		_, err := stmt.Exec(file.Name, file.Path, file.Size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) GetFilesWithoutHash() (files []models.FileData, err error) {
	rows, err := db.Query("SELECT files.name, files.path, files.size" +
		"FROM files" +
		"LEFT JOIN files_hashes ON files.id = files_hashes.file_id" +
		"WHERE files_hashes.file_id IS NULL")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Error closing rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var file models.FileData
		err := rows.Scan(&file.Name, &file.Path, &file.Size)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return
}

func (db *DB) UpdateFileHash(path string, hash string) error {
	_, err := db.Exec("UPDATE files SET hash = ? WHERE path = ?", hash, path)
	return err
}

func (db *DB) UpdateFileHashBatch(files []models.FileData) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE files SET hash = ? WHERE path = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, file := range files {
		_, err := stmt.Exec(file.Hash, file.Path)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

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
