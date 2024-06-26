package database

import (
	"database/sql"
	"fmt"
	"giles/models"
	"log"
	"sync"
)

const (
	FilesWithoutHashSql   = "SELECT id, files.path FROM files LEFT JOIN files_hashes ON files.id = files_hashes.file_id WHERE files_hashes.file_id IS NULL"
	InserFileSql          = "INSERT OR IGNORE INTO files (name, path, size) VALUES (?, ?, ?)"
	InsertFileIdHashIdSql = "INSERT INTO files_hashes (file_id, hash_id) VALUES (?, ?)"
	InsertHashSql         = "INSERT OR IGNORE INTO hashes (hash) VALUES (?);"
)

var (
	instance *DataStore
	once     sync.Once
)

type DataStore struct {
	DB *sql.DB
}

func NewDataStore() *DataStore {
	once.Do(func() {
		db, err := sql.Open("sqlite3", "./giles.db")
		if err != nil {
			panic(fmt.Errorf("error opening database: %v", err))
		}
		createTables(db)
		instance = &DataStore{DB: db}
	})
	return instance
}

func (ds *DataStore) GetFilesWithoutHash() (files []models.FileData, err error) {
	rows, err := ds.DB.Query(FilesWithoutHashSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file models.FileData
		err := rows.Scan(&file.Id, &file.Path)
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

func (ds *DataStore) InsertFile(file models.FileData) (models.FileData, error) {
	_, err := ds.DB.Exec(InserFileSql, file.Name, file.Path, file.Size)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
	return file, err
}

func (ds *DataStore) InsertFileIdHashId(file models.FileData) (models.FileData, error) {
	print("Inserting ", file.Id, " ", file.HashId, "\n")
	_, err := ds.DB.Exec(InsertFileIdHashIdSql, file.Id, file.HashId)
	if err != nil {
		log.Fatalf("Error inserting file and hash id: %v", err)
	}
	return file, err
}

func (ds *DataStore) InsertHash(file models.FileData) (models.FileData, error) {
	result, err := ds.DB.Exec(InsertHashSql, file.Hash)
	if err != nil {
		log.Fatalf("Error inserting hash: %v", err)
	}
	l, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Error inserting hash: %v", err)
	}
	if l == 0 {
		rows, err := ds.DB.Query("SELECT id FROM hashes WHERE hash = ?", file.Hash)
		if err != nil {
			log.Fatalf("Error querying hash: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&file.HashId)
			if err != nil {
				log.Fatalf("Error scanning hash: %v", err)
			}
		}
	} else {
		file.HashId = l
	}
	return file, err
}
