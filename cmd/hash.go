package cmd

import (
	"crypto/sha256"
	"database/sql"
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
		db, err := sql.Open("sqlite3", "./giles.db")
		if err != nil {
			panic(fmt.Errorf("error opening database: %v", err))
		}
		files, err := database.GetFilesWithoutHash(db)
		if err != nil {
			log.Printf("Error with query: %v", err)
			return
		}

		c1 := make(chan models.FileData)
		c2 := transform(c1, insertHash, db)
		c3 := transform(c2, database.InsertFileIdHashId, db)

		go func() {
			for result := range c3 {
				fmt.Printf("final --- %d\n", result.Id)
				println("Channel closed")
			}
		}()

		for _, i := range files {
			c1 <- i
		}
		close(c1)

		print("\rDone. \n\nProecessed ", len(files), " files\n")
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 1, "Number of workers to use")
}

func insertHash(db *sql.DB, file models.FileData) models.FileData {
	file.Hash = calcHash(file.Path)
	hashId := database.InsertHash(db, file)
	file.HashId = hashId.HashId
	return file
}

func transform(in <-chan models.FileData, transformer func(*sql.DB, models.FileData) models.FileData, db *sql.DB) <-chan models.FileData {
	out := make(chan models.FileData)
	go func() {
		for file := range in {
			fmt.Printf("transform --- %d\n", file.Id)
			out <- transformer(db, file)
		}
		close(out)
	}()
	return out
}

func calcHash(path string) (hash string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error encoutered while opening file: \"%v\"", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("Error encoutered while hashing file: \"%v\"", err)
	}
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}
