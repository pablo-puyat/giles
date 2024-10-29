package worker

import (
	"fmt"
	"giles/internal/database"
)

func BatchInsert(store *database.FileStore, filesChan <-chan database.File) <-chan database.File {
	out := make(chan database.File, 1)
	batch := make([]database.File, 0, store.BatchSize)
	go func() {
		defer close(out)
		for file := range filesChan {
			batch = append(batch, file)
			out <- file

			if len(batch) >= store.BatchSize {
				if err := store.Insert(batch); err != nil {
					fmt.Printf("Error writing batch to database: %v\n", err)
				}
				batch = batch[:0]
			}
		}

		// Process remaining files
		if len(batch) > 0 {
			if err := store.Insert(batch); err != nil {
				fmt.Printf("Error writing final batch to database: %v\n", err)
			}
		}
	}()
	return out
}
