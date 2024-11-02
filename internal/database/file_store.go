package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultDBName = "giles.db"
	defaultPerms  = 0755
)

type FileStore struct {
	db        *sql.DB
	BatchSize int
	Path      string
}

func New(dbPath string) (*FileStore, error) {
	path, err := resolveDBPath(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close() // Clean up before returning error
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Printf("using database: %s\n", path)

	return &FileStore{
		db:        db,
		BatchSize: 100,
		Path:      path,
	}, nil
}

// resolveDBPath determines the appropriate database file path.
// If dbPath is provided and valid, it will be used.
// Otherwise, it falls back to standard locations in this order:
// 1. ~/.local/share/giles/giles.db
// 2. ./giles.db in current working directory
func resolveDBPath(dbPath string) (string, error) {
	// If path is provided, try to use it
	if dbPath != "" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, defaultPerms); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		return dbPath, nil
	}

	// Try to use standard XDG data directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		dir := filepath.Join(homeDir, ".local", "share", "giles")
		if err := os.MkdirAll(dir, defaultPerms); err == nil {
			return filepath.Join(dir, defaultDBName), nil
		}
	}

	// Fall back to current working directory
	pwd, err := os.Getwd()
	if err != nil {
		// Last resort: use relative path
		return filepath.Join(".", defaultDBName), nil
	}

	return filepath.Join(pwd, defaultDBName), nil
}

func (fs *FileStore) Close() error {
	if fs.db != nil {
		return fs.db.Close()
	}
	return nil
}
