package app

import (
	"fmt"
	"log"
	"os"
	"time"
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
	// Load configuration
	config := LoadConfig()

	// Log all configuration
	log.Println("=== Configuration ===")
	log.Printf("[config] Todo path: %s", config.TodoPath)
	log.Printf("[config] Backup path: %s", config.BackupPath)
	log.Printf("[config] Model: %s", config.Model)
	log.Printf("[config] API Key: %s", maskAPIKey(config.APIKey))
	log.Printf("[config] LLM Base URL: %s", config.LLMBaseURL)
	log.Println("=====================")

	store := &FileTodoStore{Path: config.TodoPath, BackupPath: config.BackupPath}

	// Load todos
	todos, err := store.Load(false)
	if err != nil {
		log.Fatalf("Failed to load todos: %v", err)
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
