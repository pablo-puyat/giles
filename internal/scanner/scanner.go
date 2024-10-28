package scanner

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"giles/internal/database"
)

type Scanner struct {
	Progress  *Progress
	FilesChan chan database.File
}

func New() *Scanner {
	return &Scanner{
		Progress:  &Progress{},
		FilesChan: make(chan database.File, 1),
	}
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

		if info.IsDir() == true {
			return nil
		}

		if info.Name() == ".DS_Store" {
			return nil
		}

		fileHash, err := calcHash(path)
		if err != nil {
			log.Println("Error calculating hash")
			return err
		}

		fileInfo := database.File{
			Path: path,
			Name: d.Name(),
			Size: info.Size(),
			Hash: fileHash,
		}

		s.FilesChan <- fileInfo

		atomic.AddInt64(&s.Progress.ScannedFiles, 1)

		return nil
	})
}

func calcHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("Error hashing file: %v\n", err)
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
