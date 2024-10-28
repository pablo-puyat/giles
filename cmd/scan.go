package cmd

import (
	"fmt"
	"giles/internal/database"
	"giles/internal/scanner"
	"giles/internal/worker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"log"
	"time"
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
		if err := scanDir(args[0]); err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func scanDir(path string) error {
	store, err := database.New(databasePath)
	if err != nil {
		return fmt.Errorf("database access error: %w", err)
	}
	defer func(store *database.FileStore) {
		err := store.Close()
		if err != nil {
			log.Println("Error closing database")
		}
	}(store)

	s := scanner.New()
	done := make(chan struct{})

	batchDone := make(chan struct{})

	go s.DisplayProgress(done)

	go func() {
		worker.BatchProcessor(store, s.FilesChan)
		close(batchDone)
	}()

	if err := s.ScanFiles(path); err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	close(s.FilesChan)
	<-batchDone

	time.Sleep(100 * time.Millisecond)

	done <- struct{}{}
	close(done)

	return nil
}
