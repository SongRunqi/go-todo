package app

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/rs/zerolog/log"
)

// ListCommand lists all active todos
type ListCommand struct{}

func (c *ListCommand) Execute(ctx *Context) error {
	return List(ctx.Todos)
}

// BackCommand lists all backup/completed todos
type BackCommand struct{}

func (c *BackCommand) Execute(ctx *Context) error {
	backupTodos, err := ctx.Store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup todos: %w", err)
	}
	return List(&backupTodos)
}

// BackGetCommand gets a task from backup
type BackGetCommand struct{}

func (c *BackGetCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 3 {
		return fmt.Errorf("usage: back get <id>")
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	backupTodos, err := ctx.Store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup todos: %w", err)
	}

	return GetTask(&backupTodos, id)
}

// BackRestoreCommand restores a task from backup
type BackRestoreCommand struct{}

func (c *BackRestoreCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 3 {
		return fmt.Errorf("usage: back restore <id>")
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	backupTodos, err := ctx.Store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup todos: %w", err)
	}

	return RestoreTask(ctx.Todos, &backupTodos, id, ctx.Store)
}

// CompleteCommand marks a task as complete
type CompleteCommand struct{}

func (c *CompleteCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 2 {
		return fmt.Errorf("usage: complete <id>")
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	return Complete(ctx.Todos, &TodoItem{TaskID: id}, ctx.Store)
}

// DeleteCommand deletes a task
type DeleteCommand struct{}

func (c *DeleteCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 2 {
		return fmt.Errorf("usage: delete <id>")
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	return DeleteTask(ctx.Todos, id, ctx.Store)
}

// GetCommand gets a task by ID
type GetCommand struct{}

func (c *GetCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 2 {
		return fmt.Errorf("usage: get <id>")
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	return GetTask(ctx.Todos, id)
}

// UpdateCommand updates a task
type UpdateCommand struct{}

func (c *UpdateCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 2 {
		return fmt.Errorf("usage: update <task_content>")
	}

	// Preserve formatting by using original input after "update "
	todoContent := strings.TrimPrefix(ctx.Args[1], "update ")
	return UpdateTask(ctx.Todos, todoContent, ctx.Store)
}

// AICommand uses AI to process natural language input
type AICommand struct{}

func (c *AICommand) Execute(ctx *Context) error {
	nowStr := ctx.CurrentTime.Format(time.RFC3339)
	weekday := ctx.CurrentTime.Weekday().String()

	// Determine user's preferred language for task creation
	userLanguage := "English" // default
	if ctx.Config.Language == "zh-CN" || ctx.Config.Language == "zh" {
		userLanguage = "Chinese"
	} else if ctx.Config.Language == "en" || ctx.Config.Language == "en-US" {
		userLanguage = "English"
	}

	// Build context in XML format for better structure and clarity
	bytes, _ := json.Marshal(ctx.Todos)
	contextStr := fmt.Sprintf(`<context>
	<current_time>%s</current_time>
	<weekday>%s</weekday>
	<user_preferred_language>%s</user_preferred_language>
	<user_input>%s</user_input>
	<user_todos>%s</user_todos>
</context>`, nowStr, weekday, userLanguage, ctx.Args[1], string(bytes))

	logger.Debugf("AI context: %s", contextStr)

	req := OpenAIRequest{
		Model: ctx.Config.Model,
		Messages: []Msg{
			{Role: "system", Content: Cmd},
			{Role: "user", Content: contextStr},
		},
	}

	log.Info().Msgf("AI request: %s", req.Messages[1].Content)
	// Show spinner during AI request
	spin := output.NewAISpinner()
	spin.Start()

	warpIntend, err := Chat(req)
	log.Info().Msgf("AI response: %s", warpIntend)
	spin.Stop()

	if err != nil {
		output.PrintErrorWithSuggestion(
			fmt.Sprintf("AI request failed: %v", err),
			"Check your API_KEY environment variable and network connection",
		)
		return fmt.Errorf("AI request failed: %w", err)
	}

	return DoI(warpIntend, ctx.Todos, ctx.Store)
}
