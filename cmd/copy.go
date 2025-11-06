package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

var (
	copyWeek bool
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy completed tasks to clipboard",
	Long:  "Copy completed tasks to clipboard, grouped by week",
	Run: func(cmd *cobra.Command, args []string) {
		if err := app.CopyCompletedTasks(todos, store, copyWeek); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().BoolVarP(&copyWeek, "week", "w", false, "Copy tasks from current week only")
}
