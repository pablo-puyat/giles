package database

import (
	"database/sql"
	"fmt"
	"giles/models"
	"log"
	"sync"
)

const (
	FilesWithoutHashSql = "SELECT id, files.path, size FROM files WHERE hash IS NULL"
	InsertFileSql       = "INSERT OR IGNORE INTO files (name, path, size) VALUES (?, ?, ?)"
	AddHashSql          = "UPDATE files set hash = ? WHERE id = ?"
	DuplicatesSql       = "SELECT files.* FROM comic_files_hashes cfh, files WHERE cfh.hash_id in (select hash_id from comic_files_duplicates) AND cfh.file_id = files.id ORDER BY hash;"
	SingleDuplicateSql  = "select path, name from files where id in (select file_id from comic_files where file_id not in (SELECT min(file_id) file_id from comic_files_hashes where hash_id in (select hash_id from comic_files_duplicates) group by hash_id order by hash_id))order by name;"
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
		createViews(db)
		instance = &DataStore{DB: db}
	})
	return instance
}

func (ds *DataStore) GetDuplicates() (files []models.FileData, err error) {
	rows, err := ds.DB.Query(DuplicatesSql)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var file models.FileData
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

func (ds *DataStore) GetFilesWithoutHash() (files []models.FileData, err error) {
	rows, err := ds.DB.Query(FilesWithoutHashSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file models.FileData
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

func (ds *DataStore) GetFilesFrom(source string) (files []models.FileData, err error) {
	rows, err := ds.DB.Query("SELECT id, path, size, hash FROM files WHERE path LIKE ?", source+"%")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var file models.FileData
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

func (ds *DataStore) InsertFile(file models.FileData) (models.FileData, error) {
	_, err := ds.DB.Exec(InsertFileSql, file.Name, file.Path, file.Size)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
	return file, err
}

func (ds *DataStore) InsertHash(files []models.FileData) error {
	tx, err := ds.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	for _, f := range files {
		_, err := tx.Exec(AddHashSql, f.Hash, f.Id)
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
