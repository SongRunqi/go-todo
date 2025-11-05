package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// backCmd represents the back command
var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Manage backup/completed todos",
	Long: `View and manage completed todos from the backup archive.
Use subcommands to list, view, or restore completed tasks.`,
	Example: `  todo back              # List all completed todos
  todo back get 1        # View a completed todo
  todo back restore 1    # Restore a completed todo`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default: list backup todos
		backupTodos, err := store.Load(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading backup: %v\n", err)
			os.Exit(1)
		}

		if err := app.List(&backupTodos); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing backup: %v\n", err)
			os.Exit(1)
		}
	},
}

// backGetCmd represents the "back get" command
var backGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a completed todo from backup",
	Long:  `Retrieve and display detailed information about a completed todo from the backup archive.`,
	Example: `  todo back get 1
  todo back get 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s' (must be a number)\n", args[0])
			os.Exit(1)
		}

		backupTodos, err := store.Load(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading backup: %v\n", err)
			os.Exit(1)
		}

		if err := app.GetTask(&backupTodos, id); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// backRestoreCmd represents the "back restore" command
var backRestoreCmd = &cobra.Command{
	Use:   "restore <id>",
	Short: "Restore a completed todo from backup",
	Long: `Restore a completed task from the backup archive back to the active list.
The task status will be changed to 'pending' and it will be removed from backup.`,
	Example: `  todo back restore 1
  todo back restore 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s' (must be a number)\n", args[0])
			os.Exit(1)
		}

		backupTodos, err := store.Load(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading backup: %v\n", err)
			os.Exit(1)
		}

		if err := app.RestoreTask(todos, &backupTodos, id, store); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backCmd)
	backCmd.AddCommand(backGetCmd)
	backCmd.AddCommand(backRestoreCmd)
}
