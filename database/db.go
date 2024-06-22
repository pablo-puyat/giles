package database

import (
	"database/sql"
	"giles/models"
	"log"
)

const (
	FilesWithoutHashSql   = "SELECT id, files.name, files.path, files.size FROM files LEFT JOIN files_hashes ON files.id = files_hashes.file_id WHERE files_hashes.file_id IS NULL"
	InsertHashSql         = "INSERT INTO hashes (hash) VALUES (1)"
	InsertFileIdHashIdSql = "INSERT INTO files (name, path, size) VALUES (?, ?, ?)"
)

func GetFilesWithoutHash(db *sql.DB) (files []models.FileData, err error) {
	rows, err := db.Query(FilesWithoutHashSql)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file models.FileData
		err := rows.Scan(&file.Id, &file.Name)
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

func InsertHash(db *sql.DB, hash string) (hashId int64, err error) {
	result, err := db.Exec(InsertHashSql, hash)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func InsertFileIdHashId(db *sql.DB, fileId int64, hashId int64) {
	_, err := db.Exec(InsertFileIdHashIdSql, fileId, hashId)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
}
