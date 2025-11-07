package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileTodoStore_Save_And_Load(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test_todo.json")
	backupPath := filepath.Join(tmpDir, "test_backup.json")

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: backupPath,
	}

	// Create test todos
	testTodos := []TodoItem{
		{
			TaskID:     1,
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
			User:       "testuser",
			TaskName:   "Test Task 1",
			TaskDesc:   "This is a test task",
			Status:     "pending",
			DueDate:    "2025-11-06",
			Urgent:     "high",
		},
		{
			TaskID:     2,
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(48 * time.Hour),
			User:       "testuser",
			TaskName:   "Test Task 2",
			TaskDesc:   "Another test task",
			Status:     "pending",
			DueDate:    "2025-11-07",
			Urgent:     "medium",
		},
	}

	// Test Save
	err := store.Save(&testTodos, false)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		t.Fatalf("Todo file was not created")
	}

	// Test Load
	loadedTodos, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded data
	if len(loadedTodos) != len(testTodos) {
		t.Fatalf("Expected %d todos, got %d", len(testTodos), len(loadedTodos))
	}

	for i := range testTodos {
		if loadedTodos[i].TaskID != testTodos[i].TaskID {
			t.Errorf("Todo %d: expected TaskID %d, got %d", i, testTodos[i].TaskID, loadedTodos[i].TaskID)
		}
		if loadedTodos[i].TaskName != testTodos[i].TaskName {
			t.Errorf("Todo %d: expected TaskName %s, got %s", i, testTodos[i].TaskName, loadedTodos[i].TaskName)
		}
		if loadedTodos[i].Status != testTodos[i].Status {
			t.Errorf("Todo %d: expected Status %s, got %s", i, testTodos[i].Status, loadedTodos[i].Status)
		}
	}
}

func TestFileTodoStore_Save_Backup(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test_todo.json")
	backupPath := filepath.Join(tmpDir, "test_backup.json")

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: backupPath,
	}

	testTodos := []TodoItem{
		{
			TaskID:   1,
			TaskName: "Completed Task",
			Status:   "completed",
		},
	}

	// Save to backup
	err := store.Save(&testTodos, true)
	if err != nil {
		t.Fatalf("Save to backup failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file was not created")
	}

	// Load from backup
	loadedBackup, err := store.Load(true)
	if err != nil {
		t.Fatalf("Load from backup failed: %v", err)
	}

	if len(loadedBackup) != 1 {
		t.Fatalf("Expected 1 backup todo, got %d", len(loadedBackup))
	}

	if loadedBackup[0].TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", loadedBackup[0].TaskID)
	}
}

func TestFileTodoStore_Load_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "nonexistent.json")

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: todoPath + ".bak",
	}

	// Loading a non-existent file should create the file and return empty list
	todos, err := store.Load(false)
	if err != nil {
		t.Fatalf("Expected no error when loading non-existent file, got: %v", err)
	}

	// Should return empty list
	if len(todos) != 0 {
		t.Errorf("Expected 0 todos from non-existent file, got %d", len(todos))
	}

	// File should be created
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		t.Fatal("Expected file to be created, but it doesn't exist")
	}

	// Verify the created file contains empty JSON array
	data, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	var parsed []TodoItem
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Created file does not contain valid JSON: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("Expected created file to contain empty array, got %d items", len(parsed))
	}
}

func TestFileTodoStore_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(todoPath, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: todoPath + ".bak",
	}

	// Loading invalid JSON should return an error
	_, err = store.Load(false)
	if err == nil {
		t.Fatal("Expected error when loading invalid JSON, got nil")
	}
}

func TestFileTodoStore_Save_InvalidPath(t *testing.T) {
	// Use an invalid path (directory doesn't exist and can't be created)
	store := &FileTodoStore{
		Path:       "/invalid/path/that/does/not/exist/todo.json",
		BackupPath: "/invalid/path/backup.json",
	}

	testTodos := []TodoItem{
		{TaskID: 1, TaskName: "Test"},
	}

	err := store.Save(&testTodos, false)
	if err == nil {
		t.Fatal("Expected error when saving to invalid path, got nil")
	}
}

func TestFileTodoStore_Load_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "empty.json")

	// Create empty JSON array file
	err := os.WriteFile(todoPath, []byte("[]"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: todoPath + ".bak",
	}

	todos, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Expected 0 todos from empty file, got %d", len(todos))
	}
}

func TestFileTodoStore_SaveLoad_PreservesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test_todo.json")

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: todoPath + ".bak",
	}

	createTime := time.Date(2025, 11, 5, 10, 30, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 6, 18, 0, 0, 0, time.UTC)

	testTodo := TodoItem{
		TaskID:     42,
		CreateTime: createTime,
		EndTime:    endTime,
		User:       "alice",
		TaskName:   "Important Task",
		TaskDesc:   "This is very important",
		Status:     "pending",
		DueDate:    "2025-11-06",
		Urgent:     "urgent",
	}

	todos := []TodoItem{testTodo}
	err := store.Save(&todos, false)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(loaded))
	}

	// Verify all fields are preserved
	l := loaded[0]
	if l.TaskID != testTodo.TaskID {
		t.Errorf("TaskID: expected %d, got %d", testTodo.TaskID, l.TaskID)
	}
	if l.User != testTodo.User {
		t.Errorf("User: expected %s, got %s", testTodo.User, l.User)
	}
	if l.TaskName != testTodo.TaskName {
		t.Errorf("TaskName: expected %s, got %s", testTodo.TaskName, l.TaskName)
	}
	if l.TaskDesc != testTodo.TaskDesc {
		t.Errorf("TaskDesc: expected %s, got %s", testTodo.TaskDesc, l.TaskDesc)
	}
	if l.Status != testTodo.Status {
		t.Errorf("Status: expected %s, got %s", testTodo.Status, l.Status)
	}
	if l.DueDate != testTodo.DueDate {
		t.Errorf("DueDate: expected %s, got %s", testTodo.DueDate, l.DueDate)
	}
	if l.Urgent != testTodo.Urgent {
		t.Errorf("Urgent: expected %s, got %s", testTodo.Urgent, l.Urgent)
	}
	// Time comparison with truncation to handle precision differences
	if !l.CreateTime.Truncate(time.Second).Equal(testTodo.CreateTime.Truncate(time.Second)) {
		t.Errorf("CreateTime: expected %v, got %v", testTodo.CreateTime, l.CreateTime)
	}
	if !l.EndTime.Truncate(time.Second).Equal(testTodo.EndTime.Truncate(time.Second)) {
		t.Errorf("EndTime: expected %v, got %v", testTodo.EndTime, l.EndTime)
	}
}

func TestFileTodoStore_SaveLoad_MultipleItems(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test_todo.json")

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: todoPath + ".bak",
	}

	// Create 100 test items
	testTodos := make([]TodoItem, 100)
	for i := 0; i < 100; i++ {
		testTodos[i] = TodoItem{
			TaskID:   i + 1,
			TaskName: "Task " + string(rune(i)),
			Status:   "pending",
		}
	}

	err := store.Save(&testTodos, false)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != 100 {
		t.Fatalf("Expected 100 todos, got %d", len(loaded))
	}

	// Verify order is preserved
	for i := 0; i < 100; i++ {
		if loaded[i].TaskID != i+1 {
			t.Errorf("Item %d: expected TaskID %d, got %d", i, i+1, loaded[i].TaskID)
		}
	}
}

func TestFileTodoStore_JSON_Format(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test_todo.json")

	store := &FileTodoStore{
		Path:       todoPath,
		BackupPath: todoPath + ".bak",
	}

	testTodos := []TodoItem{
		{TaskID: 1, TaskName: "Test Task", Status: "pending"},
	}

	err := store.Save(&testTodos, false)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read the file and verify it's valid JSON
	data, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsed []TodoItem
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("File does not contain valid JSON: %v", err)
	}

	// Verify it's pretty-printed (indented)
	if len(data) < 50 {
		t.Error("JSON appears to not be indented (too short)")
	}
}
