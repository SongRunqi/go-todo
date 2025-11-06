package validator

import (
	"fmt"
	"strings"

	"github.com/SongRunqi/go-todo/internal/i18n"
)

// ValidateTaskID validates that a task ID is valid (must be > 0)
func ValidateTaskID(id int) error {
	if id <= 0 {
		return fmt.Errorf(i18n.T("validation.task_id_invalid"), id)
	}
	return nil
}

// ValidateTaskName validates that a task name is not empty
func ValidateTaskName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf(i18n.T("validation.task_name_empty"))
	}
	if len(trimmed) > 200 {
		return fmt.Errorf(i18n.T("validation.task_name_too_long"), len(trimmed))
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
		return fmt.Errorf(i18n.T("validation.invalid_status"), status)
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
		return fmt.Errorf(i18n.T("validation.invalid_urgency"), urgent)
	}
	return nil
}

// ValidateDescription validates task description length
func ValidateDescription(desc string) error {
	if len(desc) > 5000 {
		return fmt.Errorf(i18n.T("validation.description_too_long"), len(desc))
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
		return fmt.Errorf(i18n.T("validation.user_name_too_long"), len(trimmed))
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
