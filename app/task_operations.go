package app

import (
	"fmt"
	"sort"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/SongRunqi/go-todo/internal/validator"
)

func RestoreTask(todos *[]TodoItem, backupTodos *[]TodoItem, id int, store *FileTodoStore) error {
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	// Find the task in backup
	var taskToRestore *TodoItem
	var backupIndex int = -1
	for i := 0; i < len(*backupTodos); i++ {
		if (*backupTodos)[i].TaskID == id {
			taskToRestore = &(*backupTodos)[i]
			backupIndex = i
			break
		}
	}

	if taskToRestore == nil {
		return fmt.Errorf("task with ID %d not found in backup", id)
	}

	logger.Debugf("Found task to restore - ID %d: %s", id, taskToRestore.TaskName)

	// Change status back to pending
	restoredTask := *taskToRestore
	restoredTask.Status = "pending"

	// Add to active todos
	*todos = append(*todos, restoredTask)

	// Save updated active todos
	err := store.Save(*todos, false)
	if err != nil {
		return fmt.Errorf("failed to save active todos: %w", err)
	}

	// Remove from backup
	newBackupTodos := make([]TodoItem, 0)
	for i := 0; i < len(*backupTodos); i++ {
		if i != backupIndex {
			newBackupTodos = append(newBackupTodos, (*backupTodos)[i])
		}
	}
	*backupTodos = newBackupTodos

	// Save updated backup
	err = store.Save(*backupTodos, true)
	if err != nil {
		return fmt.Errorf("failed to update backup: %w", err)
	}

	logger.Debug("Task restored successfully")
	output.PrintTaskRestored(id, restoredTask.TaskName)
	return nil
}

func CopyCompletedTasks(todos *[]TodoItem, store *FileTodoStore, weekOnly bool) error {
	// Collect completed tasks from both main list and backup
	completedTasks := make([]TodoItem, 0)

	// Get completed tasks from main list
	for _, task := range *todos {
		completedTasks = append(completedTasks, task)
	}

	// Get completed tasks from backup
	backupTodos, err := store.Load(true)
	if err != nil {
		logger.Warnf("Failed to load backup todos: %v", err)
	} else {
		for _, task := range backupTodos {
			if task.Status == "completed" {
				completedTasks = append(completedTasks, task)
			}
		}
	}

	if len(completedTasks) == 0 {
		fmt.Println("No completed tasks found")
		return nil
	}

	// Group tasks by week
	tasksByWeek := make(map[string][]string)
	now := time.Now()

	for _, task := range completedTasks {
		// Use EndTime to determine the week
		year, week := task.EndTime.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		// If weekOnly is true, only include current week
		if weekOnly {
			currentYear, currentWeek := now.ISOWeek()
			if year != currentYear || week != currentWeek {
				continue
			}
		}

		if _, exists := tasksByWeek[weekKey]; !exists {
			tasksByWeek[weekKey] = make([]string, 0)
		}
		tasksByWeek[weekKey] = append(tasksByWeek[weekKey], task.TaskName)
	}

	if len(tasksByWeek) == 0 {
		fmt.Println("No completed tasks found for the specified time period")
		return nil
	}

	// Sort weeks
	weeks := make([]string, 0, len(tasksByWeek))
	for week := range tasksByWeek {
		weeks = append(weeks, week)
	}
	sort.Strings(weeks)

	// Format output
	output := ""
	for _, week := range weeks {
		output += fmt.Sprintf("=== %s ===\n", week)
		for i, taskName := range tasksByWeek[week] {
			output += fmt.Sprintf("%d. %s\n", i+1, taskName)
		}
		output += "\n"
	}

	// Print to stdout (can be piped to clipboard tools like pbcopy or xclip)
	fmt.Print(output)

	logger.Infof("Copied %d completed tasks from %d week(s)", len(completedTasks), len(weeks))
	return nil
}
