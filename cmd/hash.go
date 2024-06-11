package cmd

import (
	"crypto/sha256"
	"fmt"
	"giles/database"
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
		findHashes()
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
}

func findHashes() {
	db, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	files, err := db.GetFilesWithoutHash()
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	for _, file := range files {
		hash, err := hash(file.Path)
		if err != nil {
			log.Printf("Error hashing %s: %v", file.Path, err)
			continue
		}
		err = db.UpdateFileHash(file.Path, hash)
		if err != nil {
			log.Printf("Error updating hash for %s: %v", file.Path, err)
		}
	}
}

func hash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
