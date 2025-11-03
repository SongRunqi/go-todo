package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
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

func DoI(todoStr string, todos *[]TodoItem, store *FileTodoStore) {

	var intentResponse IntentResponse
	removedata := removeJsonTag(todoStr)
	err := json.Unmarshal([]byte(removedata), &intentResponse)
	if err != nil {
		log.Println("error parsing intent response:", err)
		return
	}

	log.Println("Intent:", intentResponse.Intent)
	log.Println("Number of tasks:", len(intentResponse.Tasks))

	switch intentResponse.Intent {
	case "create":
		// Handle multiple tasks separated by semicolons
		for i := range intentResponse.Tasks {
			task := &intentResponse.Tasks[i]
			CreateTask(todos, task)
			fmt.Printf("Task created: %s\n", task.TaskName)
		}
		// Save all tasks at once after creating them
		err := store.Save(todos, false)
		if err != nil {
			log.Println("[create] Failed to save todos batch:", err)
		}
	case "list":
		List(todos)
	case "complete":
		// For complete and delete, we might need additional logic
		// to extract task ID from the user input or tasks array
		if len(intentResponse.Tasks) > 0 {
			if err := Complete(todos, &intentResponse.Tasks[0], store); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	case "delete":
		if len(intentResponse.Tasks) > 0 {
			if err := DeleteTask(todos, intentResponse.Tasks[0].TaskID, store); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	default:
		log.Println("Unknown intent:", intentResponse.Intent)
	}
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

func CreateTask(todos *[]TodoItem, todo *TodoItem) {
	// Generate a unique TaskID
	id := GetLastId(todos)
	// Set the Status field to "pending"
	todo.TaskID = id
	todo.Status = "pending"
	// Add the new todo to the todos slice (but don't save yet)
	*todos = append(*todos, *todo)
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

func List(todos *[]TodoItem) {
	newTodos := sortedList(todos)
	alfredItems := TransToAlfredItem(&newTodos)
	response := AlfredResponse{Items: *alfredItems}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println("[list] Failed to marshal todos:", err)
	}
	fmt.Println(string(data))
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
	var updatedTask TodoItem

	// Try to parse as markdown first, fall back to JSON
	log.Println("[update] Input content:", todoMD)
	if strings.Contains(todoMD, "Task ID:") {
		// Parse markdown format
		lines := strings.Split(todoMD, "\n")
		log.Println("[update] Processing markdown format with", len(lines), "lines")
		inDescription := false

		for _, line := range lines {
			line = strings.TrimSpace(line)
			log.Println("[update] Processing line:", line)
			if line == "" {
				continue
			}

			// Check if this is the compact format (all fields in one line)
			if strings.Contains(line, "Task ID:") && strings.Contains(line, "Status:") &&
				strings.Contains(line, "User:") && strings.Contains(line, "Due Date:") &&
				strings.Contains(line, "Urgency:") {
				log.Println("[update] Detected compact format, parsing all fields from one line")

				// Parse all fields from the compact line using a more robust approach
				// Split the line by spaces and process each field
				fields := strings.Fields(line)
				log.Println("[update] Compact format fields:", fields)
				for i := 0; i < len(fields); i++ {
					field := fields[i]

					if field == "Task" && i+2 < len(fields) && fields[i+1] == "ID:" {
						// Task ID: value
						idStr := strings.Trim(fields[i+2], "*")
						fmt.Sscanf(idStr, "%d", &updatedTask.TaskID)
						i += 2
					} else if field == "Status:" && i+1 < len(fields) {
						// Status: value
						updatedTask.Status = strings.Trim(fields[i+1], "*")
						i += 1
					} else if field == "User:" && i+1 < len(fields) {
						// User: value
						updatedTask.User = strings.Trim(fields[i+1], "*")
						i += 1
					} else if field == "Due" && i+2 < len(fields) && fields[i+1] == "Date:" {
						// Due Date: value
						updatedTask.DueDate = strings.Trim(fields[i+2], "*")
						i += 2
					} else if field == "Urgency:" && i+1 < len(fields) {
						// Urgency: value
						updatedTask.Urgent = strings.Trim(fields[i+1], "*")
						i += 1
					}
				}

				log.Println("[update] Compact format parsed - TaskID:", updatedTask.TaskID,
					"Status:", updatedTask.Status, "User:", updatedTask.User,
					"DueDate:", updatedTask.DueDate, "Urgent:", updatedTask.Urgent)
				continue
			}

			if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "##") {
				updatedTask.TaskName = strings.TrimSpace(line[2:])
				log.Println("[update] Parsed TaskName:", updatedTask.TaskName)
			} else if strings.Contains(line, "Task ID:") {
				// Extract Task ID from list format like "- **Task ID:** 13"
				parts := strings.Split(line, "Task ID:")
				if len(parts) > 1 {
					idStr := strings.TrimSpace(parts[1])
					// Remove any ** markdown formatting
					idStr = strings.Trim(idStr, "* ")
					idStr = strings.TrimSpace(idStr)
					fmt.Sscanf(idStr, "%d", &updatedTask.TaskID)
				}
			} else if strings.Contains(line, "Task Name:") {
				// Extract Task Name from list format like "- **Task Name:** Task Title"
				parts := strings.Split(line, "Task Name:")
				if len(parts) > 1 {
					taskNameStr := strings.TrimSpace(parts[1])
					// Remove any ** markdown formatting
					taskNameStr = strings.Trim(taskNameStr, "* ")
					updatedTask.TaskName = strings.TrimSpace(taskNameStr)
					log.Println("[update] Parsed TaskName:", updatedTask.TaskName)
				}
			} else if strings.Contains(line, "Status:") {
				// Extract status from list format like "- **Status:** pending"
				parts := strings.Split(line, "Status:")
				if len(parts) > 1 {
					statusStr := strings.TrimSpace(parts[1])
					// Remove any ** markdown formatting
					statusStr = strings.Trim(statusStr, "* ")
					updatedTask.Status = strings.TrimSpace(statusStr)
					log.Println("[update] Parsed Status:", updatedTask.Status)
				}
			} else if strings.Contains(line, "User:") {
				parts := strings.Split(line, "User:")
				if len(parts) > 1 {
					userStr := strings.TrimSpace(parts[1])
					// Remove any ** markdown formatting
					userStr = strings.Trim(userStr, "* ")
					updatedTask.User = strings.TrimSpace(userStr)
					log.Println("[update] Parsed User:", updatedTask.User)
				}
			} else if strings.Contains(line, "Due Date:") {
				parts := strings.Split(line, "Due Date:")
				if len(parts) > 1 {
					dueDateStr := strings.TrimSpace(parts[1])
					// Remove any ** markdown formatting
					dueDateStr = strings.Trim(dueDateStr, "* ")
					updatedTask.DueDate = strings.TrimSpace(dueDateStr)
				}
			} else if strings.Contains(line, "Urgency:") {
				parts := strings.Split(line, "Urgency:")
				if len(parts) > 1 {
					urgencyStr := strings.TrimSpace(parts[1])
					// Remove any ** markdown formatting
					urgencyStr = strings.Trim(urgencyStr, "* ")
					updatedTask.Urgent = strings.TrimSpace(urgencyStr)
				}
			} else if strings.Contains(line, "Created:") {
				// Parse created time
				parts := strings.Split(line, "Created:")
				if len(parts) > 1 {
					createdStr := strings.TrimSpace(parts[1])
					createdStr = strings.Trim(createdStr, "* ")
					// Try to parse the time
					if t, err := time.Parse("2006-01-02 15:04:05", createdStr); err == nil {
						updatedTask.CreateTime = t
						log.Println("[update] Parsed CreateTime:", updatedTask.CreateTime)
					}
				}
			} else if strings.Contains(line, "End Time:") {
				// Parse end time
				parts := strings.Split(line, "End Time:")
				if len(parts) > 1 {
					endTimeStr := strings.TrimSpace(parts[1])
					endTimeStr = strings.Trim(endTimeStr, "* ")
					// Try to parse the time
					if t, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
						updatedTask.EndTime = t
						log.Println("[update] Parsed EndTime:", updatedTask.EndTime)
					}
				}
			} else if strings.Contains(line, "## Description") || (strings.Contains(line, "Description") && !strings.Contains(line, "##")) {
				// Start description section (handle both ## Description and just Description)
				inDescription = true
				log.Println("[update] Starting description section")
				continue
			} else if line == "---" || strings.HasPrefix(line, "Tips:") {
				// Stop parsing completely at the separator or tips section
				break
			} else if inDescription {
				// This is part of the description
				if updatedTask.TaskDesc != "" {
					updatedTask.TaskDesc += "\n"
				}
				updatedTask.TaskDesc += line
				log.Println("[update] Added to description:", line)
			}
		}
	} else {
		// Fall back to JSON parsing
		err := json.Unmarshal([]byte(todoMD), &updatedTask)
		if err != nil {
			return fmt.Errorf("invalid format, expected markdown or JSON: %w", err)
		}
	}

	if updatedTask.TaskID <= 0 {
		return fmt.Errorf("invalid task ID: %d", updatedTask.TaskID)
	}

	// Find and update the task
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == updatedTask.TaskID {
			log.Println("[update] updating task id:", updatedTask.TaskID, "name:", updatedTask.TaskName)

			// Preserve CreateTime and EndTime from original task
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
