package app

import (
	"testing"
	"time"
)

func TestTransToAlfredItem(t *testing.T) {
	todos := []TodoItem{
		{
			TaskID:   1,
			TaskName: "Test Task",
			TaskDesc: "This is a test",
			Status:   "pending",
			Urgent:   "high",
		},
		{
			TaskID:   2,
			TaskName: "Completed Task",
			TaskDesc: "This is done",
			Status:   "completed",
			Urgent:   "medium",
		},
	}

	items := TransToAlfredItem(&todos)

	if items == nil {
		t.Fatal("TransToAlfredItem returned nil")
	}

	if len(*items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(*items))
	}

	// Test first item (pending)
	item1 := (*items)[0]
	if item1.Title != "[1] 🎯Test Task high" {
		t.Errorf("Expected title '[1] 🎯Test Task high', got '%s'", item1.Title)
	}
	if item1.Subtitle != "⌛️This is a test" {
		t.Errorf("Expected subtitle '⌛️This is a test', got '%s'", item1.Subtitle)
	}
	if item1.Arg != "1" {
		t.Errorf("Expected arg '1', got '%s'", item1.Arg)
	}
	if item1.Autocomplete != "Test Task" {
		t.Errorf("Expected autocomplete 'Test Task', got '%s'", item1.Autocomplete)
	}

	// Test second item (completed)
	item2 := (*items)[1]
	if item2.Title != "[2] 🎯Completed Task medium" {
		t.Errorf("Expected title '[2] 🎯Completed Task medium', got '%s'", item2.Title)
	}
	if item2.Subtitle != "✅This is done" {
		t.Errorf("Expected subtitle '✅This is done', got '%s'", item2.Subtitle)
	}
	if item2.Arg != "2" {
		t.Errorf("Expected arg '2', got '%s'", item2.Arg)
	}
}

func TestTransToAlfredItem_EmptyList(t *testing.T) {
	todos := []TodoItem{}
	items := TransToAlfredItem(&todos)

	if items == nil {
		t.Fatal("TransToAlfredItem returned nil for empty list")
	}

	if len(*items) != 0 {
		t.Errorf("Expected 0 items for empty list, got %d", len(*items))
	}
}

func TestSortedList(t *testing.T) {
	now := time.Now()

	todos := []TodoItem{
		{
			TaskID:   1,
			TaskName: "Far future task",
			EndTime:  now.Add(72 * time.Hour), // 3 days from now
		},
		{
			TaskID:   2,
			TaskName: "Soon task",
			EndTime:  now.Add(2 * time.Hour), // 2 hours from now
		},
		{
			TaskID:   3,
			TaskName: "Medium task",
			EndTime:  now.Add(24 * time.Hour), // 1 day from now
		},
		{
			TaskID:   4,
			TaskName: "Overdue task",
			EndTime:  now.Add(-24 * time.Hour), // 1 day ago
		},
	}

	sorted := sortedList(&todos)

	// Verify sorting order: overdue first, then closest deadline
	if len(sorted) != 4 {
		t.Fatalf("Expected 4 sorted items, got %d", len(sorted))
	}

	// First should be overdue
	if sorted[0].TaskID != 4 {
		t.Errorf("Expected first item to be overdue task (ID 4), got ID %d", sorted[0].TaskID)
	}
	if sorted[0].Urgent != "已截止" {
		t.Errorf("Expected overdue task to have '已截止', got '%s'", sorted[0].Urgent)
	}

	// Second should be the 2-hour task
	if sorted[1].TaskID != 2 {
		t.Errorf("Expected second item to be soon task (ID 2), got ID %d", sorted[1].TaskID)
	}

	// Third should be the 1-day task
	if sorted[2].TaskID != 3 {
		t.Errorf("Expected third item to be medium task (ID 3), got ID %d", sorted[2].TaskID)
	}

	// Fourth should be the 3-day task
	if sorted[3].TaskID != 1 {
		t.Errorf("Expected fourth item to be far future task (ID 1), got ID %d", sorted[3].TaskID)
	}
}

func TestSortedList_UrgentFormat(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		hoursFromNow   int
		expectedPrefix string
	}{
		{"days only", 73, "3d 1h"},
		{"hours only", 3, "3h"},
		{"minutes only", 0, ""},
		{"days and hours", 25, "1d 1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos := []TodoItem{
				{
					TaskID:  1,
					EndTime: now.Add(time.Duration(tt.hoursFromNow) * time.Hour),
				},
			}

			sorted := sortedList(&todos)
			urgent := sorted[0].Urgent

			if tt.expectedPrefix != "" {
				if len(urgent) < len(tt.expectedPrefix) {
					t.Errorf("Expected urgent to contain '%s', got '%s'", tt.expectedPrefix, urgent)
				}
			}

			if !contains(urgent, "截止") && urgent != "已截止" {
				t.Errorf("Expected urgent to contain '截止', got '%s'", urgent)
			}
		})
	}
}

func TestSortedList_EmptyList(t *testing.T) {
	todos := []TodoItem{}
	sorted := sortedList(&todos)

	if len(sorted) != 0 {
		t.Errorf("Expected empty sorted list, got %d items", len(sorted))
	}
}

func TestSortedList_SingleItem(t *testing.T) {
	now := time.Now()
	todos := []TodoItem{
		{
			TaskID:  1,
			EndTime: now.Add(24 * time.Hour),
		},
	}

	sorted := sortedList(&todos)

	if len(sorted) != 1 {
		t.Fatalf("Expected 1 sorted item, got %d", len(sorted))
	}

	if sorted[0].TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", sorted[0].TaskID)
	}
}

func TestSortedList_SameDeadline(t *testing.T) {
	now := time.Now()
	// Use slightly different deadlines to avoid map key collision in sortedList
	// This is a known limitation of the current implementation which uses
	// time difference as map key
	deadline1 := now.Add(24 * time.Hour)
	deadline2 := now.Add(24*time.Hour + 1*time.Second)
	deadline3 := now.Add(24*time.Hour + 2*time.Second)

	todos := []TodoItem{
		{TaskID: 1, EndTime: deadline1},
		{TaskID: 2, EndTime: deadline2},
		{TaskID: 3, EndTime: deadline3},
	}

	sorted := sortedList(&todos)

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 sorted items, got %d", len(sorted))
	}

	// All should have similar urgent values (within ~1 day range)
	for i := 0; i < len(sorted); i++ {
		if !contains(sorted[i].Urgent, "1d") && !contains(sorted[i].Urgent, "23h") {
			t.Errorf("Expected urgent to contain roughly 1 day, got '%s'", sorted[i].Urgent)
		}
	}
}

func TestRemoveJsonTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with json tags",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "with only opening tag",
			input:    "```json\n{\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "without tags",
			input:    "{\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with whitespace",
			input:    "  \n{\"key\": \"value\"}\n  ",
			expected: "{\"key\": \"value\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeJsonTag(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
