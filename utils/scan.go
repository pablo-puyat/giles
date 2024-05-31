package utils

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type Scanner struct {
	db       *sql.DB
	dirPath  string
	hashFunc func(string) (string, error)
}

type ScannerBuilder struct {
	scanner Scanner
}

func (sb *ScannerBuilder) Build() *Scanner {
	return &sb.scanner
}

func main() {
	dirPath := flag.String("dir", ".", "Directory to scan for duplicates")
	flag.Parse()

	go func() {
		for {
			fmt.Printf("\rProcessed %d files...", atomic.LoadUint64(&processedFiles))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	duplicates := make(map[string][]string)
	err := filepath.Walk(*dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				fmt.Printf("Skipping %s (permission denied)\n", path)
				return filepath.SkipDir
			}
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		hash, err := utils.Hash(path)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", path, err)
			return nil // Continue with other files
		}
		duplicates[hash] = append(duplicates[hash], path)
		atomic.AddUint64(&processedFiles, 1)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	for _, paths := range duplicates {
		if len(paths) > 1 {
			fmt.Printf("Duplicates found:\n")
			for _, path := range paths {
				fmt.Printf("  %s\n", path)
			}
		}
	}
}
