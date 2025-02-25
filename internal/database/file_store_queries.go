package database

import (
	"database/sql"
)

func (fs *FileStore) execTx(fn func(*sql.Tx) error) error {
	tx, err := fs.db.Begin()

	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (fs *FileStore) GetFiles() (files []File, err error){
	rows, err := fs.db.Query("SELECT id, coalesce(name, original_name) as name, coalesce(path, original_path) as path2, size, hash FROM files")

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
		err := rows.Scan(&file.Id, &file.Name, &file.Path, &file.Size, &file.Hash)
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
	rows, err := fs.db.Query("SELECT id, original_name as name, original_path as path, size, hash FROM files WHERE path LIKE ?", source+"%")

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
		err := rows.Scan(&file.Id, &file.Name, &file.Path, &file.Size, &file.Hash)
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

func (fs *FileStore) Insert(files []File) error {
	return fs.execTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
            INSERT INTO files (original_path, original_name, size, hash)
            VALUES (?, ?, ?, ?)
        `)

		if err != nil {
			return err
		}

		defer stmt.Close()

		for _, file := range files {
			if _, err = stmt.Exec(
				file.Path,
				file.Name,
				file.Size,
				file.Hash,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (fs *FileStore) Update(files []File) error {
	return fs.execTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
            UPDATE files 
            SET path = ?, name = ?
            WHERE id = ?
        `)

		if err != nil {
			return err
		}

		defer stmt.Close()

		for i := 0; i < len(files); i++ {
			file := (files)[i]
			if _, err = stmt.Exec(
				file.Path,
				file.Name,
				file.Id,
			); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *FileStore) DeleteFile(path string) error {
	_, err := s.db.Exec("DELETE FROM files WHERE filepath = ?", path)
	return err
}

func (s *FileStore) GetAllFiles() ([]File, error) {
	rows, err := s.db.Query("SELECT filename, filepath, type FROM files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []File
	for rows.Next() {
		var entry File
		if err := rows.Scan(&entry.Name, &entry.Path); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
