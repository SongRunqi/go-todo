package domain

import "time"

// TodoItem represents a task item in the todo list
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
	Tags       []string  `json:"tags,omitempty"` // Tags associated with this task

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
	ScheduledTime time.Time `json:"scheduledTime"`         // The scheduled time for this occurrence
	Status        string    `json:"status"`                // pending, completed, missed, skipped
	CompletedAt   time.Time `json:"completedAt,omitempty"` // Actual completion time (may differ from scheduled time if done late)
	Notes         string    `json:"notes,omitempty"`       // Optional notes for this occurrence
}

// TodoStore defines the interface for todo storage operations
type TodoStore interface {
	Load(backup bool) ([]TodoItem, error)
	Save(todoItems []TodoItem, backup bool) error
}
