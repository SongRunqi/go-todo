package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/logger"
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
		// Initialize logger
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "info"
		}
		logger.Init(logLevel)

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
	// Add completion command
	rootCmd.AddCommand(completionCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.todo/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for todo.

To load completions:

Bash:
  $ source <(todo completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ todo completion bash > /etc/bash_completion.d/todo
  # macOS:
  $ todo completion bash > $(brew --prefix)/etc/bash_completion.d/todo

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ todo completion zsh > "${fpath[1]}/_todo"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ todo completion fish | source
  # To load completions for each session, execute once:
  $ todo completion fish > ~/.config/fish/completions/todo.fish

PowerShell:
  PS> todo completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> todo completion powershell > todo.ps1
  # and source this file from your PowerShell profile.
`,
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
