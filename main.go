package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//const path = "/Users/yitiansong/data/code/todo-go/todo.json"

var fileTodoStore = FileTodoStore{path, backupPath}

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
	var todos []TodoItem = fileTodoStore.Load(true)
	loadedbytes, _ := json.Marshal(todos)
	loadedTodos := string(loadedbytes)
	var userinput string = args[1]
	log.Println("user input is:", userinput)
	if userinput == "list" || userinput == "ls" {
		List(&todos)
		return
	} else if userinput == "back" {
		backupTodos := fileTodoStore.Load(true)
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
	ctx := "current time is" + nowStr + " and today is " + weekday + ". user input: " + args[1] + ", current todos: " + loadedTodos
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
