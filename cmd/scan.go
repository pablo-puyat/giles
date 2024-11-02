package cmd

import (
	"fmt"
	"giles/internal/database"
	"giles/internal/scanner"
	"giles/internal/worker"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
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
	start := time.Now()
	store, err := database.New(dbPath)
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
	files := s.Run(path)
	hashes := s.Hash(files)
	completed := worker.BatchInsert(store, hashes)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.DisplayProgress(completed)
	}()
	wg.Wait()
	fmt.Printf("scan completed in %v\n", time.Since(start))
	return nil
}
