package cmd

import (
	"crypto/sha256"
	"fmt"
	"giles/database"
	"giles/models"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

var (
	hashedFiles       atomic.Uint64
	progressChan      = make(chan uint64)
	completedFiles    = make(chan models.FileData, 100)
	hashesToCalculate = 0
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Hash files in the database",
	Long: `Create hash for files in database that do not have one.

Usage: giles hash`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.GetInstance()
		if err != nil {
			log.Fatalf("Database error: %v", err)
		}

		files, err := db.GetFilesWithoutHash()

		if err != nil {
			log.Fatalf("Database error: %v", err)
		}

		hashesToCalculate = len(files)
		if hashesToCalculate == 0 {
			fmt.Printf("No files to hash\n")
			return
		}

		fileChannel := make(chan models.FileData)
		var wg sync.WaitGroup

		workers, err := cmd.Flags().GetInt("workers")

		go func() {
			for count := range progressChan {
				fmt.Printf("\rHashed %d of %d files", count, hashesToCalculate)
			}
		}()

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go hash(fileChannel, &wg)
		}

		go updateHashes(completedFiles)

		for _, file := range files {
			fileChannel <- file
		}

		close(fileChannel)
		wg.Wait()
		close(progressChan)
		close(completedFiles)
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 1, "Number of workers to use")
}

func hash(tasksChannel <-chan models.FileData, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range tasksChannel {
		f, err := os.Open(file.Path)
		if err != nil {
			log.Printf("Error encoutered while opening file: \"%v\"", err)
			continue
		}
		hashValue, err := calculateSHA256(f)
		if err != nil {
			log.Printf("Error encoutered while calculating hash: \"%v\"", err)
			continue
		}

		err = f.Close()
		if err != nil {
			log.Printf("Error encoutered: \"%v\"", err)
			return
		}

		file.Hash = hashValue
		completedFiles <- file

		count := hashedFiles.Add(1)
		progressChan <- count
	}
}

func updateHashes(hashedFilesChannel <-chan models.FileData) {
	db, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Database error: \"%v\"", err)
	}

	batch := make([]models.FileData, 0, 100)
	for file := range hashedFilesChannel {
		batch = append(batch, file)

		if len(batch) == 100 {
			if err := db.UpdateFileHashBatch(batch); err != nil {
				log.Printf("Error updating hashes \"%v\"", err)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := db.UpdateFileHashBatch(batch); err != nil {
			log.Printf("Error updating hashes \"%v\"", err)
		}
	}
}

func calculateSHA256(reader io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
