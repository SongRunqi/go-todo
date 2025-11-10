package cmd

import (
	"fmt"
	"os"

	"github.com/SongRunqi/go-todo/internal/i18n"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/SongRunqi/go-todo/internal/updater"
	"github.com/SongRunqi/go-todo/internal/version"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to the latest version",
	Long: `Check for updates and upgrade to the latest version from GitHub Releases.

The upgrade command will:
  - Check GitHub Releases for the latest version
  - Download the appropriate binary for your platform
  - Verify the binary using SHA256 checksum
  - Replace the current binary with the new version
  - Create a backup before replacing (auto-rollback on failure)`,
	Example: `  # Check for updates without installing
  todo upgrade --check

  # Upgrade to the latest version
  todo upgrade`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent's PersistentPreRun to skip todo loading
		// Initialize i18n
		if err := i18n.Init(""); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize i18n: %v\n", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		runUpgrade()
	},
}

func init() {
	upgradeCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates without installing")
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade() {
	u := updater.New()

	// Check for updates
	spinner := output.NewSpinner(i18n.T("upgrade.checking"))
	spinner.Start()

	release, hasUpdate, err := u.CheckForUpdates()
	spinner.Stop()

	if err != nil {
		output.PrintError(i18n.T("upgrade.check_failed", err.Error()))
		os.Exit(1)
	}

	currentVersion := version.GetInfo().Short()

	if !hasUpdate {
		output.PrintSuccess(i18n.T("upgrade.up_to_date", currentVersion))
		return
	}

	latestVersion := release.TagName
	output.PrintInfo(i18n.T("upgrade.new_version_available", currentVersion, latestVersion))

	if checkOnly {
		fmt.Println()
		fmt.Printf("%s: %s\n", i18n.T("upgrade.release_notes"), release.Name)
		if release.Body != "" {
			fmt.Println(release.Body)
		}
		return
	}

	// Ask for confirmation
	fmt.Println()
	fmt.Printf("%s %s -> %s\n", i18n.T("upgrade.confirm_prompt"), currentVersion, latestVersion)
	fmt.Printf("%s (y/N): ", i18n.T("upgrade.continue"))

	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
		output.PrintInfo(i18n.T("upgrade.cancelled"))
		return
	}

	// Perform upgrade
	spinner = output.NewSpinner(i18n.T("upgrade.downloading"))
	spinner.Start()

	err = u.Update()
	spinner.Stop()

	if err != nil {
		output.PrintError(i18n.T("upgrade.failed", err.Error()))
		os.Exit(1)
	}

	output.PrintSuccess(i18n.T("upgrade.success", latestVersion))
	output.PrintInfo(i18n.T("upgrade.restart_required"))
}
