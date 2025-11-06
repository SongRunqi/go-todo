package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <content>",
	Short: "",
	Long:  "",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		content := args[0]

		if err := app.UpdateTask(todos, content, store); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
