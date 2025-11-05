package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <content>",
	Short: "Update a todo",
	Long: `Update an existing todo using Markdown or JSON format.
The content should include the task ID and the fields to update.`,
	Example: `  # Using Markdown
  todo update "# Updated Task

- **Task ID:** 1
- **Task Name:** Updated Task Name
- **Status:** pending
- **Urgency:** high

## Description

Updated description here"

  # Using JSON
  todo update '{"taskId": 1, "taskName": "Updated", "status": "pending"}'`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		content := args[0]

		if err := app.UpdateTask(todos, content, store); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
