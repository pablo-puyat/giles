package database

import (
	"database/sql"
	"log"
)

func (fs *FileStore) GetDuplicates() (files []File, err error) {
	rows, err := fs.db.Query(`
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
		var file File
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

func (fs *FileStore) GetFilesWithoutHash() (files []File, err error) {
	rows, err := fs.db.Query(`SELECT id, files.path, size FROM files WHERE hash IS NULL;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file File
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

func (fs *FileStore) InsertFile(file File) (File, error) {
	_, err := fs.db.Exec(` INSERT OR IGNORE INTO files (name, path, size) VALUES (?, ?, ?);`, file.Name, file.Path, file.Size)
	if err != nil {
		log.Printf("Error inserting file: %v", err)
	}
	return file, err
}

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

func (fs *FileStore) StoreBatch(files []File) error {
	tx, err := fs.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT INTO files (path, name, size)
        VALUES (?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, file := range files {
		println(file.Path)
		_, err = stmt.Exec(
			file.Path,
			file.Name,
			file.Size,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// StoreHash stores a file's hash in the database
func (fs *FileStore) StoreHash(filePath, hashType, hashValue string) error {
	_, err := fs.db.Exec(`
        INSERT INTO file_hashes (file_id, hash_type, hash_value)
        SELECT id, ?, ?
        FROM files
        WHERE path = ?
    `, hashType, hashValue, filePath)
	return err
}
