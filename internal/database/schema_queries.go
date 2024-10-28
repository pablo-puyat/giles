package database

import (
	"database/sql"
	"time"
)

type File struct {
	Id           int64     `db:"id"`
	Hash         string    `db:"hash"`
	Name         string    `db:"name"`
	OriginalName string    `db:"original_name"`
	OriginalPath string    `db:"original_path"`
	Path         string    `db:"path"`
	Size         int64     `db:"size"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

const Schema = `
CREATE TABLE IF NOT EXISTS files (
    id          	INTEGER PRIMARY KEY,
    hash        	TEXT NOT NULL,
    name        	TEXT DEFAULT NULL,
    original_name   TEXT NOT NULL,
    original_path   TEXT NOT NULL,
    path        	TEXT DEFAULT NULL,
    size       		INTEGER NOT NULL,
    created_at		DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at		DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_files_hash ON files(hash);
CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
CREATE INDEX IF NOT EXISTS idx_files_original_path ON files(original_path);
CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
`

func createTables(db *sql.DB) error {
	_, err := db.Exec(Schema)
	return err
}
