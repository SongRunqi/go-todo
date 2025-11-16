package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SongRunqi/go-todo/internal/domain"
	"github.com/SongRunqi/go-todo/internal/logger"
)

// FileTodoStore implements file-based storage for todos
type FileTodoStore struct {
	Path       string
	BackupPath string
}

// NewFileTodoStore creates a new file-based todo store
func NewFileTodoStore(path, backupPath string) *FileTodoStore {
	return &FileTodoStore{
		Path:       path,
		BackupPath: backupPath,
	}
}

// Load loads todos from file
func (f *FileTodoStore) Load(backup bool) ([]domain.TodoItem, error) {
	filePath := f.Path
	if backup {
		filePath = f.BackupPath
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，创建一个空的文件
		if os.IsNotExist(err) {
			logger.Debug("File does not exist, creating new file: " + filePath)
			emptyTodos := make([]domain.TodoItem, 0)
			// 创建空的 JSON 数组文件
			if err := f.Save(emptyTodos, backup); err != nil {
				logger.ErrorWithErr(err, "Failed to create new file")
				return emptyTodos, fmt.Errorf("failed to create new file: %w", err)
			}
			return emptyTodos, nil
		}
		logger.ErrorWithErr(err, "Failed to read file")
		return make([]domain.TodoItem, 0), fmt.Errorf("failed to read file: %w", err)
	}
	var loadingTodos []domain.TodoItem = make([]domain.TodoItem, 0)
	err = json.Unmarshal(bytes, &loadingTodos)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to parse JSON")
		return make([]domain.TodoItem, 0), fmt.Errorf("failed to parse JSON: %w", err)
	}
	return loadingTodos, nil
}

// Save saves todos to file
func (f *FileTodoStore) Save(todos []domain.TodoItem, backup bool) error {
	filePath := f.Path
	if backup {
		filePath = f.BackupPath
	}
	data, err := json.MarshalIndent(todos, "", "  ")
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
