package database

import (
	"database/sql"
	"giles/models"
	"log"
)

const (
	FilesWithoutHashSql   = "SELECT id, files.name, files.path, files.size FROM files LEFT JOIN files_hashes ON files.id = files_hashes.file_id WHERE files_hashes.file_id IS NULL"
	InserFileSql          = "INSERT INTO files (name, path, size) VALUES (?, ?, ?)"
	InsertFileIdHashIdSql = "INSERT INTO files_hashes (file_id, hash_id) VALUES (?, ?)"
	InsertHashSql         = "INSERT INTO hashes (hash) VALUES (1)"
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

	hashId, err := result.LastInsertId()
	if err != nil {
		return models.FileData{}
	}
	file.HashId = hashId
	return file
}
