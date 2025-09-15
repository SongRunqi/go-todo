package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const path = "/Users/yitiansong/data/sync/todo.json"
const backupPath = "/Users/yitiansong/data/sync/todo_back.json"

type FileTodoStore struct {
	Path       string
	BackupPath string
}

func (f *FileTodoStore) Load(backup bool) []TodoItem {
	filePath := f.Path
	if backup {
		filePath = backupPath
	}
	bytes, err := os.ReadFile(filePath)
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

func (f *FileTodoStore) Save(todos *[]TodoItem, backup bool) error {
	filePath := f.Path
	if backup {
		filePath = backupPath
	}
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
