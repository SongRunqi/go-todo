package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "",
	Long:    "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := app.List(todos); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
