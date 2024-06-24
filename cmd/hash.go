package cmd

import (
	"crypto/sha256"
	"fmt"
	"giles/database"
	"giles/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Hash files in the database",
	Long: `Create hash for files in database that do not have one.

Usage: giles hash`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		dbManager := database.NewDataStore()
		files, err := dbManager.GetFilesWithoutHash()
		if err != nil {
			log.Printf("Error with query: %v", err)
			return
		}

		c1 := make(chan TransformResult)
		c2 := transform(c1, func(file models.FileData) (models.FileData, error) {
			return insertHash(dbManager, file)
		})
		c3 := transform(c2, dbManager.InsertFileIdHashId)

		go func() {
			for r := range c3 {
				if r.Err != nil {
					fmt.Printf("final --- %v\n", r.Err)
				}
			}
		}()

		for _, f := range files {
			c1 <- TransformResult{File: f}
		}
		close(c1)

		print("\rDone. \n\nProcessed ", len(files), " files\n")
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 1, "Number of workers to use")
}

func insertHash(dbManager *database.DataStore, file models.FileData) (models.FileData, error) {
	var hash string
	hash, err := calcHash(file.Path)
	if err != nil {
		return file, err
	}
	file.Hash = hash
	hashId, err := dbManager.InsertHash(file)
	if err != nil {
		return file, err
	}
	file.HashId = hashId.HashId
	return file, nil
}

func transform(in <-chan TransformResult, transformer func(models.FileData) (models.FileData, error)) <-chan TransformResult {
	out := make(chan TransformResult)
	go func() {
		for tr := range in {
			if tr.Err != nil {
				out <- tr
			} else {
				file, err := transformer(tr.File)
				out <- TransformResult{File: file, Err: err}
			}
		}
		close(out)
	}()
	return out
}

func calcHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error encoutered while opening file: \"%v\"", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalf("Error encoutered while hashing file: \"%v\"", err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	return hash, err
}

type TransformResult struct {
	File models.FileData
	Err  error
}
