package cmd

import (
	"github.com/spf13/cobra"
)

var duplicatesCmd = &cobra.Command{
	Use:   "duplicates",
	Short: "Show duplicate files",
	Long: `Show duplicate files in the database.

Usage: giles duplicates [list] [delete] [rename]`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(duplicatesCmd)
}

func renameDuplicates() {

}
