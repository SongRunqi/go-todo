package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{
		{
			TaskID:   1,
			TaskName: "Task 1",
			Status:   "pending",
			EndTime:  time.Now().Add(24 * time.Hour),
		},
	}

	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "list"},
	}

	cmd := &ListCommand{}
	err := cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("ListCommand.Execute failed: %v", err)
	}
}

func TestBackCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	// Create backup file with test data
	backupTodos := []TodoItem{
		{
			TaskID:   1,
			TaskName: "Completed Task",
			Status:   "completed",
			EndTime:  time.Now().Add(24 * time.Hour),
		},
	}
	err := store.Save(&backupTodos, true)
	if err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "back"},
	}

	cmd := &BackCommand{}
	err = cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("BackCommand.Execute failed: %v", err)
	}
}

func TestBackGetCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	backupTodos := []TodoItem{
		{
			TaskID:     1,
			TaskName:   "Completed Task",
			Status:     "completed",
			User:       "testuser",
			DueDate:    "2025-11-06",
			Urgent:     "high",
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
		},
	}
	err := store.Save(&backupTodos, true)
	if err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "back get 1"},
	}

	cmd := &BackGetCommand{}
	err = cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("BackGetCommand.Execute failed: %v", err)
	}
}

func TestBackGetCommand_Execute_InvalidArgs(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "back get"},
	}

	cmd := &BackGetCommand{}
	err := cmd.Execute(ctx)
	if err == nil {
		t.Error("Expected error for missing task ID, got nil")
	}
}

func TestBackRestoreCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	backupTodos := []TodoItem{
		{
			TaskID:   1,
			TaskName: "Completed Task",
			Status:   "completed",
		},
	}
	err := store.Save(&backupTodos, true)
	if err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "back restore 1"},
	}

	cmd := &BackRestoreCommand{}
	err = cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("BackRestoreCommand.Execute failed: %v", err)
	}

	// Verify task was restored
	if len(todos) != 1 {
		t.Fatalf("Expected 1 task after restore, got %d", len(todos))
	}

	if todos[0].TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", todos[0].TaskID)
	}

	if todos[0].Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", todos[0].Status)
	}
}

func TestCompleteCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	// Initialize backup
	emptyBackup := []TodoItem{}
	err := store.Save(&emptyBackup, true)
	if err != nil {
		t.Fatalf("Failed to initialize backup: %v", err)
	}

	todos := []TodoItem{
		{TaskID: 1, TaskName: "Task to Complete", Status: "pending"},
	}

	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "complete 1"},
	}

	cmd := &CompleteCommand{}
	err = cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("CompleteCommand.Execute failed: %v", err)
	}

	// Verify task was removed from active list
	if len(todos) != 0 {
		t.Errorf("Expected 0 active tasks, got %d", len(todos))
	}
}

func TestCompleteCommand_Execute_InvalidArgs(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "complete"},
	}

	cmd := &CompleteCommand{}
	err := cmd.Execute(ctx)
	if err == nil {
		t.Error("Expected error for missing task ID, got nil")
	}
}

func TestDeleteCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{
		{TaskID: 1, TaskName: "Task to Delete"},
		{TaskID: 2, TaskName: "Task to Keep"},
	}

	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "delete 1"},
	}

	cmd := &DeleteCommand{}
	err := cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("DeleteCommand.Execute failed: %v", err)
	}

	// Reload to verify deletion was saved
	loaded, err := store.Load(false)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Expected 1 task after delete, got %d", len(loaded))
	}

	if loaded[0].TaskID != 2 {
		t.Errorf("Expected remaining task to have ID 2, got %d", loaded[0].TaskID)
	}
}

func TestGetCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{
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

	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "get 1"},
	}

	cmd := &GetCommand{}
	err := cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("GetCommand.Execute failed: %v", err)
	}
}

func TestGetCommand_Execute_InvalidArgs(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "get"},
	}

	cmd := &GetCommand{}
	err := cmd.Execute(ctx)
	if err == nil {
		t.Error("Expected error for missing task ID, got nil")
	}
}

func TestUpdateCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{
		{
			TaskID:     1,
			TaskName:   "Original Task",
			CreateTime: time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
		},
	}

	updateJSON := `{"taskId": 1, "taskName": "Updated Task", "status": "pending"}`
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "update " + updateJSON},
	}

	cmd := &UpdateCommand{}
	err := cmd.Execute(ctx)
	if err != nil {
		t.Fatalf("UpdateCommand.Execute failed: %v", err)
	}

	if todos[0].TaskName != "Updated Task" {
		t.Errorf("Expected TaskName 'Updated Task', got '%s'", todos[0].TaskName)
	}
}

func TestUpdateCommand_Execute_InvalidArgs(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{}
	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "update"},
	}

	cmd := &UpdateCommand{}
	err := cmd.Execute(ctx)
	if err == nil {
		t.Error("Expected error for missing update content, got nil")
	}
}

func TestRouter_Route_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{
		{TaskID: 1, TaskName: "Task 1", EndTime: time.Now().Add(24 * time.Hour)},
	}

	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo", "list"},
	}

	router := NewRouter()
	err := router.Route(ctx)
	if err != nil {
		t.Fatalf("Router.Route failed for 'list': %v", err)
	}
}

func TestRouter_Route_NoCommand(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{}

	ctx := &Context{
		Store: store,
		Todos: &todos,
		Args:  []string{"todo"}, // No command provided
	}

	router := NewRouter()
	err := router.Route(ctx)
	if err == nil {
		t.Error("Expected error for missing command, got nil")
	}
}

func TestRouter_Register(t *testing.T) {
	router := NewRouter()

	// Verify some commands are registered
	expectedCommands := []string{"list", "ls", "back", "complete", "delete", "get", "update"}

	for _, cmdName := range expectedCommands {
		if _, ok := router.commands[cmdName]; !ok {
			t.Errorf("Expected command '%s' to be registered", cmdName)
		}
	}
}

func TestAICommand_Execute_WithMockedAPI(t *testing.T) {
	// This test would require mocking the Chat function
	// For now, we'll skip it as it requires external API
	t.Skip("Skipping AI command test - requires API mocking")
}

func TestContext_Structure(t *testing.T) {
	tmpDir := t.TempDir()
	store := &FileTodoStore{
		Path:       filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
	}

	todos := []TodoItem{}
	config := &Config{
		TodoPath:   filepath.Join(tmpDir, "todo.json"),
		BackupPath: filepath.Join(tmpDir, "backup.json"),
		APIKey:     "test-key",
		Model:      "test-model",
	}

	currentTime := time.Now()

	ctx := &Context{
		Store:       store,
		Todos:       &todos,
		Args:        []string{"todo", "list"},
		CurrentTime: currentTime,
		Config:      config,
	}

	// Verify context fields are set correctly
	if ctx.Store == nil {
		t.Error("Context.Store should not be nil")
	}
	if ctx.Todos == nil {
		t.Error("Context.Todos should not be nil")
	}
	if len(ctx.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(ctx.Args))
	}
	if ctx.Config == nil {
		t.Error("Context.Config should not be nil")
	}
	if ctx.CurrentTime.IsZero() {
		t.Error("Context.CurrentTime should not be zero")
	}
}

func TestConfig_LoadConfig(t *testing.T) {
	// Set environment variables
	os.Setenv("API_KEY", "test-api-key")
	os.Setenv("model", "test-model")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("model")

	config := LoadConfig()

	if config.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey 'test-api-key', got '%s'", config.APIKey)
	}

	if config.Model != "test-model" {
		t.Errorf("Expected Model 'test-model', got '%s'", config.Model)
	}

	// Verify default paths are set
	if config.TodoPath == "" {
		t.Error("TodoPath should not be empty")
	}

	if config.BackupPath == "" {
		t.Error("BackupPath should not be empty")
	}
}
