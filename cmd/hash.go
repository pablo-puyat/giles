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
	out := make(chan TransformResult)
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
		for f := range in {
			file, err := ds.InsertFile(f.File)
			if err != nil {
				log.Printf("Error inserting file: %v", err)
				continue
			}
			out <- TransformResult{file, f.Duration, nil}
		}
		close(out)
	}()
	return out
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

func calculate(file models.FileData) TransformResult {
	st := time.Now()
	hash, err := calcHash(file.Path)
	if err != nil {
		return TransformResult{file, 0, err}
	}
	elapsed := time.Since(st)
	file.Hash = hash
	totalBytes += int(file.Size)
	processed += 1

	speed := float64(file.Size) / elapsed.Seconds() / (1024 * 1024) // Speed in MB/s for this file
	updateVelocity(speed)

	return TransformResult{file, elapsed, nil}
}

func updateVelocity(speed float64) {
	if speed < minVelocity || minVelocity == 0 {
		minVelocity = speed
	}
	if speed > maxVelocity {
		maxVelocity = speed
	}
}

func getTime() string {
	return fmt.Sprintf("%d seconds", int(time.Now().Sub(startTime).Seconds()))
}
func getVelocity() string {
	if processed == 0 {
		return ""
	}
	// Calculate avgSpeed in MB/s
	avgSpeedMB := float64(totalBytes) / time.Now().Sub(startTime).Seconds() / (1024 * 1024) // Convert bytes per second to MB/s

	if avgSpeedMB <= 0.50 {
		// Calculate avgSpeed in KB/s
		avgSpeedKB := float64(totalBytes) / time.Now().Sub(startTime).Seconds() / 1024 // Convert bytes per second to KB/s
		return fmt.Sprintf("Avg. Velocity: %.0f KB/s  Min. Velocity: %.0f KB/s  Max. Velocity: %.0f KB/s", avgSpeedKB, minVelocity/1024, maxVelocity/1024)
	}

	return fmt.Sprintf("Avg. Velocity: %.0f MB/s  Min. Velocity: %.0f MB/s  Max. Velocity: %.0f MB/s", avgSpeedMB, minVelocity, maxVelocity)
}

func statusString() string {
	return fmt.Sprintf("\rProgress: %d of %d files %s Duration: %s", processed, fileCount, getVelocity(), getTime())
}

type TransformResult struct {
	File     models.FileData
	Duration time.Duration
	Err      error
}
