package database

import (
	"database/sql"
	"log"
)

func (fs *FileStore) InsertHash(files []File) error {
	tx, err := fs.db.Begin()
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

func (fs *FileStore) Batch(files []File) error {
	tx, err := fs.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT INTO files (path, name, size, hash)
        VALUES (?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, file := range files {
		_, err = stmt.Exec(
			file.Path,
			file.Name,
			file.Size,
			file.Hash,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (fs *FileStore) GetFilesFrom(source string) (files []File, err error) {
	rows, err := fs.db.Query("SELECT id, path, size, hash FROM files WHERE path LIKE ?", source+"%")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var file File
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
