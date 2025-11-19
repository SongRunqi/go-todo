package domain

import "time"

// Tag represents a tag that can be associated with tasks
type Tag struct {
	Name       string    `json:"name"`
	Color      string    `json:"color,omitempty"`      // Optional color for display (e.g., "#FF0000")
	CreatedAt  time.Time `json:"createdAt"`
	UsageCount int       `json:"usageCount,omitempty"` // Number of tasks using this tag
}

// TagStore defines the interface for tag storage operations
type TagStore interface {
	Load() ([]Tag, error)
	Save(tags []Tag) error
}
