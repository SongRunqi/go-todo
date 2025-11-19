package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SongRunqi/go-todo/internal/domain"
	"github.com/SongRunqi/go-todo/internal/logger"
)

// FileTagStore implements file-based storage for tags
type FileTagStore struct {
	Path string
}

// NewFileTagStore creates a new file-based tag store
func NewFileTagStore(path string) *FileTagStore {
	return &FileTagStore{
		Path: path,
	}
}

// Load loads tags from file
func (f *FileTagStore) Load() ([]domain.Tag, error) {
	bytes, err := os.ReadFile(f.Path)
	if err != nil {
		// If file doesn't exist, create an empty file
		if os.IsNotExist(err) {
			logger.Debug("Tag file does not exist, creating new file: " + f.Path)
			emptyTags := make([]domain.Tag, 0)
			// Create empty JSON array file
			if err := f.Save(emptyTags); err != nil {
				logger.ErrorWithErr(err, "Failed to create new tag file")
				return emptyTags, fmt.Errorf("failed to create new tag file: %w", err)
			}
			return emptyTags, nil
		}
		logger.ErrorWithErr(err, "Failed to read tag file")
		return make([]domain.Tag, 0), fmt.Errorf("failed to read tag file: %w", err)
	}
	var loadingTags []domain.Tag = make([]domain.Tag, 0)
	err = json.Unmarshal(bytes, &loadingTags)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to parse tag JSON")
		return make([]domain.Tag, 0), fmt.Errorf("failed to parse tag JSON: %w", err)
	}
	return loadingTags, nil
}

// Save saves tags to file
func (f *FileTagStore) Save(tags []domain.Tag) error {
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	err = os.WriteFile(f.Path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write tag file: %w", err)
	}

	logger.Debug("Successfully saved tags to file")
	return nil
}
