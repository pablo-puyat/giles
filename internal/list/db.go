package list

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type FileEntry struct {
	Filename string
	Path     string
	Type     string
}

type FileStore struct {
	db *sql.DB
}

func NewFileStore(dbPath string) (*FileStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &FileStore{db: db}, nil
}

func (s *FileStore) Close() error {
	return s.db.Close()
}

func (s *FileStore) DeleteFile(path string) error {
	_, err := s.db.Exec("DELETE FROM files WHERE filepath = ?", path)
	return err
}

