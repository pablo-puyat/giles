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
	"sync"
	"time"
)

var (
	fileCount   int
	hashCmd     *cobra.Command
	processed   int
	startTime   time.Time
	totalBytes  int
	workers     int
	minVelocity float64
	maxVelocity float64
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

func addHash(in <-chan TransformResult, transformer func(models.FileData) TransformResult) <-chan TransformResult {
	out := make(chan TransformResult, workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for tr := range in {
				r := transformer(tr.File)
				out <- r
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func calculate(file models.FileData) TransformResult {
	st := time.Now()
	hash, err := calculateHash(file.Path)
	if err != nil {
		return TransformResult{file, 0, err}
	}
	elapsed := time.Since(st)
	file.Hash = hash
	totalBytes += int(file.Size)
	processed += 1

	return TransformResult{file, elapsed, nil}
}

func calculateHash(path string) (string, error) {
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

func getTime() string {
	return fmt.Sprintf("%d seconds", int(time.Now().Sub(startTime).Seconds()))
}

func getVelocity() string {
	if processed == 0 {
		return ""
	}
	v := float64(totalBytes) / time.Now().Sub(startTime).Seconds() / (1024 * 1024)
	return fmt.Sprintf("Velocity: %.0f MB/s", v)
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
	fmt.Printf("Processing %d files with %d workers\n", fileCount, workers)

	c1 := generator(files)
	c2 := addHash(c1, calculate)
	c3 := insertFiles(ds, c2)
	startTime = time.Now()
	for r := range c3 {
		print(statusString())
		if r.Err != nil {
			fmt.Printf("final error--- %v\n", r.Err)
		}
	}
}

func insertFiles(ds *database.DataStore, in <-chan TransformResult) <-chan TransformResult {
	out := make(chan TransformResult)
	go func() {
		var filesToProcess = make([]models.FileData, 0, workers)
		for tr := range in {
			filesToProcess = append(filesToProcess, tr.File)
			if len(filesToProcess) >= workers/2 {
				processFiles(ds, &filesToProcess)
			}
			out <- tr
		}
		if len(filesToProcess) > 0 {
			processFiles(ds, &filesToProcess)
		}
		close(out)
	}()
	return out
}

func processFiles(ds *database.DataStore, filesToProcess *[]models.FileData) {
	for _, f := range *filesToProcess {
		if g, err := ds.InsertFile(f); err == nil {
			if _, err = ds.InsertFileIdHashId(g); err == nil {
				print(statusString())
			}
		}
	}
	*filesToProcess = nil
}

func statusString() string {
	return fmt.Sprintf("\rProgress: %d of %d files %s Duration: %s", processed, fileCount, getVelocity(), getTime())
}

type TransformResult struct {
	File     models.FileData
	Duration time.Duration
	Err      error
}
