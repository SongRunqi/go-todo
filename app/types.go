package app

import "time"

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

// TodoItem a item of todos
type TodoItem struct {
	TaskID     int       `json:"taskId"`
	CreateTime time.Time `json:"createTime"`
	EndTime    time.Time `json:"endTime"`
	User       string    `json:"user"`
	TaskName   string    `json:"taskName"`
	TaskDesc   string    `json:"taskDesc"`
	Status     string    `json:"status"`
	DueDate    string    `json:"dueDate"`
	Urgent     string    `json:"urgent"`

	// Recurring task fields
	IsRecurring       bool   `json:"isRecurring,omitempty"`       // Whether this is a recurring task
	RecurringType     string `json:"recurringType,omitempty"`     // daily, weekly, monthly, yearly
	RecurringInterval int    `json:"recurringInterval,omitempty"` // Interval (e.g., every 2 days, every 3 weeks)
	CompletionCount   int    `json:"completionCount,omitempty"`   // Number of times completed
}

type TodoStore interface {
	Load(backup bool) ([]TodoItem, error)
	Save(todoItems []TodoItem, backup bool) error
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
