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

var processedFiles uint64

func main() {
	dirPath := flag.String("dir", ".", "Directory to scan for duplicates")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./giles.sqlite3")
	checkError(err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (id INT PRIMARY KEY, hash TEXT , name TEXT, path TEXT, size INT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	checkError(err)

	go printProcessedFiles()

	scanFiles(db, *dirPath, "")
}

func printProcessedFiles() {
	for {
		fmt.Printf("\rProcessed %d files...", atomic.LoadUint64(&processedFiles))
		time.Sleep(500 * time.Millisecond)
	}
}

func scanFiles(db *sql.DB, dirPath, extFilter string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() || (extFilter != "" && !strings.HasSuffix(strings.ToLower(path), "."+extFilter)) {
			return nil
		}

		hash, err := utils.Hash(path)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", path, err)
			return nil
		}

		_, err = db.Exec("INSERT OR IGNORE INTO files(hash, name, path, size) values(?, ?, ?, ?)", hash, info.Name(), path, info.Size())
		if err != nil {
			fmt.Printf("Error inserting into database: %v\n", err)
			return nil
		}

		atomic.AddUint64(&processedFiles, 1)
		return nil
	})

	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
