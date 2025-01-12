package cmd

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

var (
	dbPath  string
	verbose bool
	logFile string

	rootCmd = &cobra.Command{
		Use:   "giles",
		Short: "Giles is a tool to manage files.",
		Long: `Giles is a CLI / TUI to manage media files.  It can scan directories for files, find duplicates, and manage metadata.
Usage: 
- giles scan <path> 
`,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetOutput(io.Discard)

	rootCmd.PersistentFlags().StringVar(
		&dbPath,
		"database",
		"",
		"path to SQLite database file",
	)

	rootCmd.PersistentFlags().BoolVar(
		&verbose,
		"verbose",
		false,
		"enable verbose logging to stdout",
	)

	rootCmd.PersistentFlags().StringVar(
		&logFile,
		"log",
		"",
		"path to log file",
	)

	cobra.OnInitialize(func() {
		if logFile != "" {
			file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			log.SetOutput(file)
		} else if verbose {
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(io.Discard)
		}
	})

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
