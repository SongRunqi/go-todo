package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/SongRunqi/go-todo/app"
	"github.com/spf13/cobra"
)

// tasktagCmd represents the tasktag command
var tasktagCmd = &cobra.Command{
	Use:   "tasktag",
	Short: "Manage task tags",
	Long:  "Add or remove tags from tasks",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// tasktagAddCmd adds a tag to a task
var tasktagAddCmd = &cobra.Command{
	Use:   "add <task-id> <tag-name>",
	Short: "Add a tag to a task",
	Long:  "Add a tag to a task by task ID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID: %v\n", err)
			os.Exit(1)
		}
		tagName := args[1]
		if err := app.AddTagToTask(ctx.Todos, taskID, tagName, ctx.Store, ctx.TagStore); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// tasktagRemoveCmd removes a tag from a task
var tasktagRemoveCmd = &cobra.Command{
	Use:   "remove <task-id> <tag-name>",
	Short: "Remove a tag from a task",
	Long:  "Remove a tag from a task by task ID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID: %v\n", err)
			os.Exit(1)
		}
		tagName := args[1]
		if err := app.RemoveTagFromTask(ctx.Todos, taskID, tagName, ctx.Store, ctx.TagStore); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tasktagCmd)
	tasktagCmd.AddCommand(tasktagAddCmd)
	tasktagCmd.AddCommand(tasktagRemoveCmd)
}
