package cmd

import (
	"fmt"
	"giles/internal/database"
	"giles/internal/organizer"
	"github.com/spf13/cobra"
	"log"
)

var organizeCmd = &cobra.Command{
	Use:   "organize <source path> <destination path>",
	Short: "Organize files based on their hash",
	Long: `This command retrieves files from the database, organizes them into a 
specified destination directory based on their hash, and updates their 
new locations in the database.`,
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runOrganize(args[0], args[1]); err != nil {
			log.Fatalf("Error organizing files: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)
}

func runOrganize(source string, dest string) error {
	store, err := database.New(databasePath)
	if err != nil {
		return fmt.Errorf("trouble getting list of files: %w", err)
	}
	defer func(store *database.FileStore) {
		err := store.Close()
		if err != nil {
			log.Println("Error closing database")
		}
	}(store)

	files, err := store.GetFilesFrom(source)
	if err != nil {
		return fmt.Errorf("trouble getting list of files: %w", err)
	}

	if err := organizer.ByHash(files, dest); err != nil {
		return err
	}

	return nil
}
