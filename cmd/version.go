package cmd

import (
	"fmt"

	"github.com/SongRunqi/go-todo/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display detailed version information including build details",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent's PersistentPreRun to skip todo loading
	},
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetInfo()
		fmt.Println(info.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
