package cmd

import (
	"fmt"
	"os"
	"log"
	"giles/internal/list"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"giles/internal/database"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all files",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listFiles(); err != nil {
			log.Fatalf("Error listing files: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listFiles() error {
	store, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("trouble getting list of files: %w", err)
	}
	defer func(store *database.FileStore) {
		err := store.Close()
		if err != nil {
			log.Println("error closing database")
		}
	}(store)

	files, err := store.GetFiles()

	model := list.NewTableModel(files, store)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	return nil
}
