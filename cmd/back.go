package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

// backCmd represents the back command
var backCmd = &cobra.Command{
	Use:   "back",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		// Default: list backup todos
		backupTodos, err := ctx.Store.Load(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}

		if err := app.List(&backupTodos); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

// backGetCmd represents the "back get" command
var backGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.invalid_task_id"), args[0])
			os.Exit(1)
		}

		backupTodos, err := ctx.Store.Load(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}

		if err := app.GetTask(&backupTodos, id); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

// backRestoreCmd represents the "back restore" command
var backRestoreCmd = &cobra.Command{
	Use:   "restore <id>",
	Short: "",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.invalid_task_id"), args[0])
			os.Exit(1)
		}

		backupTodos, err := ctx.Store.Load(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}

		if err := app.RestoreTask(ctx.Todos, &backupTodos, id, ctx.Store); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backCmd)
	backCmd.AddCommand(backGetCmd)
	backCmd.AddCommand(backRestoreCmd)
}
