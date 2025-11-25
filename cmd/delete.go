package cmd

import (
	"fmt"

	"github.com/SongRunqi/go-todo/app"
	_ "github.com/SongRunqi/go-todo/internal/i18n"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete --id <id> [--source backup]",
	Short: "",
	Long:  "",
	Example: `todo delete --id 3
todo delete --id 3 --source backup`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getAppContext(cmd)
		return runDelete(ctx, deleteID, deleteSource)
	},
}

var (
	deleteID     int
	deleteSource string
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().IntVarP(&deleteID, "id", "i", 0, "Task ID to delete")
	deleteCmd.Flags().StringVarP(&deleteSource, "source", "s", "active", "Source to delete from: active|backup")
	_ = deleteCmd.MarkFlagRequired("id")
}

func runDelete(ctx *AppContext, id int, source string) error {
	switch source {
	case "active":
		return app.DeleteTask(ctx.Todos, id, ctx.Store)
	case "backup":
		return app.DeleteBackupTask(id, ctx.Store)
	default:
		return fmt.Errorf("invalid source %q (use active|backup)", source)
	}
}
