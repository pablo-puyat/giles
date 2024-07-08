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
	minVelocity   float64
	maxVelocity   float64
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
	go func() {
		for tr := range in {
			r := transformer(tr.File)
			out <- r
		}
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
	totalDuration += elapsed
	totalBytes += int(file.Size)
	processed += 1

	// Calculate the speed for this file and update min/max velocities
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
	return fmt.Sprintf("%d seconds", int(totalDuration.Seconds()))
}

func getVelocity() string {
	if processed == 0 || totalDuration.Seconds() == 0 {
		return ""
	}
	avgSpeed := float64(totalBytes) / totalDuration.Seconds() / (1024 * 1024) // Convert bytes per second to MB/s
	return fmt.Sprintf("Avg. Velocity: %.2f MB/s  Min. Velocity: %.2f MB/s  Max. Velocity: %.2f MB/s", avgSpeed, minVelocity, maxVelocity)
}

func statusString() string {
	return fmt.Sprintf("\rProgress: %d of %d files %s Duration: %s", processed, fileCount, getVelocity(), getTime())
}

type TransformResult struct {
	File     models.FileData
	Duration time.Duration
	Err      error
}
