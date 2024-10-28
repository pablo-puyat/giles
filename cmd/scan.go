package cmd

import (
	"giles/internal/database"
	"giles/internal/scanner"
	"giles/internal/worker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

var scanCmd = &cobra.Command{
	Use:   "scan <path>",
	Short: "Scan files in a directory",
	Long: `Recursively scan files in a directory and insert them into a database.
The name, path and size are recorded.

Usage: giles scan <directory>`,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		scanDir(args[0])
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
	store, err := database.New(databasePath)
	if err != nil {
		log.Fatalf("Error accessing database")
	}

	s := scanner.New()

	// Create buffered done channel
	done := make(chan bool, 100)

	// Start progress display
	progressWg := sync.WaitGroup{}
	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		s.DisplayProgress(done)
	}()

	go worker.BatchProcessor(store, s.FilesChan)

	// Scan files
	if err := s.ScanFiles(path); err != nil {
		log.Printf("An error occurred while scanning files: %v\n", err)
	}

	// Close channel after scanning is complete
	close(s.FilesChan)

	// Signal progress display to finish and wait for it
	done <- true
	progressWg.Wait()
}
