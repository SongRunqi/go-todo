package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all active todos",
	Long: `Display all active todos in Alfred-compatible JSON format.
Tasks are sorted by deadline with time remaining displayed.`,
	Example: `  todo list
  todo ls`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := app.List(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing todos: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
