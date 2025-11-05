package validator

import (
	"fmt"
	"strings"
)

// ValidateTaskID validates that a task ID is valid (must be > 0)
func ValidateTaskID(id int) error {
	if id <= 0 {
		return fmt.Errorf("task ID must be greater than 0, got: %d", id)
	}
	return nil
}

// ValidateTaskName validates that a task name is not empty
func ValidateTaskName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if len(trimmed) > 200 {
		return fmt.Errorf("task name too long (max 200 characters), got: %d", len(trimmed))
	}
	return nil
}

// ValidateStatus validates that status is one of the allowed values
func ValidateStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"completed": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status '%s', must be one of: pending, completed", status)
	}
	return nil
}

// ValidateUrgency validates that urgency is one of the allowed values
func ValidateUrgency(urgent string) error {
	validUrgencies := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}

	if urgent == "" {
		return nil // Urgency is optional
	}

	if !validUrgencies[urgent] {
		return fmt.Errorf("invalid urgency '%s', must be one of: low, medium, high, urgent", urgent)
	}
	return nil
}

// ValidateDescription validates task description length
func ValidateDescription(desc string) error {
	if len(desc) > 5000 {
		return fmt.Errorf("task description too long (max 5000 characters), got: %d", len(desc))
	}
	return nil
}

// ValidateUser validates user name
func ValidateUser(user string) error {
	if user == "" {
		return nil // User is optional
	}

	trimmed := strings.TrimSpace(user)
	if len(trimmed) > 100 {
		return fmt.Errorf("user name too long (max 100 characters), got: %d", len(trimmed))
	}
	return nil
}

// ValidateTodoItem validates all fields of a TodoItem
type TodoItem interface {
	GetTaskID() int
	GetTaskName() string
	GetTaskDesc() string
	GetStatus() string
	GetUrgent() string
	GetUser() string
}

// ValidateAll validates all common fields
func ValidateAll(taskID int, taskName, taskDesc, status, urgent, user string) error {
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}
	if err := ValidateTaskName(taskName); err != nil {
		return err
	}
	if err := ValidateDescription(taskDesc); err != nil {
		return err
	}
	if err := ValidateStatus(status); err != nil {
		return err
	}
	if err := ValidateUrgency(urgent); err != nil {
		return err
	}
	if err := ValidateUser(user); err != nil {
		return err
	}
	return nil
}
