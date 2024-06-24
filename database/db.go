package database

import (
	"database/sql"
	"giles/models"
	"log"
)

const (
	FilesWithoutHashSql   = "SELECT id, files.path FROM files LEFT JOIN files_hashes ON files.id = files_hashes.file_id WHERE files_hashes.file_id IS NULL"
	InserFileSql          = "INSERT INTO files (name, path, size) VALUES (?, ?, ?)"
	InsertFileIdHashIdSql = "INSERT INTO files_hashes (file_id, hash_id) VALUES (?, ?)"
	InsertHashSql         = "INSERT OR IGNORE INTO hashes (hash) VALUES (?);"
)

func GetFilesWithoutHash(db *sql.DB) (files []models.FileData, err error) {
	rows, err := db.Query(FilesWithoutHashSql)
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

func InsertFile(db *sql.DB, file models.FileData) models.FileData {
	_, err := db.Exec(InserFileSql, file.Name, file.Path, file.Size)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
	return file
}

func InsertFileIdHashId(db *sql.DB, file models.FileData) models.FileData {
	_, err := db.Exec(InsertFileIdHashIdSql, file.Id, file.HashId)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
	return file
}

func InsertHash(db *sql.DB, file models.FileData) models.FileData {
	result, err := db.Exec(InsertHashSql, file.Hash)
	if err != nil {
		log.Fatalf("Error inserting hash: %v", err)
	}
	ra, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Error inserting hash: %v", err)
	}
	if ra == 0 {
		rows, err := db.Query("SELECT id FROM hashes WHERE hash = ?", file.Hash)
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
		file.HashId = ra
	}
	return file
}
