package parser

import (
	"testing"
	"time"
)

func TestParseJSON(t *testing.T) {
	jsonStr := `{
		"taskId": 1,
		"taskName": "Test Task",
		"taskDesc": "Test Description",
		"status": "pending",
		"user": "testuser",
		"dueDate": "2025-11-06",
		"urgent": "high"
	}`

	task, err := ParseJSON(jsonStr)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if task.TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", task.TaskID)
	}
	if task.TaskName != "Test Task" {
		t.Errorf("Expected TaskName 'Test Task', got '%s'", task.TaskName)
	}
	if task.TaskDesc != "Test Description" {
		t.Errorf("Expected TaskDesc 'Test Description', got '%s'", task.TaskDesc)
	}
	if task.Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", task.Status)
	}
	if task.User != "testuser" {
		t.Errorf("Expected User 'testuser', got '%s'", task.User)
	}
	if task.DueDate != "2025-11-06" {
		t.Errorf("Expected DueDate '2025-11-06', got '%s'", task.DueDate)
	}
	if task.Urgent != "high" {
		t.Errorf("Expected Urgent 'high', got '%s'", task.Urgent)
	}
}

func TestParseJSON_Invalid(t *testing.T) {
	invalidJSON := `{invalid json`

	_, err := ParseJSON(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseMarkdown_ListFormat(t *testing.T) {
	markdown := `# Test Task Name

- **Task ID:** 1
- **Task Name:** Test Task Name
- **Status:** pending
- **User:** alice
- **Due Date:** 2025-11-07
- **Urgency:** urgent

## Description

This is the task description.
It can have multiple lines.`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if task.TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", task.TaskID)
	}
	if task.TaskName != "Test Task Name" {
		t.Errorf("Expected TaskName 'Test Task Name', got '%s'", task.TaskName)
	}
	if task.Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", task.Status)
	}
	if task.User != "alice" {
		t.Errorf("Expected User 'alice', got '%s'", task.User)
	}
	if task.DueDate != "2025-11-07" {
		t.Errorf("Expected DueDate '2025-11-07', got '%s'", task.DueDate)
	}
	if task.Urgent != "urgent" {
		t.Errorf("Expected Urgent 'urgent', got '%s'", task.Urgent)
	}

	expectedDesc := "This is the task description.\nIt can have multiple lines."
	if task.TaskDesc != expectedDesc {
		t.Errorf("Expected TaskDesc '%s', got '%s'", expectedDesc, task.TaskDesc)
	}
}

func TestParseMarkdown_WithTimestamps(t *testing.T) {
	markdown := `# Task With Timestamps

- **Task ID:** 42
- **Task Name:** Task With Timestamps
- **Status:** pending
- **User:** bob
- **Due Date:** 2025-11-10
- **Urgency:** medium
- **Created:** 2025-11-05 10:30:00
- **End Time:** 2025-11-10 18:00:00

## Description

Task description here.`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if task.TaskID != 42 {
		t.Errorf("Expected TaskID 42, got %d", task.TaskID)
	}

	expectedCreateTime := time.Date(2025, 11, 5, 10, 30, 0, 0, time.UTC)
	if !task.CreateTime.Equal(expectedCreateTime) {
		t.Errorf("Expected CreateTime %v, got %v", expectedCreateTime, task.CreateTime)
	}

	expectedEndTime := time.Date(2025, 11, 10, 18, 0, 0, 0, time.UTC)
	if !task.EndTime.Equal(expectedEndTime) {
		t.Errorf("Expected EndTime %v, got %v", expectedEndTime, task.EndTime)
	}
}

func TestParseMarkdown_StopsAtSeparator(t *testing.T) {
	markdown := `# Test Task

- **Task ID:** 1
- **Task Name:** Test Task
- **Status:** pending

## Description

This is included.

---

**Tips:** This should not be included.`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if task.TaskDesc != "This is included." {
		t.Errorf("Expected TaskDesc 'This is included.', got '%s'", task.TaskDesc)
	}
}

func TestParseMarkdown_NoTaskID(t *testing.T) {
	markdown := `# Test Task

- **Task Name:** Test Task
- **Status:** pending

## Description

Description here.`

	_, err := ParseMarkdown(markdown)
	if err == nil {
		t.Error("Expected error for missing Task ID, got nil")
	}
}

func TestParseMarkdown_CompactFormat(t *testing.T) {
	// This is a compact format: taskName: taskDesc
	markdown := `Buy groceries: Need to buy milk, bread, and eggs from the store`

	task, err := ParseMarkdown(markdown)
	// Note: This format doesn't include Task ID, so it will fail validation
	// This test verifies that the compact format is correctly parsed into TaskName and TaskDesc
	if err == nil {
		t.Error("Expected error for missing Task ID")
	}

	// Even though it fails validation, we can check if TaskName and TaskDesc were parsed
	if task.TaskName != "Buy groceries" {
		t.Errorf("Expected TaskName 'Buy groceries', got '%s'", task.TaskName)
	}
	if task.TaskDesc != "Need to buy milk, bread, and eggs from the store" {
		t.Errorf("Expected TaskDesc 'Need to buy milk, bread, and eggs from the store', got '%s'", task.TaskDesc)
	}
}

func TestParseMarkdown_CompactFormatWithTaskID(t *testing.T) {
	// Compact format combined with Task ID in markdown
	markdown := `# Task ID: 5

Review code: Check the pull request and provide feedback

Task ID: 5`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown compact format with ID failed: %v", err)
	}

	if task.TaskID != 5 {
		t.Errorf("Expected TaskID 5, got %d", task.TaskID)
	}
	if task.TaskName != "Review code" {
		t.Errorf("Expected TaskName 'Review code', got '%s'", task.TaskName)
	}
	if task.TaskDesc != "Check the pull request and provide feedback" {
		t.Errorf("Expected TaskDesc 'Check the pull request and provide feedback', got '%s'", task.TaskDesc)
	}
}

func TestParseMarkdown_MinimalFormat(t *testing.T) {
	markdown := `# Minimal Task

- **Task ID:** 99`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown minimal format failed: %v", err)
	}

	if task.TaskID != 99 {
		t.Errorf("Expected TaskID 99, got %d", task.TaskID)
	}
	if task.TaskName != "Minimal Task" {
		t.Errorf("Expected TaskName 'Minimal Task', got '%s'", task.TaskName)
	}
}

func TestParse_AutoDetectMarkdown(t *testing.T) {
	markdown := `# Auto Detect

- **Task ID:** 10
- **Task Name:** Auto Detect
- **Status:** pending`

	task, err := Parse(markdown)
	if err != nil {
		t.Fatalf("Parse auto-detect markdown failed: %v", err)
	}

	if task.TaskID != 10 {
		t.Errorf("Expected TaskID 10, got %d", task.TaskID)
	}
	if task.TaskName != "Auto Detect" {
		t.Errorf("Expected TaskName 'Auto Detect', got '%s'", task.TaskName)
	}
}

func TestParse_AutoDetectJSON(t *testing.T) {
	jsonStr := `{
		"taskId": 20,
		"taskName": "JSON Task",
		"status": "pending"
	}`

	task, err := Parse(jsonStr)
	if err != nil {
		t.Fatalf("Parse auto-detect JSON failed: %v", err)
	}

	if task.TaskID != 20 {
		t.Errorf("Expected TaskID 20, got %d", task.TaskID)
	}
	if task.TaskName != "JSON Task" {
		t.Errorf("Expected TaskName 'JSON Task', got '%s'", task.TaskName)
	}
}

func TestParse_Invalid(t *testing.T) {
	invalid := `This is neither valid markdown nor JSON`

	_, err := Parse(invalid)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestParseMarkdown_EmptyDescription(t *testing.T) {
	markdown := `# Task Without Description

- **Task ID:** 7
- **Task Name:** Task Without Description
- **Status:** pending

## Description`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if task.TaskDesc != "" {
		t.Errorf("Expected empty TaskDesc, got '%s'", task.TaskDesc)
	}
}

func TestParseMarkdown_MultipleHashSigns(t *testing.T) {
	markdown := `# Main Title

- **Task ID:** 8
- **Task Name:** Main Title

## Description

This is under ## heading.

### Subheading

This should also be included.`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	// The description should include content after ## Description
	if !contains(task.TaskDesc, "This is under ## heading") {
		t.Errorf("Expected description to contain '## heading', got '%s'", task.TaskDesc)
	}
}

func TestParseMarkdown_FieldsWithBoldMarkdown(t *testing.T) {
	markdown := `# Task

- **Task ID:** 15
- **Task Name:** **Bold Name**
- **Status:** **bold-status**
- **User:** **bold-user**`

	task, err := ParseMarkdown(markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	// Should strip the ** from values
	if task.TaskName != "Bold Name" {
		t.Errorf("Expected TaskName 'Bold Name', got '%s'", task.TaskName)
	}
	if task.Status != "bold-status" {
		t.Errorf("Expected Status 'bold-status', got '%s'", task.Status)
	}
	if task.User != "bold-user" {
		t.Errorf("Expected User 'bold-user', got '%s'", task.User)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
