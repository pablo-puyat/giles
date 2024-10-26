package worker

import (
	"fmt"
	"sync"

	"giles/internal/database"
)

func BatchProcessor(store *database.FileStore, filesChan <-chan database.File, wg *sync.WaitGroup) {
	defer wg.Done()

	batch := make([]database.File, 0, store.BatchSize)

	for file := range filesChan {
		batch = append(batch, file)

		if len(batch) >= store.BatchSize {
			if err := store.StoreBatch(batch); err != nil {
				fmt.Printf("Error writing batch to database: %v\n", err)
			}
			batch = batch[:0]
		}
	}

	// Process remaining files
	if len(batch) > 0 {
		if err := store.StoreBatch(batch); err != nil {
			fmt.Printf("Error writing final batch to database: %v\n", err)
		}
	}
}
