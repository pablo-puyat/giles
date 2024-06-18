package cmd

import (
	"fmt"
	"giles/database"
	"giles/models"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan files in a directory",
	Long: `Recursively scan files in a directory and insert them into a database.
The name, path and size are recorded.

Usage: giles scan <directory>`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := "."
		if len(args) > 0 {
			dirPath = args[0]
		}
		scanDirectory(dirPath)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

const batchSize = 100

func scanDirectory(dirPath string) {
	var wg sync.WaitGroup

	wg.Add(1)
	fileDataCh := make(chan []models.FileData, 10)

	go func() {
		defer wg.Done()
		insertToDatabase(fileDataCh)
	}()

	scanFiles(dirPath, fileDataCh)

	close(fileDataCh)
	wg.Wait()

	fmt.Println("\nScanning complete.")
}

func insertToDatabase(fileDataCh <-chan []models.FileData) {
	db, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	for files := range fileDataCh {
		if err := db.InsertFiles(files); err != nil {
			log.Printf("Error inserting files: \"%v\"", err)
		}
	}
}

func scanFiles(dirPath string, fileDataCh chan<- []models.FileData) {
	fileBuffer := make([]models.FileData, 0, batchSize)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking path: \"%v\"", err)
		}

		fileData := models.FileData{
			Name: info.Name(),
			Path: path,
			Size: info.Size(),
		}

		fileBuffer = append(fileBuffer, fileData)
		if len(fileBuffer) == batchSize {
			fileDataCh <- fileBuffer
			fileBuffer = fileBuffer[:0]
		}
		return nil
	})

	if err != nil {
		log.Printf("Error during file walk: \"%v\"", err)
	}

	if len(fileBuffer) > 0 {
		fileDataCh <- fileBuffer
	}
}
