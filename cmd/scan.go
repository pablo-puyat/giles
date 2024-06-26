package cmd

import (
	"giles/database"
	"giles/models"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan files in a directory",
	Long: `Recursively scan files in a directory and insert them into a database.
The name, path and size are recorded.

Usage: giles scan <directory>`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := "."
		if len(args) > 0 {
			dirPath = args[0]
		}
		scanDir(dirPath)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func scanDir(path string) {
	dbManager := database.NewDataStore()
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking path: \"%v\"", err)
		}

		if info.IsDir() {
			return nil
		}
		fileData := models.FileData{
			Name: info.Name(),
			Path: path,
			Size: info.Size(),
		}

		_, err = dbManager.InsertFile(fileData)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error scanning directory: %v", err)
	}
}
