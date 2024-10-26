package cmd

import (
	"fmt"
	"giles/internal/database"
	"giles/internal/scanner"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"log"
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
		scanDir(dirPath)
	},
}

type Progress struct {
	totalFiles   int64
	scannedFiles int64
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func scanDir(path string) {
	store, err := database.NewDataStore()
	if err != nil {
		log.Fatalf("Error accessing database")
	}

	s := scanner.New()

	fmt.Println("Counting files...")
	if err := s.CountFiles(path); err != nil {
		log.Fatalf("Error while counting files")
	}

	done := make(chan bool)
	go s.DisplayProgress(done)

	s.WaitGroup.Add(1)
	go worker.BatchProcessor(store, s.FilesChange, &s.WaitGroup)

	if err := s.ScanFiles(*rootDir); err != nil {
		log.Println("An error occured while scanning files.")
	}
	close(s.FilesChan)
	s.WaitGroup.Wait()
}
