package app

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkCreateTask benchmarks task creation
func BenchmarkCreateTask(b *testing.B) {
	todos := make([]TodoItem, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &TodoItem{
			TaskName: fmt.Sprintf("Task %d", i),
			TaskDesc: "Benchmark test task",
			User:     "benchuser",
			Urgent:   "medium",
		}
		CreateTask(&todos, task)
	}
}

// BenchmarkGetLastId benchmarks ID generation
func BenchmarkGetLastId(b *testing.B) {
	// Prepare a list with many items
	todos := make([]TodoItem, 1000)
	for i := 0; i < 1000; i++ {
		todos[i] = TodoItem{TaskID: i + 1}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetLastId(&todos)
	}
}

// BenchmarkList benchmarks listing and formatting todos
func BenchmarkList(b *testing.B) {
	// Prepare test data
	now := time.Now()
	todos := []TodoItem{
		{
			TaskID:     1,
			TaskName:   "Task 1",
			TaskDesc:   "Description 1",
			Status:     "pending",
			CreateTime: now,
			DueDate:    now.Add(24 * time.Hour).Format("2006-01-02"),
			Urgent:     "high",
		},
		{
			TaskID:     2,
			TaskName:   "Task 2",
			TaskDesc:   "Description 2",
			Status:     "pending",
			CreateTime: now,
			DueDate:    now.Add(48 * time.Hour).Format("2006-01-02"),
			Urgent:     "medium",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		List(&todos)
	}
}

// BenchmarkSortedList benchmarks the sorting logic
func BenchmarkSortedList(b *testing.B) {
	now := time.Now()
	todos := []TodoItem{
		{TaskID: 3, TaskName: "Task 3", DueDate: now.Add(72 * time.Hour).Format("2006-01-02"), Urgent: "low"},
		{TaskID: 1, TaskName: "Task 1", DueDate: now.Add(24 * time.Hour).Format("2006-01-02"), Urgent: "high"},
		{TaskID: 2, TaskName: "Task 2", DueDate: now.Add(48 * time.Hour).Format("2006-01-02"), Urgent: "medium"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedList(&todos)
	}
}

// BenchmarkTransToAlfredItem benchmarks Alfred item transformation
func BenchmarkTransToAlfredItem(b *testing.B) {
	now := time.Now()
	todos := []TodoItem{
		{TaskID: 1, TaskName: "Task 1", TaskDesc: "Description 1", DueDate: now.Add(24 * time.Hour).Format("2006-01-02")},
		{TaskID: 2, TaskName: "Task 2", TaskDesc: "Description 2", DueDate: now.Add(48 * time.Hour).Format("2006-01-02")},
		{TaskID: 3, TaskName: "Task 3", TaskDesc: "Description 3", DueDate: now.Add(72 * time.Hour).Format("2006-01-02")},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TransToAlfredItem(&todos)
	}
}

// BenchmarkCompleteTask benchmarks task completion
func BenchmarkCompleteTask(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup
		tmpDir := b.TempDir()
		store := &FileTodoStore{
			Path:       tmpDir + "/todos.json",
			BackupPath: tmpDir + "/backup.json",
		}

		todos := []TodoItem{
			{TaskID: 1, TaskName: "Task 1", Status: "pending"},
			{TaskID: 2, TaskName: "Task 2", Status: "pending"},
		}
		store.Save(&todos, false)

		b.StartTimer()
		Complete(&todos, &TodoItem{TaskID: 1}, store)
	}
}

// BenchmarkDeleteTask benchmarks task deletion
func BenchmarkDeleteTask(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup
		tmpDir := b.TempDir()
		store := &FileTodoStore{
			Path: tmpDir + "/todos.json",
		}

		todos := []TodoItem{
			{TaskID: 1, TaskName: "Task 1"},
			{TaskID: 2, TaskName: "Task 2"},
			{TaskID: 3, TaskName: "Task 3"},
		}

		b.StartTimer()
		DeleteTask(&todos, 2, store)
	}
}

// BenchmarkGetTask benchmarks retrieving a single task
func BenchmarkGetTask(b *testing.B) {
	now := time.Now()
	todos := []TodoItem{
		{TaskID: 1, TaskName: "Task 1", TaskDesc: "Description", Status: "pending",
			CreateTime: now, EndTime: now.Add(time.Hour), User: "user", DueDate: "2025-11-06", Urgent: "high"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTask(&todos, 1)
	}
}

// BenchmarkUpdateTask benchmarks task updates
func BenchmarkUpdateTask(b *testing.B) {
	markdown := `# Task 1
- **Task ID:** 1
- **Task Name:** Updated Task
- **Status:** pending
- **User:** testuser
- **Due Date:** 2025-11-06
- **Urgency:** high

## Description
Updated description
`

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tmpDir := b.TempDir()
		store := &FileTodoStore{Path: tmpDir + "/todos.json"}
		todos := []TodoItem{
			{TaskID: 1, TaskName: "Task 1", Status: "pending", User: "user", DueDate: "2025-11-05", Urgent: "medium"},
		}
		store.Save(&todos, false)

		b.StartTimer()
		UpdateTask(&todos, markdown, store)
	}
}

// BenchmarkCreateMultipleTasks benchmarks creating multiple tasks
func BenchmarkCreateMultipleTasks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		todos := make([]TodoItem, 0, 10)
		for j := 0; j < 10; j++ {
			task := &TodoItem{
				TaskName: fmt.Sprintf("Task %d", j),
				TaskDesc: "Description",
				User:     "user",
			}
			CreateTask(&todos, task)
		}
	}
}
