package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
)

// Context holds the execution context for commands
type Context struct {
	Store       *FileTodoStore
	Todos       *[]TodoItem
	Args        []string
	CurrentTime time.Time
	Config      *Config
}

// Command interface for all commands
type Command interface {
	Execute(ctx *Context) error
}

// Router handles command routing
type Router struct {
	commands map[string]Command
}

// NewRouter creates a new router with registered commands
func NewRouter() *Router {
	r := &Router{
		commands: make(map[string]Command),
	}

	// Register all commands
	r.Register("list", &ListCommand{})
	r.Register("ls", &ListCommand{})
	r.Register("back", &BackCommand{})
	r.Register("back get", &BackGetCommand{})
	r.Register("back restore", &BackRestoreCommand{})
	r.Register("complete", &CompleteCommand{})
	r.Register("delete", &DeleteCommand{})
	r.Register("get", &GetCommand{})
	r.Register("update", &UpdateCommand{})

	return r
}

// Register registers a command
func (r *Router) Register(name string, cmd Command) {
	r.commands[name] = cmd
}

// Route routes the user input to the appropriate command
func (r *Router) Route(ctx *Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("no command provided")
	}

	userInput := ctx.Args[1]
	logger.Debugf("Router received user input: %s", userInput)

	// Try to match exact commands first
	if cmd, ok := r.commands[userInput]; ok {
		return cmd.Execute(ctx)
	}

	// Try to match compound commands (e.g., "back get", "back restore")
	args := strings.Split(userInput, " ")
	if len(args) >= 2 {
		compoundKey := strings.Join(args[:2], " ")
		if cmd, ok := r.commands[compoundKey]; ok {
			return cmd.Execute(ctx)
		}
	}

	// Try to match simple commands with arguments
	if len(args) > 1 {
		if cmd, ok := r.commands[args[0]]; ok {
			return cmd.Execute(ctx)
		}
	}

	// If no command matched, use AI command
	aiCmd := &AICommand{}
	return aiCmd.Execute(ctx)
}
