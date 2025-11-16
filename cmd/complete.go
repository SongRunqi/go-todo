package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:   "complete <id>",
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

		task := &app.TodoItem{TaskID: id}
		if err := app.Complete(ctx.Todos, task, ctx.Store); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
