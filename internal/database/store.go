package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

var (
	instance *DataStore
	once     sync.Once
)

type DataStore struct {
	db        *sql.DB
	BatchSize int
}

func NewDataStore() (*DataStore, error) {
	db, err := sql.Open("sqlite3", "./giles.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTables(db)
	createViews(db)

	return &DataStore{
		db:        db,
		BatchSize: 100,
	}, nil
}

func (ds *DataStore) GetDuplicates() (files []FileData, err error) {
	rows, err := ds.db.Query(`
		SELECT files.* 
		FROM comic_files_hashes cfh, files 
		WHERE cfh.hash_id in (select hash_id from comic_files_duplicates) AND 
			cfh.file_id = files.id ORDER BY hash;
	`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var file FileData
		err := rows.Scan(&file.Id, &file.Path, &file.Name, &file.Size)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return files, err
}

func (ds *DataStore) GetFilesWithoutHash() (files []FileData, err error) {
	rows, err := ds.db.Query(`SELECT id, files.path, size FROM files WHERE hash IS NULL;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file FileData
		err := rows.Scan(&file.Id, &file.Path, &file.Size)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return files, err
}

func (ds *DataStore) GetFilesFrom(source string) (files []FileData, err error) {
	rows, err := ds.db.Query("SELECT id, path, size, hash FROM files WHERE path LIKE ?", source+"%")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var file FileData
		err := rows.Scan(&file.Id, &file.Path, &file.Size, &file.Hash)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return files, err
}

func (ds *DataStore) InsertFile(file FileData) (FileData, error) {
	_, err := ds.db.Exec(` INSERT OR IGNORE INTO files (name, path, size) VALUES (?, ?, ?);`, file.Name, file.Path, file.Size)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
	return file, err
}

func (ds *DataStore) InsertHash(files []FileData) error {
	tx, err := ds.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	for _, f := range files {
		_, err := tx.Exec(`UPDATE files set hash = ? WHERE id = ?`, f.Hash, f.Id)
		if err != nil {
			tx.Rollback()
			log.Printf("Error adding hash: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return err
}
