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

func (fs *FileStore) GetFilesFrom(source string) (files []File, err error) {
	rows, err := fs.db.Query("SELECT id, coalesce(name, original_name) as name, coalesce(path, original_path) as path2, size, hash FROM files WHERE path2 LIKE ?", source+"%")
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
		println(file.Id)
		files = append(files, file)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return files, err
}

func (fs *FileStore) Batch(files []File) error {
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

		for _, file := range files {
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

func (fs *FileStore) InsertHash(files []File) error {
	return fs.execTx(func(tx *sql.Tx) error {
		for _, f := range files {
			if _, err := tx.Exec(
				`UPDATE files set hash = ? WHERE id = ?`,
				f.Hash,
				f.Id,
			); err != nil {
				return err
			}
		}
		return nil
	})
}
