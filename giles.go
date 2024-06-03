package main

import (
	"database/sql"
	"flag"
	"fmt"
	"giles/utils"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var processedFiles uint64 = 0

func main() {
	dirPath := flag.String("dir", ".", "Directory to scan for duplicates")
	flag.Parse()

	db, err1 := sql.Open("sqlite3", "./giles.sqlite3")
	if err1 != nil {
		fmt.Printf("Error opening database: %v\n", err1)
		os.Exit(1)
	}
	defer db.Close()

	_, err2 := db.Exec("CREATE TABLE IF NOT EXISTS files (id INT PRIMARY KEY, hash TEXT , name TEXT, path TEXT, size INT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err2 != nil {
		fmt.Printf("Error creating table: %v\n", err2)
		os.Exit(1)
	}

	go func() {
		for {
			fmt.Printf("\rProcessed %d files...", atomic.LoadUint64(&processedFiles))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	scanFiles(db, *dirPath, "")

}

func scanFiles(db *sql.DB, dirPath, extFilter string) {

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// ... (permission error handling remains the same)
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if extFilter != "" && !strings.HasSuffix(strings.ToLower(path), "."+extFilter) {
			return nil
		}
		hash, err := utils.Hash(path)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", path, err)
			return nil // Continue with other files
		}
		name, size := info.Name(), info.Size()

		_, err = db.Exec("INSERT OR IGNORE INTO files(hash, name, path, size) values(?, ?, ?, ?)", hash, name, path, size)
		if err != nil {
			fmt.Printf("Error inserting into database: %v\n", err)
			return nil
		}

		atomic.AddUint64(&processedFiles, 1)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}
