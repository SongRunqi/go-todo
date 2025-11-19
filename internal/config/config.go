package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	TodoPath   string
	BackupPath string
	TagPath    string
	APIKey     string
	Model      string
	LLMBaseURL string
	Language   string
}

var (
	cfg Config
)

// Load loads configuration from environment variables with fallback defaults
func Load() Config {
	// Load from config file if it exists
	if cfg != (Config{}) {
		return cfg
	}
	// Get user home directory as a fallback base path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Default paths in user's home directory
	defaultTodoPath := filepath.Join(homeDir, ".todo", "todo.json")
	defaultBackupPath := filepath.Join(homeDir, ".todo", "todo_back.json")
	defaultTagPath := filepath.Join(homeDir, ".todo", "tags.json")

	// Load from environment variables or use defaults
	todoPath := getEnvOrDefault("TODO_PATH", defaultTodoPath)
	backupPath := getEnvOrDefault("TODO_BACKUP_PATH", defaultBackupPath)
	tagPath := getEnvOrDefault("TAG_PATH", defaultTagPath)

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

	// Load language configuration
	// Priority: 1. Config file 2. Auto-detect
	language := ""
	if fileConfig := loadConfigFile(homeDir); fileConfig != nil {
		language = fileConfig.Language
	}
	cfg = Config{
		TodoPath:   todoPath,
		BackupPath: backupPath,
		TagPath:    tagPath,
		APIKey:     apiKey,
		Model:      model,
		LLMBaseURL: llmBaseURL,
		Language:   language,
	}
	return cfg
}

// loadConfigFile loads configuration from the config.json file
func loadConfigFile(homeDir string) *fileConfig {
	configFile := filepath.Join(homeDir, ".todo", "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		// Config file doesn't exist or can't be read - this is fine
		return nil
	}

	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Invalid JSON - this is fine, just ignore
		return nil
	}

	return &cfg
}

// fileConfig represents the structure of the config.json file
type fileConfig struct {
	Language string `json:"language"`
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
