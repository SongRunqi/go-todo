package storage

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/SongRunqi/go-todo/internal/domain"
)

type TodoItem = domain.TodoItem

// MemoryTodoStore implements in-memory storage for testing
type MemoryTodoStore struct {
	data map[string][]TodoItem
	mu   sync.RWMutex
}

// NewMemoryStore creates a new memory-based store
func NewMemoryStore() *MemoryTodoStore {
	return &MemoryTodoStore{
		data: make(map[string][]TodoItem),
	}
}

// Load loads todos from memory
func (m *MemoryTodoStore) Load(backup bool) ([]TodoItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := "active"
	if backup {
		key = "backup"
	}

	todos, ok := m.data[key]
	if !ok {
		// Return empty slice if no data
		return []TodoItem{}, nil
	}

	// Return a copy to prevent external modifications
	result := make([]TodoItem, len(todos))
	copy(result, todos)

	return result, nil
}

// Save saves todos to memory
func (m *MemoryTodoStore) Save(todos []TodoItem, backup bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := "active"
	if backup {
		key = "backup"
	}

	// Store a copy to prevent external modifications
	data := make([]TodoItem, len(todos))
	copy(data, todos)

	m.data[key] = data
	return nil
}

// Clear clears all data (useful for testing)
func (m *MemoryTodoStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string][]TodoItem)
}

// Size returns the number of todos in the specified storage
func (m *MemoryTodoStore) Size(backup bool) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := "active"
	if backup {
		key = "backup"
	}

	if todos, ok := m.data[key]; ok {
		return len(todos)
	}
	return 0
}

// ToJSON converts the store data to JSON (for debugging)
func (m *MemoryTodoStore) ToJSON(backup bool) (string, error) {
	todos, err := m.Load(backup)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(data), nil
}

// FromJSON loads data from JSON string (for testing)
func (m *MemoryTodoStore) FromJSON(jsonData string, backup bool) error {
	var todos []TodoItem
	if err := json.Unmarshal([]byte(jsonData), &todos); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return m.Save(todos, backup)
}
