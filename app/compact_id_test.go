package app

import (
	"testing"
	"time"
)

// TestGetLastId_CompactScenario tests that GetLastId returns unique incrementing IDs
// when tasks are added sequentially, simulating the compact task ID generation
func TestGetLastId_CompactScenario(t *testing.T) {
	// Simulate the scenario in CompactTasks where we:
	// 1. Start with existing backup tasks
	// 2. Add summary tasks one by one with GetLastId

	backupTodos := []TodoItem{
		{TaskID: 5, TaskName: "Pending Task", Status: "pending"},
		{TaskID: 10, TaskName: "Another Task", Status: "pending"},
	}

	// Simulate adding first summary task
	id1 := GetLastId(&backupTodos)
	if id1 != 11 {
		t.Errorf("First summary task ID should be 11, got %d", id1)
	}

	summaryTask1 := TodoItem{
		TaskID:   id1,
		TaskName: "Summary Week 1",
		Status:   "completed",
		User:     "System",
	}
	backupTodos = append(backupTodos, summaryTask1)

	// Simulate adding second summary task
	id2 := GetLastId(&backupTodos)
	if id2 != 12 {
		t.Errorf("Second summary task ID should be 12, got %d", id2)
	}

	summaryTask2 := TodoItem{
		TaskID:   id2,
		TaskName: "Summary Week 2",
		Status:   "completed",
		User:     "System",
	}
	backupTodos = append(backupTodos, summaryTask2)

	// Simulate adding third summary task
	id3 := GetLastId(&backupTodos)
	if id3 != 13 {
		t.Errorf("Third summary task ID should be 13, got %d", id3)
	}

	// Verify all IDs are unique
	idSet := make(map[int]bool)
	for _, task := range backupTodos {
		if idSet[task.TaskID] {
			t.Errorf("Duplicate TaskID found: %d", task.TaskID)
		}
		idSet[task.TaskID] = true
	}

	// Verify no zero IDs
	for _, task := range backupTodos {
		if task.TaskID == 0 {
			t.Errorf("Found task with ID 0: %s", task.TaskName)
		}
	}
}

// TestGetLastId_EmptyList tests GetLastId with empty list
func TestGetLastId_EmptyList(t *testing.T) {
	todos := []TodoItem{}
	id := GetLastId(&todos)
	if id != 1 {
		t.Errorf("Expected ID 1 for empty list, got %d", id)
	}
}

// TestGetLastId_SingleTask tests GetLastId with one task
func TestGetLastId_SingleTask(t *testing.T) {
	todos := []TodoItem{
		{TaskID: 5, TaskName: "Task 5"},
	}
	id := GetLastId(&todos)
	if id != 6 {
		t.Errorf("Expected ID 6, got %d", id)
	}
}

// TestGetLastId_NonSequentialIDs tests GetLastId with non-sequential IDs
func TestGetLastId_NonSequentialIDs(t *testing.T) {
	todos := []TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
		{TaskID: 5, TaskName: "Task 5"},
		{TaskID: 3, TaskName: "Task 3"},
		{TaskID: 10, TaskName: "Task 10"},
	}
	id := GetLastId(&todos)
	// Should return max(1,5,3,10) + 1 = 11
	if id != 11 {
		t.Errorf("Expected ID 11, got %d", id)
	}
}

// TestCompactTaskIDGeneration_Integration simulates the compact workflow
func TestCompactTaskIDGeneration_Integration(t *testing.T) {
	// Simulate initial backup with tasks from different periods
	backupTodos := []TodoItem{
		{TaskID: 1, TaskName: "Task 1", Status: "completed", EndTime: time.Now().AddDate(0, 0, -14)},
		{TaskID: 2, TaskName: "Task 2", Status: "completed", EndTime: time.Now().AddDate(0, 0, -14)},
		{TaskID: 3, TaskName: "Task 3", Status: "completed", EndTime: time.Now().AddDate(0, 0, -7)},
		{TaskID: 4, TaskName: "Task 4", Status: "deleted", EndTime: time.Now().AddDate(0, 0, -7)},
		{TaskID: 5, TaskName: "Active Task", Status: "pending", EndTime: time.Now()},
	}

	// Simulate compact: remove completed/deleted tasks
	tasksToRemove := map[int]bool{0: true, 1: true, 2: true, 3: true} // indices of completed/deleted
	newBackupTodos := make([]TodoItem, 0)
	for i, task := range backupTodos {
		if !tasksToRemove[i] {
			newBackupTodos = append(newBackupTodos, task)
		}
	}

	// Should only have the pending task (ID 5)
	if len(newBackupTodos) != 1 {
		t.Errorf("Expected 1 task after removal, got %d", len(newBackupTodos))
	}

	// Simulate adding 2 summary tasks (for 2 weeks)
	for i := 0; i < 2; i++ {
		id := GetLastId(&newBackupTodos)
		summaryTask := TodoItem{
			TaskID:   id,
			TaskName: "Summary " + string(rune('A'+i)),
			Status:   "completed",
			User:     "System",
		}
		newBackupTodos = append(newBackupTodos, summaryTask)
	}

	// Verify final state
	if len(newBackupTodos) != 3 {
		t.Errorf("Expected 3 tasks (1 pending + 2 summaries), got %d", len(newBackupTodos))
	}

	// Verify IDs: should be 5, 6, 7
	expectedIDs := []int{5, 6, 7}
	for i, task := range newBackupTodos {
		if task.TaskID != expectedIDs[i] {
			t.Errorf("Task %d: expected ID %d, got %d", i, expectedIDs[i], task.TaskID)
		}
		if task.TaskID == 0 {
			t.Errorf("Task %d has ID 0: %s", i, task.TaskName)
		}
	}

	// Verify no duplicate IDs
	idSet := make(map[int]bool)
	for _, task := range newBackupTodos {
		if idSet[task.TaskID] {
			t.Errorf("Duplicate TaskID found: %d (TaskName: %s)", task.TaskID, task.TaskName)
		}
		idSet[task.TaskID] = true
	}
}
