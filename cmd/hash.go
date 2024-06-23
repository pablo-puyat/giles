package cmd

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"giles/database"
	"giles/models"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
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
		gen := func() <-chan models.FileData {
			r, err := database.GetFilesWithoutHash(db)
			if err != nil {
				log.Print("No files to hash")
				return nil
			}

			out := make(chan models.FileData)
			go func() {
				for _, i := range r {
					out <- i
				}
				close(out)
			}()
			return out
		}()
		h := transform(gen, calculateSHA256, db)
		sqr := transform(h, database.InsertHash, db)
		transform(sqr, database.InsertFileIdHashId, db)
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.Flags().IntP("workers", "w", 1, "Number of workers to use")
}

func calculateSHA256(db *sql.DB, file models.FileData) models.FileData {
	f, err := os.Open(file.Path)
	if err != nil {
		log.Printf("Error encoutered while opening file: \"%v\"", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("Error encoutered while hashing file: \"%v\"", err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))
	file.Hash = hash
	return file
}

func transform(in <-chan models.FileData, transformer func(*sql.DB, models.FileData) models.FileData, db *sql.DB) <-chan models.FileData {
	out := make(chan models.FileData)
	go func() {
		for file := range in {
			out <- transformer(db, file)
		}
		close(out)
	}()
	return out
}
