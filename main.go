package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type AlfredItem struct {
	UID          string          `json:"uid,omitempty"`
	Title        string          `json:"title"`
	Subtitle     string          `json:"subtitle,omitempty"`
	Arg          string          `json:"arg,omitempty"`
	Autocomplete string          `json:"autocomplete,omitempty"`
	Icon         *Icon           `json:"icon,omitempty"`
	Text         *AlfredItemText `json:"text"`
}

type AlfredItemText struct {
	Copy      string `json:"copy"`
	Largetype string `json:"largetype"`
}
type AlfredResponse struct {
	Items []AlfredItem `json:"items"`
}

// Icon represents the icon for an item
type Icon struct {
	Path string `json:"path"`
}
type TodoItem struct {
	TaskID     int       `json:"taskId"`
	CreateTime time.Time `json:"createTime"`
	EndTime    time.Time `json:"endTime"`
	User       string    `json:"user"`
	TaskName   string    `json:"taskName"`
	TaskDesc   string    `json:"taskDesc"`
	Status     string    `json:"status"`
	DueDate    string    `json:"dueDate"`
	Urgent     string    `json:"urgent"`
}

type IntentResponse struct {
	Intent string     `json:"intent"`
	Tasks  []TodoItem `json:"tasks,omitempty"`
}

const path = "/Users/yitiansong/data/sync/todo.json"
const backupPath = "/Users/yitiansong/data/sync/todo_back.json"

//const path = "/Users/yitiansong/data/code/todo-go/todo.json"

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
const p = `
You are a todo helper agent, your task is return a todo item depending on user Messages.I hope you return json, and the field is as follows:
{
	"user": "",  if not specified, return You
	"taskName": "",
	"taskDesc"
}
`

type Msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIReasoning struct {
	Effort string `json:"effort"`
}
type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []Msg  `json:"messages"`
}

type OpenAIChoices struct {
	Message Msg `json:"message"`
}
type OpenAIResponse struct {
	Choices []OpenAIChoices `json:"choices"`
}

func LoadTodos(path string) []TodoItem {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Println("[load] read file err:", err)
	}
	var loadingTodos []TodoItem = make([]TodoItem, 0)
	err = json.Unmarshal(bytes, &loadingTodos)
	if err != nil {
		log.Println("[load] parse err", err.Error())
	}
	return loadingTodos
}
func main() {
	args := os.Args
	log.Println("args lens:", len(args))
	if len(args) < 2 {
		log.Println("args must be exits")
	}

	// starting request api
	log.Println("this is a test")

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	weekday := now.Weekday().String()
	// todo item
	var todos []TodoItem = LoadTodos(path)
	loadedbytes, _ := json.Marshal(todos)
	loadedTodos := string(loadedbytes)
	var userinput string = args[1]
	log.Println("user input is:", userinput)
	if userinput == "list" || userinput == "ls" {
		List(&todos)
		return
	} else if userinput == "back" {
		backupTodos := LoadTodos(backupPath)
		List(&backupTodos)
		return
	} else if arg := strings.Split(userinput, " "); len(arg) > 1 {
		log.Println("arg is:", arg[0])
		if arg[0] == "complete" {
			id, err := strconv.Atoi(arg[1])
			log.Println("[pcomplete] id is:", id)
			if err != nil {
				log.Println("[pcomplete] error occurs:", err)
			}
			Complete(&todos, &TodoItem{TaskID: id})
			return
		} else if arg[0] == "delete" {
			id, err := strconv.Atoi(arg[1])
			log.Println("[pdelete] id is:", id)
			if err != nil {
				log.Println("[pdelete] error occurs:", err)
			}
			DeleteTask(&todos, id)
			return
		} else if arg[0] == "get" {
			id, err := strconv.Atoi(arg[1])
			log.Println("[pget] id is:", id)
			if err != nil {
				log.Println("[pget] error occurs:", err)
			}
			GetTask(&todos, id)
			return
		} else if arg[0] == "update" {
			if len(arg) < 2 {
				log.Println("[pupdate] missing todo item JSON")
				return
			}
			todoJSON := strings.Join(arg[1:], " ")
			log.Println("[pupdate] updating with JSON:", todoJSON)
			UpdateTask(&todos, todoJSON)
			return
		}
	}
	ctx := "current time is" + nowStr + " and today is " + weekday + args[1] + ", current todos: " + loadedTodos
	log.Println("ctx is:", string(ctx))
	req := OpenAIRequest{
		Model: "deepseek-chat",
		Messages: []Msg{
			{Role: "system", Content: cmd},
			{Role: "user", Content: ctx},
		},
	}
	warpIntend, err := Chat(req)
	if err != nil {
		log.Println("error occurs:", err)
		return
	}
	// unwarp
	DoI(warpIntend, &todos)

}

func Chat(req OpenAIRequest) (string, error) {
	// struct -> json
	b, _ := json.Marshal(req)

	// create a client
	client := &http.Client{}
	// create a  http request
	request, err := http.NewRequest(http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(b))
	if err != nil {
		log.Println("[command]error occured when create a request:", err)
		return "", err
	}

	api := os.Getenv("DEEPSEEK_API_KEY")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+api)
	request.Header.Set("Accept", "application/json")

	// do request
	res, err := client.Do(request)
	if err != nil {
		log.Println("[command]error occured when get a response:", err)
		return "", err
	}
	defer res.Body.Close()

	// handle response
	if res.StatusCode != http.StatusOK {
		log.Println("[command]API returned status:", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("[command]error reading response:", err)
		return "", err
	}

	log.Println("Raw response:", string(resBody))

	var openAiResponse = OpenAIResponse{}
	err = json.Unmarshal(resBody, &openAiResponse)
	if err != nil {
		log.Println("[command]error occured when parse a response:", err)
		return "", err
	}

	log.Println("response:", openAiResponse)

	// get the ai response
	msg := openAiResponse.Choices[0].Message.Content
	return msg, nil
}

func removeJsonTag(str string) string {
	s := strings.Replace(str, "```json", "", 1)
	s = strings.Replace(s, "```", "", 1)
	return strings.TrimSpace(s)

}
func DoI(todoStr string, todos *[]TodoItem) {
	// TODO(human): Implement the response parsing and task handling logic
	// 1. Parse the AI response into IntentResponse struct (use removeJsonTag first)
	// 2. Extract the intent and handle each case appropriately
	// 3. For "create" intent, iterate through tasks array and create each task
	// 4. For other intents, call the appropriate function

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
		err := SaveTodos(todos, path)
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
			backupTodos := LoadTodos(backupPath)

			// Add completed task to backup
			backupTodos = append(backupTodos, completedTask)

			// Save completed task to backup file
			err := SaveTodos(&backupTodos, backupPath)
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
			err = SaveTodos(todos, path)
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

func SaveTodos(todos *[]TodoItem, filePath string) error {
	data, err := json.MarshalIndent(*todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Println("[save] Successfully saved todos to file")
	return nil
}

func sortedList(todos *[]TodoItem) []TodoItem {
	score := make(map[int64]int)
	now := time.Now().Unix()
	// assign score with task id, the less score, the higher priority
	for i, v := range *todos {
		s := v.EndTime.Unix() - now
		score[s] = i
	}

	times := make([]int64, 0)
	for k := range maps.Keys(score) {
		times = append(times, k)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	var newTodos []TodoItem = make([]TodoItem, 0)
	for _, v := range times {
		if item := &(*todos)[score[v]]; v < 0 {
			item.Urgent = "已截止"
		} else {
			days := v / 86400
			hours := (v % 86400) / 3600
			minutes := (v % 3600) / 60
			seconds := v % 60
			tip := "还有"
			if days > 0 {
				tip = tip + strconv.FormatInt(days, 10) + "d "
			}
			if hours > 0 {
				tip = tip + strconv.FormatInt(hours, 10) + "h "
			}
			if minutes > 0 {
				tip = tip + strconv.FormatInt(minutes, 10) + "m "
			}
			if seconds > 0 {
				tip = tip + strconv.FormatInt(seconds, 10) + "s "
			}
			item.Urgent = tip + "截止"
		}
		newTodos = append(newTodos, (*todos)[score[v]])
	}
	return newTodos
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

func TransToAlfredItem(todos *[]TodoItem) *[]AlfredItem {
	var items = make([]AlfredItem, 0)
	for i := 0; i < len(*todos); i++ {
		item := AlfredItem{}
		item.Title = "🎯" + (*todos)[i].TaskName + " " + (*todos)[i].Urgent
		completed := (*todos)[i].Status == "completed"
		var prefix string = ""
		if completed {
			prefix = "✅"
		} else {
			prefix = "⌛️"
		}
		item.Subtitle = prefix + (*todos)[i].TaskDesc
		item.Arg = strconv.Itoa((*todos)[i].TaskID)
		item.Autocomplete = (*todos)[i].TaskName
		items = append(items, item)
	}
	return &items
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
			err := SaveTodos(todos, path)
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
	SaveTodos(&newTodos, path)
}
