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
		ds := database.NewDataStore()
		files, err := ds.GetFilesWithoutHash()
		if err != nil {
			log.Printf("Error with query: %v", err)
			return
		}

		c1 := generator(files)
		c2 := transform(c1, func(file models.FileData) (models.FileData, error) {
			file, err := insertHash(ds, file)
			if err != nil {
				log.Fatalf("Error inserting hash: %v", err)
			}
			return file, err
		})
		c3 := transform(c2, ds.InsertFileIdHashId)

		for r := range c3 {
			if r.Err != nil {
				fmt.Printf("final --- %v\n", r.Err)
			}
		}

		print("\rDone. \n\nProcessed ", len(files), " files\n")
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 1, "Number of workers to use")
}

func generator(files []models.FileData) <-chan TransformResult {
	out := make(chan TransformResult)
	go func() {
		for _, f := range files {
			out <- TransformResult{File: f}
		}
		close(out)
	}()
	return out
}

func insertHash(ds *database.DataStore, file models.FileData) (models.FileData, error) {
	var hash string
	hash, err := calcHash(file.Path)
	if err != nil {
		return file, err
	}
	file.Hash = hash
	hashId, err := ds.InsertHash(file)
	if err != nil {
		return file, err
	}
	file.HashId = hashId.HashId
	println("Inserting ", hash, " for id ", file.Id)
	return file, nil
}

func transform(in <-chan TransformResult, transformer func(models.FileData) (models.FileData, error)) <-chan TransformResult {
	out := make(chan TransformResult)
	go func() {
		for tr := range in {
			println("Transforming ", tr.File.Id)
			file, err := transformer(tr.File)
			out <- TransformResult{File: file, Err: err}
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
