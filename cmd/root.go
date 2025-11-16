package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// AppContext holds the application context shared across commands
type AppContext struct {
	Store       *app.FileTodoStore
	Todos       *[]app.TodoItem
	Config      app.Config
	CurrentTime time.Time
}

var (
	descriptionsOnce                 sync.Once
	updateSubcommandDescriptionsFunc func()
)

// getAppContext retrieves the AppContext from the command context
func getAppContext(cmd *cobra.Command) *AppContext {
	return cmd.Context().Value("appContext").(*AppContext)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "todo command",
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "info"
		}
		logger.Init(logLevel)

		// Initialize configuration
		config := app.LoadConfig()

		// Initialize i18n (may have been initialized in init(), reinit with config language)
		if config.Language != "" {
			if err := i18n.SetLanguage(config.Language); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set language: %v\n", err)
			}
		}

		// Update subcommand descriptions (once) after all commands are registered
		if updateSubcommandDescriptionsFunc != nil {
			descriptionsOnce.Do(updateSubcommandDescriptionsFunc)
		}

		// Initialize store
		store := &app.FileTodoStore{
			Path:       config.TodoPath,
			BackupPath: config.BackupPath,
		}

		// Load todos
		loadedTodos, err := store.Load(false)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.loading_todos"), err)
			os.Exit(1)
		}
		// Allocate a new slice on the heap to avoid dangling pointer
		todosList := loadedTodos
		todos := &todosList
		currentTime := time.Now()

		// Create AppContext and attach it to the command context
		appCtx := &AppContext{
			Store:       store,
			Todos:       todos,
			Config:      config,
			CurrentTime: currentTime,
		}

		ctx := context.WithValue(cmd.Context(), "appContext", appCtx)
		cmd.SetContext(ctx)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Silence errors in rootCmd so we can handle them ourselves
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Initialize i18n early for command descriptions
	// Priority: 1. Config file 2. Auto-detect
	cfg := app.LoadConfig()
	if err := i18n.Init(cfg.Language); err != nil {
		// Silently fall back to English if i18n fails during init
		// This is acceptable since init() can't easily report errors
	}

	// Set command descriptions
	rootCmd.Short = i18n.T("cmd.root.short")
	rootCmd.Long = i18n.T("cmd.root.long")
	completionCmd.Short = i18n.T("cmd.root.completion.short")
	completionCmd.Long = i18n.T("cmd.root.completion.long")

	// Add completion command
	rootCmd.AddCommand(completionCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.todo/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Define the updateSubcommandDescriptionsFunc after rootCmd is initialized
	updateSubcommandDescriptionsFunc = func() {
		for _, c := range rootCmd.Commands() {
			switch c.Name() {
			case "list":
				c.Short = i18n.T("cmd.list.short")
				c.Long = i18n.T("cmd.list.long")
			case "get":
				c.Short = i18n.T("cmd.get.short")
				c.Long = i18n.T("cmd.get.long")
			case "complete":
				c.Short = i18n.T("cmd.complete.short")
				c.Long = i18n.T("cmd.complete.long")
			case "delete":
				c.Short = i18n.T("cmd.delete.short")
				c.Long = i18n.T("cmd.delete.long")
			case "init":
				c.Short = i18n.T("cmd.init.short")
				c.Long = i18n.T("cmd.init.long")
			case "update":
				c.Short = i18n.T("cmd.update.short")
				c.Long = i18n.T("cmd.update.long")
			case "back":
				c.Short = i18n.T("cmd.back.short")
				c.Long = i18n.T("cmd.back.long")
				// Set back subcommands
				for _, sc := range c.Commands() {
					switch sc.Name() {
					case "get":
						sc.Short = i18n.T("cmd.back.get.short")
						sc.Long = i18n.T("cmd.back.get.long")
					case "restore":
						sc.Short = i18n.T("cmd.back.restore.short")
						sc.Long = i18n.T("cmd.back.restore.long")
					}
				}
			case "lang":
				c.Short = i18n.T("cmd.lang.short")
				c.Long = i18n.T("cmd.lang.long")
				// Set lang subcommands
				for _, sc := range c.Commands() {
					switch sc.Name() {
					case "list":
						sc.Short = i18n.T("cmd.lang.list.short")
						sc.Long = i18n.T("cmd.lang.list.long")
					case "set":
						sc.Short = i18n.T("cmd.lang.set.short")
						sc.Long = i18n.T("cmd.lang.set.long")
					case "current":
						sc.Short = i18n.T("cmd.lang.current.short")
						sc.Long = i18n.T("cmd.lang.current.long")
					}
				}
			}
		}
	}

	// Call it once now to set descriptions for all already-registered subcommands
	updateSubcommandDescriptionsFunc()
}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "",
	Long:                  "",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
