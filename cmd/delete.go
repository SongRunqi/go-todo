package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a todo permanently",
	Long: `Permanently remove a task from the active list.
This action cannot be undone. The task will not be moved to backup.`,
	Example: `  todo delete 1
  todo delete 10`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s' (must be a number)\n", args[0])
			os.Exit(1)
		}

		if err := app.DeleteTask(todos, id, store); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
