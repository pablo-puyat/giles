package database

import "log"

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
