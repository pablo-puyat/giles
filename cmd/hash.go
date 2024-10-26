package cmd

import (
	"crypto/sha256"
	"fmt"
	"giles/internal/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type TransformResult struct {
	File     database.File
	Duration time.Duration
	Err      error
}

var (
	fileCount  int
	hashCmd    *cobra.Command
	processed  int
	startTime  time.Time
	totalBytes int
	workers    int
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
	logFile, err := setupLogging()
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	workers, _ = cmd.Flags().GetInt("workers")

	ds, err := database.NewDataStore()
	if err != nil {
		log.Printf("Error with query: %v\n", err)
		return
	}

	files, err := ds.GetFilesWithoutHash()
	if err != nil {
		log.Printf("Error with query: %v\n", err)
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
			log.Printf("Error processing file %s: %v\n", r.File.Name, r.Err)
		}
	}
	fmt.Println("\nDone.")
}

func addHash(in <-chan TransformResult, transformer func(database.File) TransformResult) <-chan TransformResult {
	out := make(chan TransformResult, workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for tr := range in {
				if tr.Err != nil {
					out <- tr
					continue
				}
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

func calculate(file database.File) TransformResult {
	st := time.Now()
	hash, err := calculateHash(file.Path)
	if err != nil {
		log.Printf("Error calculating hash: %v\n", err)
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
		log.Printf("Error opening file: %v\n", err)
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("Error hashing file: %v\n", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func generator(files []database.File) <-chan TransformResult {
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

func insertFiles(ds *database.FileStore, in <-chan TransformResult) <-chan TransformResult {
	out := make(chan TransformResult)
	go func() {
		var filesToProcess = make([]database.File, 0, workers)
		for tr := range in {
			if tr.Err != nil {
				out <- tr
				continue
			}
			filesToProcess = append(filesToProcess, tr.File)
			if len(filesToProcess) >= workers/2 {
				if err := ds.InsertHash(filesToProcess); err != nil {
					tr.Err = err
				}
			}
			out <- tr
		}
		if len(filesToProcess) > 0 {
			if err := ds.InsertHash(filesToProcess); err != nil {
				log.Printf("Error inserting hash: %v\n", err)
			}
		}
		close(out)
	}()
	return out
}

func setupLogging() (*os.File, error) {
	logFile, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
	return logFile, err
}

func statusString() string {
	return fmt.Sprintf("\rProgress: %d of %d files %s Duration: %s", processed, fileCount, getVelocity(), getTime())
}
