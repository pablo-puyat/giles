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

	stmt, err := db.Prepare("INSERT OR IGNORE INTO files(name, path, size) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("Error inserting preparing statement")
	}
	defer stmt.Close()

	for _, file := range files {
		_, err := stmt.Exec(file.Name, file.Path, file.Size)
		if err != nil {
			return err // Or log and continue with other files
		}
	}
	return nil
}

func (db *DB) GetFilesWithoutHash() (files []models.FileData, err error) {
	rows, err := db.Query("SELECT name, path, size FROM files WHERE hash IS NULL")
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
