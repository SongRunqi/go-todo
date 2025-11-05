package app

import (
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	TodoPath   string
	BackupPath string
	APIKey     string
	Model      string
	LLMBaseURL string
	Language   string
}

// LoadConfig loads configuration from environment variables with fallback defaults
func LoadConfig() Config {
	// Get user home directory as a fallback base path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Default paths in user's home directory
	defaultTodoPath := filepath.Join(homeDir, ".todo", "todo.json")
	defaultBackupPath := filepath.Join(homeDir, ".todo", "todo_back.json")

	// Load from environment variables or use defaults
	todoPath := getEnvOrDefault("TODO_PATH", defaultTodoPath)
	backupPath := getEnvOrDefault("TODO_BACKUP_PATH", defaultBackupPath)

	// Ensure the directory exists
	todoDir := filepath.Dir(todoPath)
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		// If we can't create the directory, fall back to current directory
		todoPath = "todo.json"
		backupPath = "todo_back.json"
	}

	// Load AI/LLM configuration
	apiKey := os.Getenv("API_KEY")
	model := getEnvOrDefault("model", "deepseek-chat")
	llmBaseURL := getEnvOrDefault("LLM_BASE_URL", "https://api.deepseek.com/chat/completions")

	// Load language configuration (defaults to auto-detect from environment)
	language := getEnvOrDefault("TODO_LANG", "")

	return Config{
		TodoPath:   todoPath,
		BackupPath: backupPath,
		APIKey:     apiKey,
		Model:      model,
		LLMBaseURL: llmBaseURL,
		Language:   language,
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
