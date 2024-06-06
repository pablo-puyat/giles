package main

import (
	"flag"
	"fmt"
	"giles/database"
	"giles/models"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dirPath := flag.String("dir", ".", "Directory to scan for duplicates")
	extFilter := flag.String("ext", "", "File extension to filter (e.g., txt, jpg)")
	flag.Parse()

	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	progressCh := make(chan int, 5)
	databaseCh := make(chan []models.FileData)
	go printProcessedFiles(progressCh)
	go insertToDatabase(db, databaseCh)
	scanFiles(*dirPath, *extFilter, progressCh, databaseCh)
	close(progressCh)
	close(databaseCh)

}

func insertToDatabase(db *database.DB, databaseCh <-chan []models.FileData) {
	for files := range databaseCh {
		err := db.InsertFiles(files)
		if err != nil {
			log.Printf("Error inserting files: %v", err)
		}
	}
}

func printProcessedFiles(progressCh <-chan int) {
	for count := range progressCh {
		fmt.Printf("\rProcessed %d files", count)
	}
}

func scanFiles(dirPath, extFilter string, progressCh chan<- int, databaseCh chan<- []models.FileData) {
	var (
		processedFiles int
		fileBuffer     []models.FileData
	)

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() || (extFilter != "" && !strings.HasSuffix(strings.ToLower(info.Name()), "."+extFilter)) {
			return nil
		}

		fileData := models.FileData{
			Name: info.Name(),
			Path: path,
			Size: info.Size(),
		}

		fileBuffer = append(fileBuffer, fileData)
		if len(fileBuffer) >= 100 {
			databaseCh <- fileBuffer
			fileBuffer = nil // Reset the buffer after sending
		}

		processedFiles++
		progressCh <- processedFiles
		return nil
	})

	// Send any remaining files in the buffer
	if len(fileBuffer) > 0 {
		databaseCh <- fileBuffer
	}
}
