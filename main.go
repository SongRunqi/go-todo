package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type TodoItem struct {
	TaskID     int       `json:"taskId"`
	Intend     string    `json:"intend"`
	CreateTime time.Time `json:"createTime"`
	EndTime    time.Time `json:"endTime"`
	User       string    `json:"user"`
	TaskName   string    `json:"taskName"`
	TaskDesc   string    `json:"taskDesc"`
	Status     string    `json:"status"`
}

const path = "/Users/yitiansong/data/sync/todo.json"

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

<ability>

return format should following this format:
{
	"intend": "tag placed here",
	"taskId": -1,
	"user": "if not metioned, You is default",
	"createTime": "use current time",
	"endTime" "palce end time based on the current time"
	"taskName": "create a nice task name",
	"taskDesc": "give a clear description"
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

func main() {
	args := os.Args
	fmt.Printf("args lens:%d\n", len(args))
	if len(args) < 2 {
		fmt.Printf("args must be exits")
	}

	// starting request api

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	weekday := now.Weekday().String()
	ctx := "current time is" + nowStr + " and today is " + weekday + args[1]
	fmt.Println("ctx is :", ctx)
	req := OpenAIRequest{
		Model: "gpt-4o",
		Messages: []Msg{
			{Role: "system", Content: cmd},
			{Role: "user", Content: ctx},
		},
	}
	warpIntend, err := Chat(req)
	if err != nil {
		fmt.Println("error occurs:" + err.Error())
		return
	}
	// unwarp
	DoI(warpIntend)

}

func Chat(req OpenAIRequest) (string, error) {
	// struct -> json
	b, _ := json.Marshal(req)

	// create a client
	client := &http.Client{}
	// create a  http request
	request, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		fmt.Printf("[command]error occured when create a request: %s", err.Error())
		return "", err
	}

	api := os.Getenv("OPENAI_API_KEY")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+api)
	request.Header.Set("Accept", "application/json")

	// do request
	res, err := client.Do(request)
	if err != nil {
		fmt.Printf("[command]error occured when get a response: %s", err.Error())
		return "", err
	}
	defer res.Body.Close()

	// handle response
	if res.StatusCode != http.StatusOK {
		fmt.Printf("[command]API returned status %d\n", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("[command]error reading response: %s\n", err.Error())
		return "", err
	}

	fmt.Printf("Raw response: %s\n", string(resBody))

	var openAiResponse = OpenAIResponse{}
	err = json.Unmarshal(resBody, &openAiResponse)
	if err != nil {
		fmt.Printf("[command]error occured when parse a response: %s", err.Error())
		return "", err
	}

	fmt.Printf("response: %+v", openAiResponse)

	// get the ai response
	msg := openAiResponse.Choices[0].Message.Content
	return msg, nil
}

func DoI(todoStr string) {
	var todo = TodoItem{}
	err := json.Unmarshal([]byte(todoStr), &todo)
	if err != nil {
		fmt.Println("error occuurs: ", err.Error())
	}
	intend := todo.Intend
	switch intend {
	case "create":
		CreateTask(todo)
	}
}

func LoadTodos(path string) ([]TodoItem, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return []TodoItem{}, err
	}

	var items []TodoItem
	if err := json.Unmarshal(b, &items); err != nil {
		return items, err
	}
	return items, nil
}

func saveTodos(path string, todos []TodoItem) {
	b, err := json.MarshalIndent(todos, "", " ")
	if err != nil {
		fmt.Println("[save] when marshal todoitems", err.Error())
	}

	os.WriteFile(path, b, 0644)
}
func CreateTask(todo TodoItem) {
	todos, err := LoadTodos(path)
	if err != nil {
		fmt.Println("[create] load todos:", err.Error())
	}
	if len(todos) < 1 {
		todo.TaskID = 1
	} else {
		lastID := todos[len(todos)-1].TaskID
		todo.TaskID = lastID + 1
	}
	todos = append(todos, todo)
	saveTodos(path, todos)

}
