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
		dirPath := "." // Default to current directory
		if len(args) > 0 {
			dirPath = args[0]
		}
		blah(dirPath)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

const batchSize = 100

func blah(dirPath string) {

	var wg sync.WaitGroup // for synchronization
	wg.Add(1)
	fileDataCh := make(chan []models.FileData, 10) // Buffer channel for better performance

	go func() {
		defer wg.Done()
		insertToDatabase(fileDataCh)
	}()

	scanFiles(dirPath, fileDataCh)

	close(fileDataCh) // Signal that scanning is done
	wg.Wait()         // Wait for inserts to complete

	fmt.Println("\nScanning complete.") // Print a clean completion message
}

func insertToDatabase(fileDataCh <-chan []models.FileData) {
	db, err := database.GetInstance() // Get the singleton instance
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	for files := range fileDataCh {
		if err := db.InsertFiles(files); err != nil {
			log.Printf("Error inserting files: %v", err)
			// ... (error handling)
		}
	}
}

func scanFiles(dirPath string, fileDataCh chan<- []models.FileData) {
	fileBuffer := make([]models.FileData, 0, batchSize)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle errors immediately in the walk function
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		fileData := models.FileData{
			Name: info.Name(),
			Path: path,
			Size: info.Size(),
		}

		fileBuffer = append(fileBuffer, fileData)
		if len(fileBuffer) == batchSize {
			fileDataCh <- fileBuffer // Send a full batch
			fileBuffer = fileBuffer[:0]
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error during file walk: %v", err)
	}

	if len(fileBuffer) > 0 {
		fileDataCh <- fileBuffer // Send any remaining files
	}
}
