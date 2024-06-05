package main

import (
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dirPath := flag.String("dir", ".", "Directory to scan for duplicates")
	extFilter := flag.String("ext", "", "File extension to filter (e.g., txt, jpg)")
	flag.Parse()

	progressCh := make(chan int, 100)
	go printProcessedFiles(progressCh)
	scanFiles(*dirPath, *extFilter, progressCh)
	close(progressCh)

}

func printProcessedFiles(progressCh <-chan int) {
	for count := range progressCh {
		//print("\r%d", count + " ")
		fmt.Printf("\rProcessed %d files", count)
	}
}

func scanFiles(dirPath, extFilter string, progressCh chan<- int) {
	var processedFiles int
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() || (extFilter != "" && !strings.HasSuffix(strings.ToLower(info.Name()), "."+extFilter)) {
			return nil
		}

		//start := time.Now() // Get the current time before the hash calculation
		//_, err = utils.Hash(path)
		//duration := time.Since(start)
		//if err != nil {
		//	log.Printf("Error hashing %s: %v", path, err) // Log non-fatal errors
		//	return nil
		//}
		//if duration > time.Second { // Check if duration is over 1 second
		//fileSize := formatBytes(info.Size())
		//processedFiles = fmt.Sprintf("%s (size: %s, hash time: %v)", path, fileSize, duration)
		//processedFiles = fmt.Sprintf("%s (size: %s)", path, fileSize)
		processedFiles++
		progressCh <- processedFiles
		//}
		return nil
	})
}

func formatBytes(bytes int64) string {
	const (
		_        = iota
		KB int64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%dTB", int64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%dGB", int64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%dMB", int64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%dKB", int64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
