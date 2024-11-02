package scanner

import (
	"crypto/sha256"
	"fmt"
	"giles/internal/database"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Scanner struct {
	Files chan database.File
}

func New() *Scanner {
	return &Scanner{}
}

// walks the directory tree and sends FileInfo to the channel
func (s *Scanner) Run(root string) <-chan database.File {
	out := make(chan database.File)
	go func() {
		defer close(out)
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if info.Name() == ".DS_Store" {
				return nil
			}

			fileInfo := database.File{
				Path: path,
				Name: d.Name(),
				Size: info.Size(),
			}

			out <- fileInfo

			return nil
		})
	}()
	return out
}

func (s *Scanner) Hash(in <-chan database.File) <-chan database.File {
	out := make(chan database.File)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range in {
				hash, err := hash(file.Path)
				if err != nil {
					log.Printf("Error hashing file: %v", err)
				}
				file.Hash = hash
				out <- file
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func hash(path string) (string, error) {
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
