package main

import (
	"database/sql"
	"flag"
	"fmt"
	"giles/utils"
	"log" // Use log package for errors
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var processedFiles uint64

func main() {
	dirPath := flag.String("dir", ".", "Directory to scan for duplicates")
	extFilter := flag.String("ext", "", "File extension to filter (e.g., txt, jpg)")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./giles.sqlite3")
	if err != nil {
		log.Fatalf("Error opening database: %v", err) // Log and exit on fatal errors
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY,
        hash TEXT,
        name TEXT,
        path TEXT,
        size INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	go printProcessedFiles()
	scanFiles(db, *dirPath, *extFilter)
}

func printProcessedFiles() {
	for {
		fmt.Printf("\rProcessed %d files...", atomic.LoadUint64(&processedFiles))
		time.Sleep(500 * time.Millisecond)
	}
}

func scanFiles(db *sql.DB, dirPath, extFilter string) {
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		// Combined condition for efficiency and readability
		if err != nil || !info.Mode().IsRegular() || (extFilter != "" && !strings.HasSuffix(strings.ToLower(info.Name()), "."+extFilter)) {
			return nil
		}

		hash, err := utils.Hash(path)
		if err != nil {
			log.Printf("Error hashing %s: %v", path, err) // Log non-fatal errors
			return nil
		}

		if _, err := db.Exec( // Use if statement for better error handling
			"INSERT OR IGNORE INTO files(hash, name, path, size) values(?, ?, ?, ?)",
			hash, info.Name(), path, info.Size(),
		); err != nil {
			log.Printf("Error inserting into database: %v", err)
			return nil
		}

		atomic.AddUint64(&processedFiles, 1)
		return nil
	})
}
