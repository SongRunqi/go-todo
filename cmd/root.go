package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

var (
	cfgFile string
	verbose bool
)

// Global context shared across commands
var (
	store       *app.FileTodoStore
	todos       *[]app.TodoItem
	config      app.Config
	currentTime time.Time
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "todo [natural language input]",
	Short: "AI-powered todo management CLI",
	Long: `Todo-Go is an AI-powered command-line todo management application.
It supports both structured commands and natural language input powered by LLM.

Examples:
  # Natural language (AI-powered)
  todo "Buy groceries tomorrow evening"
  todo "Write report by Friday; Call client tomorrow"

  # Structured commands
  todo list
  todo get 1
  todo complete 1`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize configuration
		config = app.LoadConfig()

		// Initialize store
		store = &app.FileTodoStore{
			Path:       config.TodoPath,
			BackupPath: config.BackupPath,
		}

		// Load todos
		var err error
		loadedTodos, err := store.Load(false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", err)
			os.Exit(1)
		}
		todos = &loadedTodos
		currentTime = time.Now()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided and args exist, treat as natural language
		if len(args) > 0 {
			handleNaturalLanguage(args)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.todo/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// handleNaturalLanguage processes natural language input using AI
func handleNaturalLanguage(args []string) {
	userInput := args[0]

	ctx := &app.Context{
		Store:       store,
		Todos:       todos,
		Args:        append([]string{"todo"}, userInput),
		CurrentTime: currentTime,
		Config:      &config,
	}

	// Use AICommand to process natural language
	aiCmd := &app.AICommand{}
	if err := aiCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
