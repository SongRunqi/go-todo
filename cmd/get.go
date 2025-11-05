package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a specific todo by ID",
	Long: `Retrieve and display detailed information about a specific todo.
Output is in Markdown format for easy editing.`,
	Example: `  todo get 1
  todo get 42`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s' (must be a number)\n", args[0])
			os.Exit(1)
		}

		if err := app.GetTask(todos, id); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
