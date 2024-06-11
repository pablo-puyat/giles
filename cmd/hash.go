package cmd

import (
	"github.com/spf13/cobra"
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Hash files in the database",
	Long: `Create hash for files in database that do not have one.

Usage: giles hash`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		print("Hello")
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
}
