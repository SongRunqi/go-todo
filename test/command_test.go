package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/SongRunqi/go-todo/app"
)

func TestCreateTask(t *testing.T) {
	todos := []app.TodoItem{}

	newTask := &app.TodoItem{
		TaskName: "New Task",
		TaskDesc: "Task description",
		User:     "testuser",
	}

	err := app.CreateTask(&todos, newTask)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if len(todos) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(todos))
	}

	if todos[0].TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", todos[0].TaskID)
	}

	if todos[0].Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", todos[0].Status)
	}

	if todos[0].TaskName != "New Task" {
		t.Errorf("Expected TaskName 'New Task', got '%s'", todos[0].TaskName)
	}
}

func TestCreateTask_MultipleItems(t *testing.T) {
	todos := []app.TodoItem{}

	for i := 1; i <= 5; i++ {
		task := &app.TodoItem{
			TaskName: "Task " + string(rune(i)),
		}
		err := app.CreateTask(&todos, task)
		if err != nil {
			t.Fatalf("CreateTask failed at iteration %d: %v", i, err)
		}
	}

	if len(todos) != 5 {
		t.Fatalf("Expected 5 tasks, got %d", len(todos))
	}

	// Verify IDs are sequential
	for i := 0; i < 5; i++ {
		if todos[i].TaskID != i+1 {
			t.Errorf("Expected TaskID %d, got %d", i+1, todos[i].TaskID)
		}
	}
}

func TestGetLastId(t *testing.T) {
	tests := []struct {
		name     string
		todos    []app.TodoItem
		expected int
	}{
		{
			name:     "empty list",
			todos:    []app.TodoItem{},
			expected: 1,
		},
		{
			name: "single item",
			todos: []app.TodoItem{
				{TaskID: 1},
			},
			expected: 2,
		},
		{
			name: "sequential IDs",
			todos: []app.TodoItem{
				{TaskID: 1},
				{TaskID: 2},
				{TaskID: 3},
			},
			expected: 4,
		},
		{
			name: "non-sequential IDs",
			todos: []app.TodoItem{
				{TaskID: 5},
				{TaskID: 2},
				{TaskID: 8},
			},
			expected: 9,
		},
		{
			name: "with gaps",
			todos: []app.TodoItem{
				{TaskID: 1},
				{TaskID: 5},
				{TaskID: 3},
			},
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.GetLastId(&tt.todos)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	// Initialize with empty backup
	emptyBackup := []app.TodoItem{}
	err := store.Save(&emptyBackup, true)
	if err != nil {
		t.Fatalf("Failed to initialize backup: %v", err)
	}

	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1", Status: "pending"},
		{TaskID: 2, TaskName: "Task 2", Status: "pending"},
		{TaskID: 3, TaskName: "Task 3", Status: "pending"},
	}

	// Complete task 2
	err = app.Complete(&todos, &app.TodoItem{TaskID: 2}, store)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	// Verify task was removed from active list
	if len(todos) != 2 {
		t.Fatalf("Expected 2 tasks remaining, got %d", len(todos))
	}

	// Verify the right task was removed
	for _, todo := range todos {
		if todo.TaskID == 2 {
			t.Error("Task 2 should have been removed from active list")
		}
	}

	// Verify task was added to backup
	backup, err := store.Load(true)
	if err != nil {
		t.Fatalf("Failed to load backup: %v", err)
	}

	if len(backup) != 1 {
		t.Fatalf("Expected 1 task in backup, got %d", len(backup))
	}

	if backup[0].TaskID != 2 {
		t.Errorf("Expected backup to contain task 2, got task %d", backup[0].TaskID)
	}

	if backup[0].Status != "completed" {
		t.Errorf("Expected backup task status 'completed', got '%s'", backup[0].Status)
	}
}

func TestComplete_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
	}

	tests := []struct {
		name   string
		taskID int
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"non-existent ID", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.Complete(&todos, &app.TodoItem{TaskID: tt.taskID}, store)
			if err == nil {
				t.Error("Expected error for invalid task ID, got nil")
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
		{TaskID: 2, TaskName: "Task 2"},
		{TaskID: 3, TaskName: "Task 3"},
	}

	err := app.DeleteTask(&todos, 2, store)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify task was removed
	if len(todos) != 3 { // Note: DeleteTask doesn't modify the slice in place
		// Let's reload from store
		todos, err = store.Load(false)
		if err != nil {
			t.Fatalf("Failed to reload: %v", err)
		}
	}

	// The saved list should have 2 items
	todos, err = store.Load(false)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(todos) != 2 {
		t.Fatalf("Expected 2 tasks after deletion, got %d", len(todos))
	}

	// Verify the right task was deleted
	for _, todo := range todos {
		if todo.TaskID == 2 {
			t.Error("Task 2 should have been deleted")
		}
	}
}

func TestDeleteBackupTask(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	backupTodos := []app.TodoItem{
		{TaskID: 1, TaskName: "Archived Task 1"},
		{TaskID: 2, TaskName: "Archived Task 2"},
	}
	if err := store.Save(backupTodos, true); err != nil {
		t.Fatalf("Failed to seed backup: %v", err)
	}

	if err := app.DeleteBackupTask(1, store); err != nil {
		t.Fatalf("DeleteBackupTask failed: %v", err)
	}

	loaded, err := store.Load(true)
	if err != nil {
		t.Fatalf("Failed to load backup after deletion: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("Expected 1 task left in backup, got %d", len(loaded))
	}
	if loaded[0].TaskID != 2 {
		t.Errorf("Expected remaining TaskID 2, got %d", loaded[0].TaskID)
	}
}

func TestDeleteTask_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
	}

	tests := []struct {
		name   string
		taskID int
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"non-existent ID", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.DeleteTask(&todos, tt.taskID, store)
			if err == nil {
				t.Error("Expected error for invalid task ID, got nil")
			}
		})
	}
}

func TestRestoreTask(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Active Task", Status: "pending"},
	}

	backupTodos := []app.TodoItem{
		{TaskID: 2, TaskName: "Completed Task", Status: "completed"},
		{TaskID: 3, TaskName: "Another Completed", Status: "completed"},
	}

	err := app.RestoreTask(&todos, &backupTodos, 2, store)
	if err != nil {
		t.Fatalf("RestoreTask failed: %v", err)
	}

	// Verify task was added to active list
	if len(todos) != 2 {
		t.Fatalf("Expected 2 active tasks, got %d", len(todos))
	}

	// Verify the restored task has pending status
	found := false
	for _, todo := range todos {
		if todo.TaskID == 2 {
			found = true
			if todo.Status != "pending" {
				t.Errorf("Expected restored task to have 'pending' status, got '%s'", todo.Status)
			}
		}
	}

	if !found {
		t.Error("Task 2 was not found in active list after restore")
	}

	// Verify task was removed from backup
	if len(backupTodos) != 1 {
		t.Fatalf("Expected 1 task in backup after restore, got %d", len(backupTodos))
	}

	if backupTodos[0].TaskID == 2 {
		t.Error("Task 2 should have been removed from backup")
	}
}

func TestRestoreTask_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{}
	backupTodos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
	}

	tests := []struct {
		name   string
		taskID int
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"non-existent ID", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.RestoreTask(&todos, &backupTodos, tt.taskID, store)
			if err == nil {
				t.Error("Expected error for invalid task ID, got nil")
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	todos := []app.TodoItem{
		{
			TaskID:     1,
			TaskName:   "Test Task",
			TaskDesc:   "Description",
			Status:     "pending",
			User:       "testuser",
			DueDate:    "2025-11-06",
			Urgent:     "high",
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
		},
	}

	err := app.GetTask(&todos, 1)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
}

func TestGetTask_InvalidID(t *testing.T) {
	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
	}

	tests := []struct {
		name   string
		taskID int
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"non-existent ID", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.GetTask(&todos, tt.taskID)
			if err == nil {
				t.Error("Expected error for invalid task ID, got nil")
			}
		})
	}
}

func TestList(t *testing.T) {
	todos := []app.TodoItem{
		{
			TaskID:   1,
			TaskName: "Task 1",
			TaskDesc: "Description 1",
			Status:   "pending",
			Urgent:   "high",
			EndTime:  time.Now().Add(24 * time.Hour),
		},
		{
			TaskID:   2,
			TaskName: "Task 2",
			TaskDesc: "Description 2",
			Status:   "completed",
			Urgent:   "low",
			EndTime:  time.Now().Add(48 * time.Hour),
		},
	}

	err := app.List(&todos)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestList_EmptyList(t *testing.T) {
	todos := []app.TodoItem{}

	err := app.List(&todos)
	if err != nil {
		t.Fatalf("List with empty todos failed: %v", err)
	}
}

func TestDoI_CreateIntent(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{}

	intentJSON := `{
		"intent": "create",
		"tasks": [
			{
				"taskId": -1,
				"user": "testuser",
				"createTime": "2025-11-05T10:00:00Z",
				"endTime": "2025-11-06T18:00:00Z",
				"taskName": "Test Task",
				"taskDesc": "Test Description",
				"dueDate": "2025-11-06",
				"urgent": "high"
			}
		]
	}`

	err := app.DoI(intentJSON, &todos, store)
	if err != nil {
		t.Fatalf("DoI create intent failed: %v", err)
	}

	if len(todos) != 1 {
		t.Fatalf("Expected 1 task after create, got %d", len(todos))
	}

	if todos[0].TaskName != "Test Task" {
		t.Errorf("Expected TaskName 'Test Task', got '%s'", todos[0].TaskName)
	}
}

func TestDoI_ListIntent(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{
			TaskID:   1,
			TaskName: "Task 1",
			Status:   "pending",
			EndTime:  time.Now().Add(24 * time.Hour),
		},
	}

	intentJSON := `{
		"intent": "list"
	}`

	err := app.DoI(intentJSON, &todos, store)
	if err != nil {
		t.Fatalf("DoI list intent failed: %v", err)
	}
}

func TestDoI_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{}

	invalidJSON := `{invalid json`

	err := app.DoI(invalidJSON, &todos, store)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestDoI_UnknownIntent(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{}

	intentJSON := `{
		"intent": "unknown_intent"
	}`

	err := app.DoI(intentJSON, &todos, store)
	if err == nil {
		t.Error("Expected error for unknown intent, got nil")
	}
}

func TestUpdateTask(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{
			TaskID:     1,
			TaskName:   "Original Task",
			TaskDesc:   "Original Description",
			Status:     "pending",
			User:       "user1",
			DueDate:    "2025-11-06",
			Urgent:     "medium",
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
		},
	}

	// Test JSON update
	updateJSON := `{
		"taskId": 1,
		"taskName": "Updated Task",
		"taskDesc": "Updated Description",
		"status": "pending",
		"user": "user1",
		"dueDate": "2025-11-07",
		"urgent": "high"
	}`

	err := app.UpdateTask(&todos, updateJSON, store)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if todos[0].TaskName != "Updated Task" {
		t.Errorf("Expected TaskName 'Updated Task', got '%s'", todos[0].TaskName)
	}

	if todos[0].TaskDesc != "Updated Description" {
		t.Errorf("Expected TaskDesc 'Updated Description', got '%s'", todos[0].TaskDesc)
	}

	if todos[0].Urgent != "high" {
		t.Errorf("Expected Urgent 'high', got '%s'", todos[0].Urgent)
	}
}

func TestUpdateTask_Markdown(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{
			TaskID:     1,
			TaskName:   "Original Task",
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
		},
	}

	markdown := `# Updated Task Name

- **Task ID:** 1
- **Task Name:** Updated Task Name
- **Status:** pending
- **User:** alice
- **Due Date:** 2025-11-07
- **Urgency:** urgent

## Description

This is the updated description.`

	err := app.UpdateTask(&todos, markdown, store)
	if err != nil {
		t.Fatalf("UpdateTask with markdown failed: %v", err)
	}

	if todos[0].TaskName != "Updated Task Name" {
		t.Errorf("Expected TaskName 'Updated Task Name', got '%s'", todos[0].TaskName)
	}

	if todos[0].User != "alice" {
		t.Errorf("Expected User 'alice', got '%s'", todos[0].User)
	}

	if todos[0].Urgent != "urgent" {
		t.Errorf("Expected Urgent 'urgent', got '%s'", todos[0].Urgent)
	}
}

func TestUpdateTask_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	store := &app.FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []app.TodoItem{
		{TaskID: 1, TaskName: "Task 1"},
	}

	updateJSON := `{
		"taskId": 999,
		"taskName": "Non-existent"
	}`

	err := app.UpdateTask(&todos, updateJSON, store)
	if err == nil {
		t.Error("Expected error for non-existent task ID, got nil")
	}
}

func TestIntentResponseUnmarshal(t *testing.T) {
	jsonStr := `{
		"intent": "create",
		"tasks": [
			{
				"taskId": 1,
				"taskName": "Test",
				"status": "pending"
			}
		]
	}`

	var intent app.IntentResponse
	err := json.Unmarshal([]byte(jsonStr), &intent)
	if err != nil {
		t.Fatalf("Failed to unmarshal IntentResponse: %v", err)
	}

	if intent.Intent != "create" {
		t.Errorf("Expected intent 'create', got '%s'", intent.Intent)
	}

	if len(intent.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(intent.Tasks))
	}

	if intent.Tasks[0].TaskName != "Test" {
		t.Errorf("Expected task name 'Test', got '%s'", intent.Tasks[0].TaskName)
	}
}
