package cmd

import (
	"fmt"
	"os"

	"github.com/SongRunqi/go-todo/app"
	"github.com/spf13/cobra"
)

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
	Long:  "Manage tags that can be associated with tasks",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// tagListCmd lists all tags
var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	Long:  "List all available tags",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		if err := app.ListTags(ctx.TagStore); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// tagAddCmd adds a new tag
var tagAddCmd = &cobra.Command{
	Use:   "add <name> [color]",
	Short: "Add a new tag",
	Long:  "Add a new tag with an optional color",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		name := args[0]
		color := ""
		if len(args) > 1 {
			color = args[1]
		}
		if err := app.AddTag(ctx.TagStore, name, color); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// tagDeleteCmd deletes a tag
var tagDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a tag",
	Long:  "Delete a tag and remove it from all tasks",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		name := args[0]
		if err := app.DeleteTag(ctx.TagStore, ctx.Store, name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// tagRenameCmd renames a tag
var tagRenameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a tag",
	Long:  "Rename a tag and update it in all tasks",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getAppContext(cmd)
		oldName := args[0]
		newName := args[1]
		if err := app.RenameTag(ctx.TagStore, ctx.Store, oldName, newName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagAddCmd)
	tagCmd.AddCommand(tagDeleteCmd)
	tagCmd.AddCommand(tagRenameCmd)
}
