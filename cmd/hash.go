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
		fileCount = len(files)
		fmt.Printf("Calculating hash for %d files\n", fileCount)

		c1 := generator(files)

		c2 := transformBuffered(c1, func(file models.FileData) (models.FileData, error) {
			file, err := calculate(file)
			if err != nil {
				log.Printf("Error inserting hash: %v", err)
			}
			return file, err
		})

		c3 := make(chan TransformResult)

		go func() {
			var filesToInsert []models.FileData
			for file := range c2 {
				filesToInsert = append(filesToInsert, file.File)
				if len(filesToInsert) == 8 {
					for _, f := range filesToInsert {
						file, err := ds.InsertFile(f)
						if err != nil {
							log.Printf("Error inserting file: %v", err)
							continue
						}
						file, err = ds.InsertFileIdHashId(file)
						if err != nil {
							log.Printf("Error inserting file and hash id: %v", err)
						}
						processed++
						c3 <- TransformResult{File: file, Err: err}

					}
					filesToInsert = nil
				}
			}
			if len(filesToInsert) > 0 {
				for _, f := range filesToInsert {
					file, err := ds.InsertFile(f)
					if err != nil {
						log.Printf("Error inserting file: %v", err)
						continue
					}
					file, err = ds.InsertFileIdHashId(file)
					if err != nil {
						log.Printf("Error inserting file and hash id: %v", err)
					}
					processed++
					c3 <- TransformResult{File: file, Err: err}
				}
				filesToInsert = nil
			}
			close(c3)
		}()

		for r := range c3 {
			print("\r Processed ", processed, " of ", fileCount, " files")
			if r.Err != nil {
				fmt.Printf("final error--- %v\n", r.Err)
			}
		}

		print("\rDone. \n\nProcessed ", len(files), " files\n")
	},
}
var processed int = 0
var fileCount int = 0

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

func calculate(file models.FileData) (models.FileData, error) {
	hash, err := calcHash(file.Path)
	if err != nil {
		return file, err
	}
	file.Hash = hash
	return file, nil
}

func transformBuffered(in <-chan TransformResult, transformer func(models.FileData) (models.FileData, error)) <-chan TransformResult {
	var wc int = 8
	out := make(chan TransformResult, wc)

	wg := sync.WaitGroup{}
	wg.Add(wc)
	for i := 0; i < wc; i++ {
		go func() {
			for tr := range in {
				file, err := transformer(tr.File)
				out <- TransformResult{File: file, Err: err}
			}
			print("\rFor Loop Done. \n")
			defer wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		print("\rWait Groups Done. \n")
		close(out)
	}()
	return out
}

func calcHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error encoutered while opening file: \"%v\"", err)
		return "", err
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
