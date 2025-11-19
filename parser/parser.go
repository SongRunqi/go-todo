package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/domain"
)

// TodoItem is an alias to domain.TodoItem for backward compatibility
type TodoItem = domain.TodoItem

// ParseMarkdown parses a markdown-formatted string into a TodoItem
// Supports both list format and compact format
func ParseMarkdown(content string) (TodoItem, error) {
	var task TodoItem
	lines := strings.Split(content, "\n")
	inDescription := false

	log.Println("[parser] Processing markdown format with", len(lines), "lines")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for compact format (all fields in one line)
		if isCompactFormat(line) {
			log.Println("[parser] Detected compact format")
			parseCompactFormat(line, &task)
			continue
		}

		// Parse different field types
		if parseTitle(line, &task) {
			continue
		}
		if parseTaskID(line, &task) {
			continue
		}
		if parseTaskName(line, &task) {
			continue
		}
		if parseStatus(line, &task) {
			continue
		}
		if parseUser(line, &task) {
			continue
		}
		if parseDueDate(line, &task) {
			continue
		}
		if parseUrgency(line, &task) {
			continue
		}
		if parseCreateTime(line, &task) {
			continue
		}
		if parseEndTime(line, &task) {
			continue
		}

		// Check for description section start
		if strings.Contains(line, "## Description") ||
		   (strings.Contains(line, "Description") && !strings.Contains(line, "##")) {
			inDescription = true
			log.Println("[parser] Starting description section")
			continue
		}

		// Stop at separator or tips
		if line == "---" || strings.HasPrefix(line, "Tips:") || strings.HasPrefix(line, "**Tips:") {
			break
		}

		// Collect description lines
		if inDescription {
			if task.TaskDesc != "" {
				task.TaskDesc += "\n"
			}
			task.TaskDesc += line
			log.Println("[parser] Added to description:", line)
		}
	}

	if task.TaskID <= 0 {
		return task, fmt.Errorf("failed to parse task ID from markdown")
	}

	return task, nil
}

// ParseJSON parses a JSON-formatted string into a TodoItem
func ParseJSON(content string) (TodoItem, error) {
	var task TodoItem
	err := json.Unmarshal([]byte(content), &task)
	if err != nil {
		return task, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return task, nil
}

// Parse attempts to parse content as either Markdown or JSON
func Parse(content string) (TodoItem, error) {
	// Try markdown first if it contains markdown indicators
	if strings.Contains(content, "Task ID:") || strings.HasPrefix(strings.TrimSpace(content), "#") {
		task, err := ParseMarkdown(content)
		if err == nil {
			return task, nil
		}
		log.Println("[parser] Markdown parsing failed, trying JSON:", err)
	}

	// Try JSON
	task, err := ParseJSON(content)
	if err != nil {
		return task, fmt.Errorf("invalid format, expected markdown or JSON: %w", err)
	}
	return task, nil
}

// Helper functions for parsing specific fields

func isCompactFormat(line string) bool {
	// Check for simple "taskName: taskDesc" format
	// Must have exactly one colon and no markdown indicators
	if !strings.Contains(line, ":") {
		return false
	}
	// Should not contain markdown list or heading indicators
	if strings.HasPrefix(strings.TrimSpace(line), "-") ||
	   strings.HasPrefix(strings.TrimSpace(line), "#") ||
	   strings.HasPrefix(strings.TrimSpace(line), "*") {
		return false
	}
	// Should not contain other field markers
	if strings.Contains(line, "Task ID:") ||
	   strings.Contains(line, "Status:") ||
	   strings.Contains(line, "User:") {
		return false
	}
	// Should have content before and after the colon
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		taskName := strings.TrimSpace(parts[0])
		taskDesc := strings.TrimSpace(parts[1])
		return taskName != "" && taskDesc != ""
	}
	return false
}

func parseCompactFormat(line string, task *TodoItem) {
	// Parse simple "taskName: taskDesc" format
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		task.TaskName = strings.TrimSpace(parts[0])
		task.TaskDesc = strings.TrimSpace(parts[1])
		log.Println("[parser] Compact format parsed - TaskName:", task.TaskName,
			"TaskDesc:", task.TaskDesc)
	}
}

func parseTitle(line string, task *TodoItem) bool {
	if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "##") {
		task.TaskName = strings.TrimSpace(line[2:])
		log.Println("[parser] Parsed TaskName from title:", task.TaskName)
		return true
	}
	return false
}

func parseTaskID(line string, task *TodoItem) bool {
	if !strings.Contains(line, "Task ID:") {
		return false
	}

	parts := strings.Split(line, "Task ID:")
	if len(parts) > 1 {
		idStr := strings.TrimSpace(parts[1])
		idStr = strings.Trim(idStr, "* ")
		idStr = strings.TrimSpace(idStr)
		fmt.Sscanf(idStr, "%d", &task.TaskID)
		log.Println("[parser] Parsed TaskID:", task.TaskID)
		return true
	}
	return false
}

func parseTaskName(line string, task *TodoItem) bool {
	if !strings.Contains(line, "Task Name:") {
		return false
	}

	parts := strings.Split(line, "Task Name:")
	if len(parts) > 1 {
		nameStr := strings.TrimSpace(parts[1])
		nameStr = strings.Trim(nameStr, "* ")
		task.TaskName = strings.TrimSpace(nameStr)
		log.Println("[parser] Parsed TaskName:", task.TaskName)
		return true
	}
	return false
}

func parseStatus(line string, task *TodoItem) bool {
	if !strings.Contains(line, "Status:") || strings.Contains(line, "Task ID:") {
		return false
	}

	parts := strings.Split(line, "Status:")
	if len(parts) > 1 {
		statusStr := strings.TrimSpace(parts[1])
		statusStr = strings.Trim(statusStr, "* ")
		task.Status = strings.TrimSpace(statusStr)
		log.Println("[parser] Parsed Status:", task.Status)
		return true
	}
	return false
}

func parseUser(line string, task *TodoItem) bool {
	if !strings.Contains(line, "User:") || strings.Contains(line, "Task ID:") {
		return false
	}

	parts := strings.Split(line, "User:")
	if len(parts) > 1 {
		userStr := strings.TrimSpace(parts[1])
		userStr = strings.Trim(userStr, "* ")
		task.User = strings.TrimSpace(userStr)
		log.Println("[parser] Parsed User:", task.User)
		return true
	}
	return false
}

func parseDueDate(line string, task *TodoItem) bool {
	if !strings.Contains(line, "Due Date:") || strings.Contains(line, "Task ID:") {
		return false
	}

	parts := strings.Split(line, "Due Date:")
	if len(parts) > 1 {
		dateStr := strings.TrimSpace(parts[1])
		dateStr = strings.Trim(dateStr, "* ")
		task.DueDate = strings.TrimSpace(dateStr)
		log.Println("[parser] Parsed DueDate:", task.DueDate)
		return true
	}
	return false
}

func parseUrgency(line string, task *TodoItem) bool {
	if !strings.Contains(line, "Urgency:") || strings.Contains(line, "Task ID:") {
		return false
	}

	parts := strings.Split(line, "Urgency:")
	if len(parts) > 1 {
		urgencyStr := strings.TrimSpace(parts[1])
		urgencyStr = strings.Trim(urgencyStr, "* ")
		task.Urgent = strings.TrimSpace(urgencyStr)
		log.Println("[parser] Parsed Urgency:", task.Urgent)
		return true
	}
	return false
}

func parseCreateTime(line string, task *TodoItem) bool {
	if !strings.Contains(line, "Created:") {
		return false
	}

	parts := strings.Split(line, "Created:")
	if len(parts) > 1 {
		timeStr := strings.TrimSpace(parts[1])
		timeStr = strings.Trim(timeStr, "* ")
		if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
			task.CreateTime = t
			log.Println("[parser] Parsed CreateTime:", task.CreateTime)
			return true
		}
	}
	return false
}

func parseEndTime(line string, task *TodoItem) bool {
	if !strings.Contains(line, "End Time:") {
		return false
	}

	parts := strings.Split(line, "End Time:")
	if len(parts) > 1 {
		timeStr := strings.TrimSpace(parts[1])
		timeStr = strings.Trim(timeStr, "* ")
		if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
			task.EndTime = t
			log.Println("[parser] Parsed EndTime:", task.EndTime)
			return true
		}
	}
	return false
}
