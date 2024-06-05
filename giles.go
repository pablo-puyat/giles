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
		fmt.Printf("\rProcessed %d files", count)
	}
}

func scanFiles(dirPath, extFilter string, progressCh chan<- int) {
	var processedFiles int
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() || (extFilter != "" && !strings.HasSuffix(strings.ToLower(info.Name()), "."+extFilter)) {
			return nil
		}

		processedFiles++
		progressCh <- processedFiles
		return nil
	})
}
