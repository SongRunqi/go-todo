package app

import "github.com/SongRunqi/go-todo/internal/domain"

// Re-export domain types for backward compatibility
// This allows existing code to continue using app.TodoItem
type TodoItem = domain.TodoItem
type OccurrenceRecord = domain.OccurrenceRecord
type TodoStore = domain.TodoStore

// AlfredResponse the json structure "return" to Alfred
// Alfred1
type AlfredResponse struct {
	Items []AlfredItem `json:"items"`
}

// AlfredItem Alfred json item
type AlfredItem struct {
	UID          string          `json:"uid,omitempty"`
	Title        string          `json:"title"`
	Subtitle     string          `json:"subtitle,omitempty"`
	Arg          string          `json:"arg,omitempty"`
	Autocomplete string          `json:"autocomplete,omitempty"`
	Icon         *Icon           `json:"icon,omitempty"`
	Text         *AlfredItemText `json:"text"`
}

// Icon represents the icon for an item
type Icon struct {
	Path string `json:"path"`
}

type AlfredItemText struct {
	Copy      string `json:"copy"`
	Largetype string `json:"largetype"`
}

// The expect response from AI

// IntentResponse
type IntentResponse struct {
	Intent string     `json:"intent"`
	Tasks  []TodoItem `json:"tasks,omitempty"`
}

// AI provider

// openai response
type OpenAIResponse struct {
	Choices []OpenAIChoices `json:"choices"`
}

type OpenAIChoices struct {
	Message Msg `json:"message"`
}

type Msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openai request
type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []Msg  `json:"messages"`
}

type AgentCommand struct {
	Name string `json:"name"`
	Args []string
}

type FileTodoStore struct {
	Path       string
	BackupPath string
}
