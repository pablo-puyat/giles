package cmd

import (
	"fmt"
	"giles/internal/database"
	"giles/internal/organizer"
	"github.com/spf13/cobra"
	"log"
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
	return nil
	/*
		ds, err := database.New(databasePath)
		if err != nil {
			return fmt.Errorf("trouble getting list of files: %w", err)
		}

		files, err := ds.GetFilesFrom(source)
		if err != nil {
			return fmt.Errorf("trouble getting list of files: %w", err)
		}

		hashOrganizer := organizer.NewOrganizer(destination)

		return organizeFiles(files, hashOrganizer)
	*/
}

func organizeFiles(files []database.File, organizer *organizer.Organizer) error {
	organized := organizer.OrganizeFiles(files, destination)
	fmt.Printf("Organizing %d files\n", len(organized))
	//newPath := generateNewPath(file.Path) // Implement this function based on your renaming logic
	//err := os.Rename(file.Path, newPath)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to rename file %s: %w", file.Path, err)
	//}
	//files[i].Path = newPath
	//TODO: batch these calls
	//for _, file := range organized {
	//fmt.Println(file.Name)
	//fmt.Println(file.Path)
	//err = db.UpdateFileLocation(file.ID, file.Path)
	//if err != nil {
	//	return fmt.Errorf("failed to update location for file %s: %w", file.ID, err)
	//}
	//}

	fmt.Println("All files organized and database updated")
	return nil
}
