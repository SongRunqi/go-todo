package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AlfredItem struct {
	UID          string `json:"uid,omitempty"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle,omitempty"`
	Arg          string `json:"arg,omitempty"`
	Autocomplete string `json:"autocomplete,omitempty"`
	Icon         *Icon  `json:"icon,omitempty"`
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
	Intend     string    `json:"intend"`
	CreateTime time.Time `json:"createTime"`
	EndTime    time.Time `json:"endTime"`
	User       string    `json:"user"`
	TaskName   string    `json:"taskName"`
	TaskDesc   string    `json:"taskDesc"`
	Status     string    `json:"status"`
	DueDate    string    `json:"dueDate"`
}

const path = "/Users/yitiansong/data/sync/todo.json"

//const path = "/Users/yitiansong/data/code/todo-go/todo.json"

const cmd = `
<System>
You are a todo helper agent, your task is get to know what the user want to execute.Get the ability you can do in the <ability> tag. return the match item name in the ability tag. 
<System>

<ability>
<item>
	<name>create<name>
	<desc>user want to create a task<desc>
<item>
<item>
	<name>delete<name>
	<desc>user want to delete a task<desc>
<item>
<item>
	<name>list<name>
	<desc>user want to see all the todolist<desc>
<item>
<item>
	<name>complete<name>
	<desc>user want to complete a task<desc>
<item>

<ability>

return format should following this format:
{
	"intend": "tag placed here",
	"taskId": -1,
	"user": "if not metioned, You is default",
	"createTime": "use current time",
	"endTime" "palce end time based on the current time"
	"taskName": "create a nice task name",
	"taskDesc": "give a clear description",
    "dueDate": "give a clear due date"
}

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
		//LogDebug( "[load] read file err :", err)
	}
	var loadingTodos []TodoItem = make([]TodoItem, 0)
	err = json.Unmarshal(bytes, &loadingTodos)
	if err != nil {
		//LogDebug( "[load] parse err")
	}
	return loadingTodos
}
func main() {
	args := os.Args
	LogDebug("args lens:%d\n", map[string]any{"len": len(args)})
	if len(args) < 2 {
		LogDebug("args must be exits", map[string]any{})
	}

	// starting request api

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	weekday := now.Weekday().String()
	// todo item
	var todos []TodoItem = LoadTodos(path)
	loadedbytes, _ := json.Marshal(todos)
	loadedTodos := string(loadedbytes)
	var userinput string = args[1]
	LogDebug("user input is :"+userinput, map[string]any{})
	if userinput == "list" || userinput == "ls" {
		List(&todos)
		return
	} else if arg := strings.Split(userinput, " "); len(arg) > 1 {
		LogDebug("arg is :"+arg[0], map[string]any{})
		if arg[0] == "complete" {
			id, err := strconv.Atoi(arg[1])
			LogDebug("[pcomplete] id is :"+strconv.Itoa(id), map[string]any{})
			if err != nil {
				LogDebug("[pcomplete] error occurs:"+err.Error(), map[string]any{"err": err})
			}
			Complete(&todos, &TodoItem{TaskID: id})
			return
		} else if arg[0] == "delete" {
			id, err := strconv.Atoi(arg[1])
			LogDebug("[pdelete] id is :"+strconv.Itoa(id), map[string]any{})
			if err != nil {
				LogDebug("[pdelete] error occurs:"+err.Error(), map[string]any{"err": err})
			}
			DeleteTask(&todos, id)
			return
		}
	}
	ctx := "current time is" + nowStr + " and today is " + weekday + args[1] + ", current todos: " + loadedTodos
	LogDebug("ctx is :"+string(ctx), map[string]any{})
	req := OpenAIRequest{
		Model: "gpt-4o",
		Messages: []Msg{
			{Role: "system", Content: cmd},
			{Role: "user", Content: ctx},
		},
	}
	warpIntend, err := Chat(req)
	if err != nil {
		LogDebug("error occurs:"+err.Error(), map[string]any{"err": err})
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
	request, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		LogDebug("[command]error occured when create a request: %s", map[string]any{"err": err.Error()})
		return "", err
	}

	api := os.Getenv("OPENAI_API_KEY")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+api)
	request.Header.Set("Accept", "application/json")

	// do request
	res, err := client.Do(request)
	if err != nil {
		LogDebug("[command]error occured when get a response: %s", map[string]any{"err": err.Error()})
		return "", err
	}
	defer res.Body.Close()

	// handle response
	if res.StatusCode != http.StatusOK {
		LogDebug("[command]API returned status %d\n", map[string]any{"statusCode": res.StatusCode})
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		LogDebug("[command]error reading response: %s\n", map[string]any{"err": err.Error()})
		return "", err
	}

	LogDebug("Raw response: %s\n", map[string]any{"row res": string(resBody)})

	var openAiResponse = OpenAIResponse{}
	err = json.Unmarshal(resBody, &openAiResponse)
	if err != nil {
		LogDebug("[command]error occured when parse a response: %s", map[string]any{"err": err.Error()})
		return "", err
	}

	LogDebug("response: %+v", map[string]any{"[api][response]": openAiResponse})

	// get the ai response
	msg := openAiResponse.Choices[0].Message.Content
	return msg, nil
}

func DoI(todoStr string, todos *[]TodoItem) {
	var todo = TodoItem{}
	err := json.Unmarshal([]byte(todoStr), &todo)
	if err != nil {
		LogDebug("error occuurs: ", map[string]any{"err": err.Error()})
	}
	intend := todo.Intend
	switch intend {
	case "create":
		CreateTask(todos, &todo)
	case "list":
		List(todos)
	case "complete":
		Complete(todos, &todo)
	case "delete":
		DeleteTask(todos, todo.TaskID)
	}

}

func Complete(todos *[]TodoItem, todo *TodoItem) {
	id := todo.TaskID
	if id <= 0 {
		LogDebug("[complete]id is invalid", map[string]any{})
		return
	}
	for i := 0; i < len((*todos)); i++ {
		if (*todos)[i].TaskID == id {
			LogDebug("[complete] task id is :"+strconv.Itoa(id), map[string]any{"name": (*todos)[i].TaskName, "desc": (*todos)[i].TaskDesc})
			(*todos)[i].Status = "completed"
			SaveTodos(todos, path)
			LogDebug("[complete] task saved", map[string]any{})
			fmt.Println("task completed successfully")
			return
		}
	}
	LogDebug("[complete] task not found", map[string]any{"id": id})
	fmt.Printf("Task with ID %d not found\n", id)
}

func CreateTask(todos *[]TodoItem, todo *TodoItem) {
	// TODO(human): Implement the core task creation logic
	// 1. Generate a unique TaskID (hint: find the highest existing ID + 1)
	id := GetLastId(todos)
	// 2. Set the Status field to "pending" or "active"
	todo.TaskID = id
	todo.Status = "pending"
	// 3. Add the new todo to the todos slice
	*todos = append(*todos, *todo)
	// 4. Call SaveTodos to persist to file
	err := SaveTodos(todos, path)
	if err != nil {
		LogDebug("[create] Failed to save todo: %v\n", map[string]any{"err": err})
	}
}

func GetLastId(todos *[]TodoItem) int {
	todoList := *todos
	length := len(todoList)
	if length < 1 {
		return 1
	}
	return todoList[length-1].TaskID + 1
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

	LogDebug("[save] Successfully saved todos to file", map[string]any{})
	return nil
}

func List(todos *[]TodoItem) {
	alfredItems := TransToAlfredItem(todos)
	response := AlfredResponse{Items: *alfredItems}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		LogDebug("[list] Failed to marshal todos: %v\n", map[string]any{"err": err})
	}
	fmt.Println(string(data))
}

func TransToAlfredItem(todos *[]TodoItem) *[]AlfredItem {
	var items = make([]AlfredItem, 0)
	for i := 0; i < len(*todos); i++ {
		item := AlfredItem{}
		item.Title = (*todos)[i].TaskName + "⏰Due date" + (*todos)[i].DueDate
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

func LogDebug(message string, data map[string]any) {

	logFilePath := "/tmp/todo-app.log"

	logEntry := map[string]any{
		"data": time.Now().Format(time.RFC3339),
		"log":  message,
	}

	for k, v := range data {
		logEntry[k] = v
	}
	jsonData, _ := json.Marshal(logEntry)
	writeToLogFile(logFilePath, string(jsonData))
}

func writeToLogFile(logFilePath, content string) error {
	// Check and rotate log file if necessary
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content + "\n")
	return err
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
