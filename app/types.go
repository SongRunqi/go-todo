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
	EndTime    time.Time `json:"endTime"` // For recurring tasks: next scheduled occurrence time
	User       string    `json:"user"`
	TaskName   string    `json:"taskName"`
	TaskDesc   string    `json:"taskDesc"`
	Status     string    `json:"status"` // For recurring tasks: active, paused, completed, cancelled. For non-recurring: pending, completed
	DueDate    string    `json:"dueDate"`
	Urgent     string    `json:"urgent"`

	// Event duration (for tasks with specific time ranges, e.g., "2pm to 3pm")
	EventDuration time.Duration `json:"eventDuration,omitempty"` // Duration of the event (e.g., 1 hour)

	// Recurring task fields
	IsRecurring       bool   `json:"isRecurring,omitempty"`       // Whether this is a recurring task
	RecurringType     string `json:"recurringType,omitempty"`     // daily, weekly, monthly, yearly
	RecurringInterval int    `json:"recurringInterval,omitempty"` // Interval (e.g., every 2 days, every 3 weeks)
	RecurringWeekdays []int  `json:"recurringWeekdays,omitempty"` // For weekly: specific weekdays (0=Sun, 1=Mon...6=Sat). Empty means all days.
	RecurringMaxCount int    `json:"recurringMaxCount,omitempty"` // Maximum number of times to repeat (0 = infinite)
	CompletionCount   int    `json:"completionCount,omitempty"`   // Number of periods completed

	// Occurrence tracking for recurring tasks
	OccurrenceHistory []OccurrenceRecord `json:"occurrenceHistory,omitempty"` // History of all scheduled occurrences

	// Deprecated fields (kept for backward compatibility, will be migrated to OccurrenceHistory)
	CurrentPeriodCompletions []string `json:"currentPeriodCompletions,omitempty"` // DEPRECATED: Use OccurrenceHistory instead
}

// OccurrenceRecord represents a single occurrence/instance of a recurring task
type OccurrenceRecord struct {
	ScheduledTime time.Time `json:"scheduledTime"` // The scheduled time for this occurrence
	Status        string    `json:"status"`        // pending, completed, missed, skipped
	CompletedAt   time.Time `json:"completedAt,omitempty"` // Actual completion time (may differ from scheduled time if done late)
	Notes         string    `json:"notes,omitempty"` // Optional notes for this occurrence
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
