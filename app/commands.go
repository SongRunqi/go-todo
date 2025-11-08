package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
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
	contextStr := fmt.Sprintf(`<context>
	<current_time>%s</current_time>
	<weekday>%s</weekday>
	<user_preferred_language>%s</user_preferred_language>
	<user_input>%s</user_input>
</context>`, nowStr, weekday, userLanguage, ctx.Args[1])

	logger.Debugf("AI context: %s", contextStr)

	req := OpenAIRequest{
		Model: ctx.Config.Model,
		Messages: []Msg{
			{Role: "system", Content: cmd},
			{Role: "user", Content: contextStr},
		},
	}

	// Show spinner during AI request
	spin := output.NewAISpinner()
	spin.Start()

	warpIntend, err := Chat(req)
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

// ReminderSetCommand sets reminder times for a task
type ReminderSetCommand struct{}

func (c *ReminderSetCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 3 {
		return fmt.Errorf("usage: reminder set <id> <duration1> [duration2 ...]")
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	// Parse reminder durations
	var reminderTimes []time.Duration
	for i := 2; i < len(args); i++ {
		duration, err := parseDuration(args[i])
		if err != nil {
			return fmt.Errorf("invalid duration '%s': %w", args[i], err)
		}
		// Convert to negative duration (reminder is before the event)
		reminderTimes = append(reminderTimes, -duration)
	}

	if len(reminderTimes) == 0 {
		return fmt.Errorf("at least one reminder duration is required")
	}

	return SetReminders(ctx.Todos, id, reminderTimes, ctx.Store)
}

// ReminderEnableCommand enables reminders for a task
type ReminderEnableCommand struct{}

func (c *ReminderEnableCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 2 {
		return fmt.Errorf("usage: reminder enable <id>")
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	return EnableReminders(ctx.Todos, id, ctx.Store)
}

// ReminderDisableCommand disables reminders for a task
type ReminderDisableCommand struct{}

func (c *ReminderDisableCommand) Execute(ctx *Context) error {
	args := strings.Split(ctx.Args[1], " ")
	if len(args) < 2 {
		return fmt.Errorf("usage: reminder disable <id>")
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	return DisableReminders(ctx.Todos, id, ctx.Store)
}

// parseDuration parses a duration string like "1h", "30m", "1d", "2h30m"
func parseDuration(s string) (time.Duration, error) {
	// Handle days (not supported by time.ParseDuration)
	var days int
	if strings.Contains(s, "d") {
		parts := strings.Split(s, "d")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid duration format")
		}
		var err error
		days, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		s = parts[1]
	}

	var duration time.Duration
	if s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, err
		}
		duration = d
	}

	duration += time.Duration(days) * 24 * time.Hour
	return duration, nil
}
