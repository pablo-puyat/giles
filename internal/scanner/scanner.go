package scanner

import (
	"io/fs"
	"path/filepath"
	"sync"
	"sync/atomic"

	"giles/internal/database"
)

type Scanner struct {
	Progress  *Progress
	FilesChan chan database.File
	WaitGroup sync.WaitGroup
}

func New() *Scanner {
	return &Scanner{
		Progress:  &Progress{},
		FilesChan: make(chan database.File, 1000),
	}
}

// CountFiles counts the total number of files in the directory tree
func (s *Scanner) CountFiles(root string) error {
	var count int64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}
	atomic.StoreInt64(&s.Progress.TotalFiles, count)
	return nil
}

// ScanFiles walks the directory tree and sends FileInfo to the channel
func (s *Scanner) ScanFiles(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		fileInfo := database.File{
			Path: path,
			Name: d.Name(),
			Size: info.Size(),
		}

		s.FilesChan <- fileInfo

		if !d.IsDir() {
			atomic.AddInt64(&s.Progress.ScannedFiles, 1)
		}

		return nil
	})
}
