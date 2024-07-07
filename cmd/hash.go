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
	"time"
)

var (
	fileCount     int
	hashCmd       *cobra.Command
	processed     int
	totalDuration time.Duration
	totalBytes    int
	workers       int
)

func init() {
	hashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Hash files in the database",
		Long:  `Create hash for files in database that do not have one.`,
		Args:  cobra.NoArgs,
		Run:   hashFiles,
	}
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 1, "Number of workers to use")
}

func hashFiles(cmd *cobra.Command, args []string) {
	workers, _ = cmd.Flags().GetInt("workers")

	ds := database.NewDataStore()
	files, err := ds.GetFilesWithoutHash()
	if err != nil {
		log.Printf("Error with query: %v", err)
		return
	}
	fileCount = len(files)
	fmt.Printf("Calculating hash for %d files\n", fileCount)

	c1 := generator(files)
	c2 := transformBuffered(c1, calculate)
	c3 := insertFiles(ds, c2)

	for r := range c3 {
		print("\r Processed ", processed, " of ", fileCount, " files", " Total bytes: ", totalBytes, " Total duration: ", totalDuration, " Average speed: ", getAverageSpeed())
		if r.Err != nil {
			fmt.Printf("final error--- %v\n", r.Err)
		}
	}
	print("\rDone. \n\nProcessed ", len(files), " files\n")
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

func transformBuffered(in <-chan TransformResult, transformer func(models.FileData) (models.FileData, error)) <-chan TransformResult {
	out := make(chan TransformResult, workers)
	go func() {
		for tr := range in {
			file, err := transformer(tr.File)
			out <- TransformResult{File: file, Err: err}
		}
		close(out)
	}()
	return out
}

func insertFiles(ds *database.DataStore, in <-chan TransformResult) <-chan TransformResult {
	out := make(chan TransformResult)
	go func() {
		var filesToInsert []models.FileData
		for file := range in {
			filesToInsert = append(filesToInsert, file.File)
			if len(filesToInsert) == workers {
				batchInsertFiles(ds, filesToInsert, out)
				filesToInsert = nil
			}
		}
		if len(filesToInsert) > 0 {
			batchInsertFiles(ds, filesToInsert, out)
		}
		close(out)
	}()
	return out
}

func batchInsertFiles(ds *database.DataStore, files []models.FileData, out chan<- TransformResult) {
	for _, f := range files {
		file, err := ds.InsertFile(f)
		if err != nil {
			log.Printf("Error inserting file: %v", err)
			continue
		}
		file, err = ds.InsertFileIdHashId(file)
		processed++
		out <- TransformResult{File: file, Err: err}
	}
}

func calculate(file models.FileData) (models.FileData, error) {
	st := time.Now()
	hash, err := calcHash(file.Path)
	if err != nil {
		return file, err
	}
	file.Hash = hash
	elapsed := time.Since(st)
	totalDuration += elapsed
	print("\r Processed ", processed, " of ", fileCount, " files", " Total bytes: ", totalBytes, " Total duration: ", totalDuration, " Average speed: ", getAverageSpeed())
	return file, nil
}

func calcHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening file: \"%v\"", err)
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalf("Error hashing file: \"%v\"", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getAverageSpeed() time.Duration {
	if processed == 0 {
		return 0
	}
	return totalDuration / time.Duration(processed)
}

type TransformResult struct {
	File models.FileData
	Err  error
}
