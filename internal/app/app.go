package app

import (
	"giles/internal/database"
	"os"

	"github.com/charmbracelet/log"
)

type App struct {
	Logger   *log.Logger
	Database *database.FileStore
}

func New(dbPath string) (*App, error) {
	db, err := database.New(dbPath)
	if err != nil {
		return nil, err
	}
	return &App{
		Logger:   log.New(os.Stderr),
		Database: db,
	}, nil
}
