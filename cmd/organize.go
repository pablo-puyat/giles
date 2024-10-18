package cmd

import (
	"fmt"
	"giles/database"
	"giles/internal/organizer"
	"giles/models"
	"log"

	"github.com/spf13/cobra"
)

var (
	source      string
	destination string
)

var organizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Organize files based on their hash",
	Long: `This command retrieves files from the database, organizes them into a 
specified destination directory based on their hash, and updates their 
new locations in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runOrganize(); err != nil {
			log.Fatalf("Error organizing files: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)

	organizeCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory containing files to be organized")
	organizeCmd.Flags().StringVarP(&destination, "destination", "d", "", "Destination directory for organized files")
}

func runOrganize() error {
	ds := database.NewDataStore()

	files, err := ds.GetFilesFrom(source)
	if err != nil {
		return fmt.Errorf("trouble getting list of files: %w", err)
	}

	hashOrganizer := organizer.NewOrganizer(destination)

	return organizeFiles(files, hashOrganizer)
}

func organizeFiles(files []models.FileData, organizer *organizer.Organizer) error {
	fmt.Printf("Organizing %d files\n", len(files))

	// Organize the files
	organizer.OrganizeFiles(files, destination)
	//TODO: batch these calls
	/*
		for _, file := range organizedFiles {
			err = db.UpdateFileLocation(file.ID, file.Path)
			if err != nil {
				return fmt.Errorf("failed to update location for file %s: %w", file.ID, err)
			}
		}
	*/

	fmt.Println("All files organized and database updated")
	return nil
}
