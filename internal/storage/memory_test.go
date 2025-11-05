package storage

import (
	"testing"
	"time"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	if store == nil {
		t.Fatal("NewMemoryStore returned nil")
	}

	if store.data == nil {
		t.Error("store.data is nil")
	}
}

func TestMemoryStore_SaveAndLoad(t *testing.T) {
	store := NewMemoryStore()

	todos := []TodoItem{
		{
			TaskID:     1,
			TaskName:   "Test Task",
			Status:     "pending",
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
		},
	}

	// Save
	err := store.Save(todos, false)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(loaded))
	}

	if loaded[0].TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", loaded[0].TaskID)
	}

	if loaded[0].TaskName != "Test Task" {
		t.Errorf("Expected TaskName 'Test Task', got '%s'", loaded[0].TaskName)
	}
}

func TestMemoryStore_Backup(t *testing.T) {
	store := NewMemoryStore()

	activeTodos := []TodoItem{
		{TaskID: 1, TaskName: "Active"},
	}

	backupTodos := []TodoItem{
		{TaskID: 2, TaskName: "Completed", Status: "completed"},
	}

	// Save active
	if err := store.Save(activeTodos, false); err != nil {
		t.Fatalf("Save active failed: %v", err)
	}

	// Save backup
	if err := store.Save(backupTodos, true); err != nil {
		t.Fatalf("Save backup failed: %v", err)
	}

	// Load active
	loaded, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load active failed: %v", err)
	}

	if len(loaded) != 1 || loaded[0].TaskID != 1 {
		t.Error("Active todos not loaded correctly")
	}

	// Load backup
	loadedBackup, err := store.Load(true)
	if err != nil {
		t.Fatalf("Load backup failed: %v", err)
	}

	if len(loadedBackup) != 1 || loadedBackup[0].TaskID != 2 {
		t.Error("Backup todos not loaded correctly")
	}
}

func TestMemoryStore_LoadEmpty(t *testing.T) {
	store := NewMemoryStore()

	todos, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load empty failed: %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Expected 0 todos, got %d", len(todos))
	}
}

func TestMemoryStore_Clear(t *testing.T) {
	store := NewMemoryStore()

	// Add data
	todos := []TodoItem{{TaskID: 1, TaskName: "Test"}}
	store.Save(todos, false)

	// Clear
	store.Clear()

	// Verify cleared
	loaded, err := store.Load(false)
	if err != nil {
		t.Fatalf("Load after clear failed: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("Expected 0 todos after clear, got %d", len(loaded))
	}
}

func TestMemoryStore_Size(t *testing.T) {
	store := NewMemoryStore()

	// Initially empty
	if size := store.Size(false); size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}

	// Add 3 todos
	todos := []TodoItem{
		{TaskID: 1},
		{TaskID: 2},
		{TaskID: 3},
	}
	store.Save(todos, false)

	if size := store.Size(false); size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}
}

func TestMemoryStore_Isolation(t *testing.T) {
	store := NewMemoryStore()

	original := []TodoItem{
		{TaskID: 1, TaskName: "Original"},
	}

	// Save
	store.Save(original, false)

	// Modify original after save
	original[0].TaskName = "Modified"

	// Load and check it wasn't affected
	loaded, _ := store.Load(false)
	if loaded[0].TaskName != "Original" {
		t.Error("Store data was affected by external modification")
	}

	// Modify loaded data
	loaded[0].TaskName = "Modified Again"

	// Load again and check store wasn't affected
	loaded2, _ := store.Load(false)
	if loaded2[0].TaskName != "Original" {
		t.Error("Store data was affected by loaded data modification")
	}
}

func TestMemoryStore_JSONConversion(t *testing.T) {
	store := NewMemoryStore()

	jsonData := `[
		{
			"taskId": 1,
			"taskName": "Test Task",
			"status": "pending"
		}
	]`

	// Load from JSON
	err := store.FromJSON(jsonData, false)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Verify loaded
	todos, _ := store.Load(false)
	if len(todos) != 1 || todos[0].TaskID != 1 {
		t.Error("FromJSON didn't load data correctly")
	}

	// Convert to JSON
	output, err := store.ToJSON(false)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if output == "" {
		t.Error("ToJSON returned empty string")
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			todos := []TodoItem{{TaskID: i, TaskName: "Task"}}
			store.Save(todos, false)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = store.Load(false)
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Should not crash (race detector will catch issues)
	t.Log("Concurrent access test completed successfully")
}
