package app

import (
	"fmt"
	"os"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
)

// maskAPIKey masks the API key for logging
func maskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func main() {
	// Initialize logger
	logger.Init("info")

	// Load configuration
	config := LoadConfig()

	// Log all configuration
	logger.Info("=== Configuration ===")
	logger.Infof("Todo path: %s", config.TodoPath)
	logger.Infof("Backup path: %s", config.BackupPath)
	logger.Infof("Model: %s", config.Model)
	logger.Infof("API Key: %s", maskAPIKey(config.APIKey))
	logger.Infof("LLM Base URL: %s", config.LLMBaseURL)
	logger.Info("=====================")

	store := &FileTodoStore{Path: config.TodoPath, BackupPath: config.BackupPath}

	// Load todos
	todos, err := store.Load(false)
	if err != nil {
		logger.Fatalf("Failed to load todos: %v", err)
	}

	// Create execution context
	ctx := &Context{
		Store:       store,
		Todos:       &todos,
		Args:        os.Args,
		CurrentTime: time.Now(),
		Config:      &config,
	}

	// Create router and route command
	router := NewRouter()
	if err := router.Route(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
