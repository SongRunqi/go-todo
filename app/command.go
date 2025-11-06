package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/SongRunqi/go-todo/parser"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/validator"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

const cmd = `
<System>
You are a todo helper agent. Your task is to analyze user input and determine their intent along with any tasks they want to create.

Key behaviors:
1. Identify the user's primary intent from the <ability> tag options
2. If the user wants to create tasks, treat ';' as a separator for multiple tasks
3. Return intent as a separate, independent attribute
4. Return tasks array only when user wants to create tasks (intent="create")

<ability>
<item>
	<name>create</name>
	<desc>user wants to create one or more tasks</desc>
</item>
<item>
	<name>delete</name>
	<desc>user wants to delete a task</desc>
</item>
<item>
	<name>list</name>
	<desc>user wants to see all the todolist</desc>
</item>
<item>
	<name>complete</name>
	<desc>user wants to complete a task</desc>
</item>
</ability>

Return format (remove markdown code fence):
{
	"intent": "create|delete|list|complete",
	"tasks": [
		{
			"taskId": -1,
			"user": "if not mentioned, You is default",
			"createTime": "use current time",
			"endTime": "place end time based on the current time",
			"taskName": "Extract a clear, concise title from the user's input. Use key words from their message without adding creative interpretations.",
			"taskDesc": "Summarize the user's input directly and factually. Use the exact words and intent from the user's message. Do not add creative interpretations or assumptions. Keep it concise (1-2 sentences) and preserve the original meaning.",
			"dueDate": "give a clear due date",
			"urgent": "low, medium, high, urgent, select one, default is medium, calculate this by time left"
		}
	]
}

Note: Only include "tasks" array when intent is "create". For other intents, omit the tasks field or return empty array.

`

func DoI(todoStr string, todos *[]TodoItem, store *FileTodoStore) error {

	var intentResponse IntentResponse
	removedata := removeJsonTag(todoStr)
	err := json.Unmarshal([]byte(removedata), &intentResponse)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to parse intent response")
		return fmt.Errorf("failed to parse intent response: %w", err)
	}

	logger.Infof("Intent: %s, Number of tasks: %d", intentResponse.Intent, len(intentResponse.Tasks))

	switch intentResponse.Intent {
	case "create":
		// Handle multiple tasks separated by semicolons
		for i := range intentResponse.Tasks {
			task := &intentResponse.Tasks[i]
			if err := CreateTask(todos, task); err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}
			output.PrintTaskCreated(task.TaskID, task.TaskName)
		}
		// Save all tasks at once after creating them
		err := store.Save(todos, false)
		if err != nil {
			return fmt.Errorf("failed to save todos batch: %w", err)
		}
	case "list":
		if err := List(todos); err != nil {
			return fmt.Errorf("failed to list todos: %w", err)
		}
	case "complete":
		// For complete and delete, we might need additional logic
		// to extract task ID from the user input or tasks array
		if len(intentResponse.Tasks) > 0 {
			if err := Complete(todos, &intentResponse.Tasks[0], store); err != nil {
				return fmt.Errorf("failed to complete task: %w", err)
			}
		}
	case "delete":
		if len(intentResponse.Tasks) > 0 {
			if err := DeleteTask(todos, intentResponse.Tasks[0].TaskID, store); err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}
		}
	default:
		logger.Warnf("Unknown intent: %s", intentResponse.Intent)
		return fmt.Errorf("unknown intent: %s", intentResponse.Intent)
	}
	return nil
}

func Complete(todos *[]TodoItem, todo *TodoItem, store *FileTodoStore) error {
	id := todo.TaskID
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			taskName := (*todos)[i].TaskName
			logger.Debugf("Completing task ID %d: %s - %s", id, (*todos)[i].TaskName, (*todos)[i].TaskDesc)

			// Set the task as completed (keep it in the main list)
			(*todos)[i].Status = "completed"

			// Save updated todos to original file
			err := store.Save(todos, false)
			if err != nil {
				return fmt.Errorf("failed to save updated todos: %w", err)
			}

			logger.Debug("Task marked as completed")
			output.PrintTaskCompleted(id, taskName)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

func CreateTask(todos *[]TodoItem, todo *TodoItem) error {
	// Validate task fields
	if err := validator.ValidateTaskName(todo.TaskName); err != nil {
		return err
	}
	if err := validator.ValidateStatus("pending"); err != nil {
		return err
	}
	if todo.Urgent != "" {
		if err := validator.ValidateUrgency(todo.Urgent); err != nil {
			return err
		}
	}
	if todo.TaskDesc != "" {
		if err := validator.ValidateDescription(todo.TaskDesc); err != nil {
			return err
		}
	}
	if todo.User != "" {
		if err := validator.ValidateUser(todo.User); err != nil {
			return err
		}
	}

	// Generate a unique TaskID
	id := GetLastId(todos)
	// Set the Status field to "pending"
	todo.TaskID = id
	todo.Status = "pending"
	// Add the new todo to the todos slice (but don't save yet)
	*todos = append(*todos, *todo)
	return nil
}

func GetLastId(todos *[]TodoItem) int {
	todoList := *todos
	length := len(todoList)
	if length < 1 {
		return 1
	}

	// Find the maximum TaskID to ensure uniqueness
	maxID := 0
	for _, todo := range todoList {
		if todo.TaskID > maxID {
			maxID = todo.TaskID
		}
	}
	return maxID + 1
}

func List(todos *[]TodoItem) error {
	newTodos := sortedList(todos)
	alfredItems := TransToAlfredItem(&newTodos)
	response := AlfredResponse{Items: *alfredItems}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func GetTask(todos *[]TodoItem, id int) error {
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			task := &(*todos)[i]
			logger.Debugf("Found task ID %d: %s", id, task.TaskName)

			// Format task as markdown
			// Only show Created and End Time if they have valid values
			createdTime := ""
			if !task.CreateTime.IsZero() {
				createdTime = task.CreateTime.Format("2006-01-02 15:04:05")
			}
			endTime := ""
			if !task.EndTime.IsZero() {
				endTime = task.EndTime.Format("2006-01-02 15:04:05")
			}

			md := fmt.Sprintf(`# %s

- **%s:** %d
- **%s:** %s
- **%s:** %s
- **%s:** %s
- **%s:** %s
- **%s:** %s%s%s

## %s

%s

---

**%s:** %s`,
				task.TaskName,
				i18n.T("field.task_id"), task.TaskID,
				i18n.T("field.task_name"), task.TaskName,
				i18n.T("field.status"), task.Status,
				i18n.T("field.user"), task.User,
				i18n.T("field.due_date"), task.DueDate,
				i18n.T("field.urgency"), task.Urgent,
				func() string {
					if createdTime != "" {
						return "\n- **" + i18n.T("field.created") + ":** " + createdTime
					}
					return ""
				}(),
				func() string {
					if endTime != "" {
						return "\n- **" + i18n.T("field.end_time") + ":** " + endTime
					}
					return ""
				}(),
				i18n.T("field.description"),
				task.TaskDesc,
				i18n.T("field.tips"), i18n.T("tip.edit_markdown"))

			fmt.Println(md)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

func UpdateTask(todos *[]TodoItem, todoMD string, store *FileTodoStore) error {
	logger.Debugf("Updating task with content: %s", todoMD)

	// Parse the input using the parser package
	parsedTask, err := parser.Parse(todoMD)
	if err != nil {
		return fmt.Errorf("failed to parse task update: %w", err)
	}

	// Normalize status first to handle Chinese/English variations
	normalizedStatus := validator.NormalizeStatus(parsedTask.Status)

	// Convert parser.TodoItem to main.TodoItem
	updatedTask := TodoItem{
		TaskID:     parsedTask.TaskID,
		CreateTime: parsedTask.CreateTime,
		EndTime:    parsedTask.EndTime,
		User:       parsedTask.User,
		TaskName:   parsedTask.TaskName,
		TaskDesc:   parsedTask.TaskDesc,
		Status:     normalizedStatus,
		DueDate:    parsedTask.DueDate,
		Urgent:     parsedTask.Urgent,
	}

	// Validate task ID
	if err := validator.ValidateTaskID(updatedTask.TaskID); err != nil {
		return err
	}

	// Validate other fields
	if err := validator.ValidateTaskName(updatedTask.TaskName); err != nil {
		return err
	}
	if err := validator.ValidateStatus(updatedTask.Status); err != nil {
		return err
	}
	if updatedTask.Urgent != "" {
		if err := validator.ValidateUrgency(updatedTask.Urgent); err != nil {
			return err
		}
	}
	if updatedTask.TaskDesc != "" {
		if err := validator.ValidateDescription(updatedTask.TaskDesc); err != nil {
			return err
		}
	}
	if updatedTask.User != "" {
		if err := validator.ValidateUser(updatedTask.User); err != nil {
			return err
		}
	}

	// Find and update the task
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == updatedTask.TaskID {
			logger.Debugf("Updating task ID %d: %s", updatedTask.TaskID, updatedTask.TaskName)

			// Preserve CreateTime and EndTime from original task if not provided
			if updatedTask.CreateTime.IsZero() {
				updatedTask.CreateTime = (*todos)[i].CreateTime
			}
			if updatedTask.EndTime.IsZero() {
				updatedTask.EndTime = (*todos)[i].EndTime
			}

			// Update the task in place
			(*todos)[i] = updatedTask

			// Save to file
			err := store.Save(todos, false)
			if err != nil {
				return fmt.Errorf("failed to save task: %w", err)
			}

			logger.Debug("Task updated and saved successfully")
			output.PrintTaskUpdated(updatedTask.TaskID, updatedTask.TaskName)

			// Return the updated task as JSON
			data, err := json.MarshalIndent(&updatedTask, "", "  ")
			if err != nil {
				logger.ErrorWithErr(err, "Failed to marshal updated task")
			} else {
				fmt.Println(string(data))
			}
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", updatedTask.TaskID)
}

func DeleteTask(todos *[]TodoItem, id int, store *FileTodoStore) error {
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	var deletedTask *TodoItem
	taskIndex := -1
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			deletedTask = &(*todos)[i]
			taskIndex = i
			break
		}
	}

	if deletedTask == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	taskName := deletedTask.TaskName
	logger.Debugf("Deleting task ID %d: %s", id, taskName)

	// Mark task as deleted
	deletedTask.Status = "deleted"

	// Load existing backup todos
	backupTodos, err := store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup: %w", err)
	}

	// Add deleted task to backup
	backupTodos = append(backupTodos, *deletedTask)

	// Save deleted task to backup file
	err = store.Save(&backupTodos, true)
	if err != nil {
		return fmt.Errorf("failed to save to backup: %w", err)
	}

	// Remove task from main todos
	newTodos := make([]TodoItem, 0)
	for i := 0; i < len(*todos); i++ {
		if i != taskIndex {
			newTodos = append(newTodos, (*todos)[i])
		}
	}
	*todos = newTodos

	// Save updated todos
	err = store.Save(todos, false)
	if err != nil {
		return fmt.Errorf("failed to save after deletion: %w", err)
	}

	logger.Debug("Task moved to backup with 'deleted' status")
	output.PrintTaskDeleted(id)
	return nil
}

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
	err := store.Save(todos, false)
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
	err = store.Save(backupTodos, true)
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
		if task.Status == "completed" {
			completedTasks = append(completedTasks, task)
		}
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

func CompactTasks(store *FileTodoStore, period string) error {
	// Validate period
	if period != "week" && period != "month" {
		return fmt.Errorf("invalid period: %s (must be 'week' or 'month')", period)
	}

	// Load backup tasks
	backupTodos, err := store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup: %w", err)
	}

	// Filter completed and deleted tasks
	completedTasks := make([]TodoItem, 0)
	deletedTasks := make([]TodoItem, 0)

	for _, task := range backupTodos {
		if task.Status == "completed" {
			completedTasks = append(completedTasks, task)
		} else if task.Status == "deleted" {
			deletedTasks = append(deletedTasks, task)
		}
	}

	if len(completedTasks) == 0 && len(deletedTasks) == 0 {
		fmt.Println("No completed or deleted tasks found in backup")
		return nil
	}

	// Group tasks by period
	type PeriodStats struct {
		Completed []string
		Deleted   []string
	}

	tasksByPeriod := make(map[string]*PeriodStats)

	// Process completed tasks
	for _, task := range completedTasks {
		periodKey := getPeriodKey(task.EndTime, period)
		if _, exists := tasksByPeriod[periodKey]; !exists {
			tasksByPeriod[periodKey] = &PeriodStats{
				Completed: make([]string, 0),
				Deleted:   make([]string, 0),
			}
		}
		tasksByPeriod[periodKey].Completed = append(tasksByPeriod[periodKey].Completed, task.TaskName)
	}

	// Process deleted tasks
	for _, task := range deletedTasks {
		periodKey := getPeriodKey(task.EndTime, period)
		if _, exists := tasksByPeriod[periodKey]; !exists {
			tasksByPeriod[periodKey] = &PeriodStats{
				Completed: make([]string, 0),
				Deleted:   make([]string, 0),
			}
		}
		tasksByPeriod[periodKey].Deleted = append(tasksByPeriod[periodKey].Deleted, task.TaskName)
	}

	// Sort periods
	periods := make([]string, 0, len(tasksByPeriod))
	for p := range tasksByPeriod {
		periods = append(periods, p)
	}
	sort.Strings(periods)

	// Format and display summary
	fmt.Println("==============================================")
	fmt.Printf("Task Summary (by %s)\n", period)
	fmt.Println("==============================================")

	totalCompleted := 0
	totalDeleted := 0

	for _, p := range periods {
		stats := tasksByPeriod[p]
		completedCount := len(stats.Completed)
		deletedCount := len(stats.Deleted)

		totalCompleted += completedCount
		totalDeleted += deletedCount

		fmt.Printf("📅 %s\n", p)
		fmt.Printf("   ✅ Completed: %d tasks\n", completedCount)
		if completedCount > 0 {
			for i, taskName := range stats.Completed {
				fmt.Printf("      %d. %s\n", i+1, taskName)
			}
		}
		fmt.Printf("   🗑️  Deleted: %d tasks\n", deletedCount)
		if deletedCount > 0 {
			for i, taskName := range stats.Deleted {
				fmt.Printf("      %d. %s\n", i+1, taskName)
			}
		}
		fmt.Println()
	}

	fmt.Println("==============================================")
	fmt.Printf("Total: %d completed, %d deleted (%d periods)\n", totalCompleted, totalDeleted, len(periods))
	fmt.Println("==============================================")

	logger.Infof("Compacted %d tasks from %d %ss", totalCompleted+totalDeleted, len(periods), period)
	return nil
}

func getPeriodKey(t time.Time, period string) string {
	if period == "week" {
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	} else { // month
		return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
	}
}
