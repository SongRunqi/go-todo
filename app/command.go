package app

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/SongRunqi/go-todo/parser"
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
		log.Println("error parsing intent response:", err)
		return fmt.Errorf("failed to parse intent response: %w", err)
	}

	log.Println("Intent:", intentResponse.Intent)
	log.Println("Number of tasks:", len(intentResponse.Tasks))

	switch intentResponse.Intent {
	case "create":
		// Handle multiple tasks separated by semicolons
		for i := range intentResponse.Tasks {
			task := &intentResponse.Tasks[i]
			if err := CreateTask(todos, task); err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}
			fmt.Printf("Task created: %s\n", task.TaskName)
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
		log.Println("Unknown intent:", intentResponse.Intent)
		return fmt.Errorf("unknown intent: %s", intentResponse.Intent)
	}
	return nil
}

func Complete(todos *[]TodoItem, todo *TodoItem, store *FileTodoStore) error {
	id := todo.TaskID
	if id <= 0 {
		return fmt.Errorf("invalid task ID: %d", id)
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			log.Println("[complete] task id is:", id, "name:", (*todos)[i].TaskName, "desc:", (*todos)[i].TaskDesc)

			// Set the task as completed
			completedTask := (*todos)[i]
			completedTask.Status = "completed"

			// Load existing backup todos
			backupTodos, err := store.Load(true)
			if err != nil {
				return fmt.Errorf("failed to load backup: %w", err)
			}

			// Add completed task to backup
			backupTodos = append(backupTodos, completedTask)

			// Save completed task to backup file
			err = store.Save(&backupTodos, true)
			if err != nil {
				return fmt.Errorf("failed to save to backup: %w", err)
			}

			// Remove completed task from original todos
			newTodos := make([]TodoItem, 0)
			for j := 0; j < len(*todos); j++ {
				if j != i { // Skip the completed task
					newTodos = append(newTodos, (*todos)[j])
				}
			}
			*todos = newTodos

			// Save updated todos (without completed task) to original file
			err = store.Save(todos, false)
			if err != nil {
				return fmt.Errorf("failed to save updated todos: %w", err)
			}

			log.Println("[complete] task moved to backup and removed from active todos")
			fmt.Printf("Task %d completed and archived successfully\n", id)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

func CreateTask(todos *[]TodoItem, todo *TodoItem) error {
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
	if id <= 0 {
		return fmt.Errorf("invalid task ID: %d", id)
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			task := &(*todos)[i]
			log.Println("[get] found task id:", id, "name:", task.TaskName)

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

- **Task ID:** %d
- **Task Name:** %s
- **Status:** %s
- **User:** %s
- **Due Date:** %s
- **Urgency:** %s%s%s

## Description

%s

---

**Tips:** To update this task, copy this markdown and modify the fields above.`,
				task.TaskName,
				task.TaskID,
				task.TaskName,
				task.Status,
				task.User,
				task.DueDate,
				task.Urgent,
				func() string {
					if createdTime != "" {
						return "\n- **Created:** " + createdTime
					}
					return ""
				}(),
				func() string {
					if endTime != "" {
						return "\n- **End Time:** " + endTime
					}
					return ""
				}(),
				task.TaskDesc)

			fmt.Println(md)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

func UpdateTask(todos *[]TodoItem, todoMD string, store *FileTodoStore) error {
	log.Println("[update] Input content:", todoMD)

	// Parse the input using the parser package
	parsedTask, err := parser.Parse(todoMD)
	if err != nil {
		return fmt.Errorf("failed to parse task update: %w", err)
	}

	// Convert parser.TodoItem to main.TodoItem
	updatedTask := TodoItem{
		TaskID:     parsedTask.TaskID,
		CreateTime: parsedTask.CreateTime,
		EndTime:    parsedTask.EndTime,
		User:       parsedTask.User,
		TaskName:   parsedTask.TaskName,
		TaskDesc:   parsedTask.TaskDesc,
		Status:     parsedTask.Status,
		DueDate:    parsedTask.DueDate,
		Urgent:     parsedTask.Urgent,
	}

	if updatedTask.TaskID <= 0 {
		return fmt.Errorf("invalid task ID: %d", updatedTask.TaskID)
	}

	// Find and update the task
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == updatedTask.TaskID {
			log.Println("[update] updating task id:", updatedTask.TaskID, "name:", updatedTask.TaskName)

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

			log.Println("[update] task updated and saved")
			fmt.Printf("Task %d updated successfully\n", updatedTask.TaskID)

			// Return the updated task as JSON
			data, err := json.MarshalIndent(&updatedTask, "", "  ")
			if err != nil {
				log.Println("[update] Failed to marshal updated task:", err)
			} else {
				fmt.Println(string(data))
			}
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", updatedTask.TaskID)
}

func DeleteTask(todos *[]TodoItem, id int, store *FileTodoStore) error {
	if id <= 0 {
		return fmt.Errorf("invalid task ID: %d", id)
	}

	found := false
	newTodos := make([]TodoItem, 0)
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			found = true
			continue
		}
		newTodos = append(newTodos, (*todos)[i])
	}

	if !found {
		return fmt.Errorf("task with ID %d not found", id)
	}

	err := store.Save(&newTodos, false)
	if err != nil {
		return fmt.Errorf("failed to save after deletion: %w", err)
	}

	fmt.Printf("Task %d deleted successfully\n", id)
	return nil
}

func RestoreTask(todos *[]TodoItem, backupTodos *[]TodoItem, id int, store *FileTodoStore) error {
	if id <= 0 {
		return fmt.Errorf("invalid task ID: %d", id)
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

	log.Println("[restore] found task id:", id, "name:", taskToRestore.TaskName)

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

	log.Println("[restore] task restored successfully")
	fmt.Printf("Task %d (%s) restored successfully\n", id, restoredTask.TaskName)
	return nil
}
