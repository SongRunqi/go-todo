package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type FileTodoStore struct {
	Path       string
	BackupPath string
}

func (f *FileTodoStore) Load(backup bool) ([]TodoItem, error) {
	filePath := f.Path
	if backup {
		filePath = f.BackupPath
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("[load] read file err:", err)
		return make([]TodoItem, 0), fmt.Errorf("failed to read file: %w", err)
	}
	var loadingTodos []TodoItem = make([]TodoItem, 0)
	err = json.Unmarshal(bytes, &loadingTodos)
	if err != nil {
		log.Println("[load] parse err", err.Error())
		return make([]TodoItem, 0), fmt.Errorf("failed to parse JSON: %w", err)
	}
	return loadingTodos, nil
}

func (f *FileTodoStore) Save(todos *[]TodoItem, backup bool) error {
	filePath := f.Path
	if backup {
		filePath = f.BackupPath
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
