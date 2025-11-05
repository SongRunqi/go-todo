package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:   "complete <id>",
	Short: "Mark a todo as completed",
	Long: `Mark a task as completed and move it to the backup archive.
The task will be removed from the active list and stored in the backup file.`,
	Example: `  todo complete 1
  todo complete 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s' (must be a number)\n", args[0])
			os.Exit(1)
		}

		task := &app.TodoItem{TaskID: id}
		if err := app.Complete(todos, task, store); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
