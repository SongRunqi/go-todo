package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

var (
	compactPeriod string
)

// compactCmd represents the compact command
var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Compact and summarize completed/deleted tasks",
	Long:  "Compact and summarize completed and deleted tasks from backup by week or month",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		if err := app.CompactTasks(ctx.Store, compactPeriod); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(compactCmd)
	compactCmd.Flags().StringVarP(&compactPeriod, "period", "p", "week", "Grouping period: 'week' or 'month'")
}
