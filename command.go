package main

import (
	"encoding/json"
	"fmt"
	"log"
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
			"taskName": "create a nice task name",
			"taskDesc": "give a clear description",
			"dueDate": "give a clear due date",
			"urgent": "low, medium, high, urgent, select one, default is medium, calculate this by time left"
		}
	]
}

Note: Only include "tasks" array when intent is "create". For other intents, omit the tasks field or return empty array.

`

func DoI(todoStr string, todos *[]TodoItem) {

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
		err := fileTodoStore.Save(todos, false)
		if err != nil {
			log.Println("[create] Failed to save todos batch:", err)
		}
	case "list":
		List(todos)
	case "complete":
		// For complete and delete, we might need additional logic
		// to extract task ID from the user input or tasks array
		if len(intentResponse.Tasks) > 0 {
			Complete(todos, &intentResponse.Tasks[0])
		}
	case "delete":
		if len(intentResponse.Tasks) > 0 {
			DeleteTask(todos, intentResponse.Tasks[0].TaskID)
		}
	default:
		log.Println("Unknown intent:", intentResponse.Intent)
	}
}

func Complete(todos *[]TodoItem, todo *TodoItem) {
	id := todo.TaskID
	if id <= 0 {
		log.Println("[complete]id is invalid")
		return
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			log.Println("[complete] task id is:", id, "name:", (*todos)[i].TaskName, "desc:", (*todos)[i].TaskDesc)

			// Set the task as completed
			completedTask := (*todos)[i]
			completedTask.Status = "completed"

			// Load existing backup todos
			backupTodos := fileTodoStore.Load(true)

			// Add completed task to backup
			backupTodos = append(backupTodos, completedTask)

			// Save completed task to backup file
			err := fileTodoStore.Save(&backupTodos, true)
			if err != nil {
				log.Println("[complete] Failed to save to backup:", err)
				fmt.Printf("Failed to backup completed task: %v\n", err)
				return
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
			err = fileTodoStore.Save(todos, false)
			if err != nil {
				log.Println("[complete] Failed to save updated todos:", err)
				fmt.Printf("Failed to save updated todos: %v\n", err)
				return
			}

			log.Println("[complete] task moved to backup and removed from active todos")
			fmt.Printf("Task %d completed and archived successfully\n", id)
			return
		}
	}
	log.Println("[complete] task not found, id:", id)
	fmt.Printf("Task with ID %d not found\n", id)
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

func GetTask(todos *[]TodoItem, id int) {
	if id <= 0 {
		log.Println("[get]id is invalid")
		fmt.Println("Invalid task ID")
		return
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			task := &(*todos)[i]
			log.Println("[get] found task id:", id, "name:", task.TaskName)

			// Create a filtered version without createTime and taskId
			filteredTask := struct {
				EndTime  time.Time `json:"endTime"`
				User     string    `json:"user"`
				TaskName string    `json:"taskName"`
				TaskDesc string    `json:"taskDesc"`
				Status   string    `json:"status"`
				DueDate  string    `json:"dueDate"`
				Urgent   string    `json:"urgent"`
			}{
				EndTime:  task.EndTime,
				User:     task.User,
				TaskName: task.TaskName,
				TaskDesc: task.TaskDesc,
				Status:   task.Status,
				DueDate:  task.DueDate,
				Urgent:   task.Urgent,
			}

			data, err := json.MarshalIndent(filteredTask, "", "  ")
			if err != nil {
				log.Println("[get] Failed to marshal task:", err)
			}
			fmt.Println(string(data))
			return
		}
	}
	log.Println("[get] task not found, id:", id)
	fmt.Printf("Task with ID %d not found\n", id)
}

func UpdateTask(todos *[]TodoItem, todoJSON string) {
	var updatedTask TodoItem
	err := json.Unmarshal([]byte(todoJSON), &updatedTask)
	if err != nil {
		log.Println("[update] Failed to parse JSON:", err)
		fmt.Printf("Invalid JSON format: %v\n", err)
		return
	}

	if updatedTask.TaskID <= 0 {
		log.Println("[update] taskId is invalid")
		fmt.Println("Invalid task ID in JSON")
		return
	}

	// Find and update the task
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == updatedTask.TaskID {
			log.Println("[update] updating task id:", updatedTask.TaskID, "name:", updatedTask.TaskName)

			// Update the task in place
			(*todos)[i] = updatedTask

			// Save to file
			err := fileTodoStore.Save(todos, false)
			if err != nil {
				log.Println("[update] Failed to save updated task:", err)
				fmt.Printf("Failed to save task: %v\n", err)
				return
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
			return
		}
	}
	log.Println("[update] task not found, id:", updatedTask.TaskID)
	fmt.Printf("Task with ID %d not found\n", updatedTask.TaskID)
}

func DeleteTask(todos *[]TodoItem, id int) {
	newTodos := make([]TodoItem, 0)
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			continue
		}
		newTodos = append(newTodos, (*todos)[i])
	}
	fileTodoStore.Save(&newTodos, false)
}
