package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	dbPath string
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "giles",
		Short: "Giles is a tool to manage files.",
		Long: `Giles is a CLI / TUI to manage media files.  It can scan directories for files, find duplicates, and manage metadata.
Usage: 
- giles scan <path> 
`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add persistent flag to root command
	rootCmd.PersistentFlags().StringVar(
		&dbPath,
		"database",
		"",
		"path to SQLite database file",
	)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
