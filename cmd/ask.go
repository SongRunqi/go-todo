package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/SongRunqi/go-todo/app"
	"github.com/SongRunqi/go-todo/internal/i18n"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type OpenAIRequest = app.OpenAIRequest
type Msg = app.Msg

// askCmd wraps the natural language handler so it behaves like a normal command.
var askCmd = &cobra.Command{
	Use:   "ask [natural language input]",
	Short: "",
	Long:  "",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := ask(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cmd.root.error.general"), err)
			os.Exit(1)

		}
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}

func ask(args []string) error {
	cfg := app.LoadConfig()

	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05")
	weekday := now.Weekday()

	userLanguage := "English" // default
	if cfg.Language == "zh-CN" || cfg.Language == "zh" {
		userLanguage = "Chinese"
	} else if cfg.Language == "en" || cfg.Language == "en-US" {
		userLanguage = "English"
	}

	// load todos
	fileStore := app.FileTodoStore{cfg.TodoPath, cfg.BackupPath}
	load, err := fileStore.Load(false)
	if err != nil {
		fmt.Println(err)
	}

	// Build context in XML format for better structure and clarity
	bytes, _ := json.Marshal(load)
	contextStr := fmt.Sprintf(`<context>
	<current_time>%s</current_time>
	<weekday>%s</weekday>
	<user_preferred_language>%s</user_preferred_language>
	<user_input>%s</user_input>
	<user_todos>%s</user_todos>
</context>`, nowStr, weekday, userLanguage, args[0], string(bytes))

	logger.Debugf("AI context: %s", contextStr)

	req := OpenAIRequest{
		Model: cfg.Model,
		Messages: []Msg{
			{Role: "system", Content: app.Cmd},
			{Role: "user", Content: contextStr},
		},
	}

	log.Info().Msgf("AI request: %s", req.Messages[1].Content)
	// Show spinner during AI request
	spin := output.NewAISpinner()
	spin.Start()

	warpIntend, err := app.Chat(req)
	log.Info().Msgf("AI response: %s", warpIntend)
	spin.Stop()

	if err != nil {
		output.PrintErrorWithSuggestion(
			fmt.Sprintf("AI request failed: %v", err),
			"Check your API_KEY environment variable and network connection",
		)
		return fmt.Errorf("AI request failed: %w", err)
	}

	return app.DoI(warpIntend, &load, &fileStore)

}
