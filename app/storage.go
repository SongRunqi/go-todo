package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SongRunqi/go-todo/internal/logger"
)

func (f *FileTodoStore) Load(backup bool) ([]TodoItem, error) {
	filePath := f.Path
	if backup {
		filePath = f.BackupPath
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，创建一个空的文件
		if os.IsNotExist(err) {
			logger.Debug("File does not exist, creating new file: " + filePath)
			emptyTodos := make([]TodoItem, 0)
			// 创建空的 JSON 数组文件
			if err := f.Save(&emptyTodos, backup); err != nil {
				logger.ErrorWithErr(err, "Failed to create new file")
				return emptyTodos, fmt.Errorf("failed to create new file: %w", err)
			}
			return emptyTodos, nil
		}
		logger.ErrorWithErr(err, "Failed to read file")
		return make([]TodoItem, 0), fmt.Errorf("failed to read file: %w", err)
	}
	var loadingTodos []TodoItem = make([]TodoItem, 0)
	err = json.Unmarshal(bytes, &loadingTodos)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to parse JSON")
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

	logger.Debug("Successfully saved todos to file")
	return nil
}
