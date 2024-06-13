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
		if len(files) == 0 {
			fmt.Printf("No files to hash\n")
			return
		}
		fmt.Printf("Hashing %d files\n", len(files))
		tasksChannel := make(chan models.FileData)
		var wg sync.WaitGroup

		workers, err := cmd.Flags().GetInt("workers")

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go hash(tasksChannel, &wg)
		}

		for _, file := range files {
			tasksChannel <- file
		}
		close(tasksChannel)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 7, "Number of workers to use")
}

func hash(tasksChannel <-chan models.FileData, wg *sync.WaitGroup) {
	defer wg.Done()
	db, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Database error: \"%v\"", err)
	}
	for file := range tasksChannel {
		f, err := os.Open(file.Path)
		if err != nil {
			log.Printf("Error encoutered while opening file: \"%v\"", err)
			continue
		}
		defer f.Close()

		hashValue, err := calculateSHA256(f)
		if err != nil {
			log.Printf("Error encoutered while calculating has: \"%v\"", err)
			continue
		}

		if err := db.UpdateFileHash(file.Path, hashValue); err != nil {
			log.Printf("Error inserting hash \"%v\"", err)
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
