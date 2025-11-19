package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/domain"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/repository"
)

// Re-export domain types for backward compatibility
type Tag = domain.Tag
type TagStore = domain.TagStore
type FileTagStore = repository.FileTagStore

// ListTags lists all available tags
func ListTags(tagStore TagStore) error {
	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	fmt.Println("Available tags:")
	for _, tag := range tags {
		usageInfo := ""
		if tag.UsageCount > 0 {
			usageInfo = fmt.Sprintf(" (used by %d task(s))", tag.UsageCount)
		}
		fmt.Printf("  - %s%s\n", tag.Name, usageInfo)
	}

	return nil
}

// AddTag creates a new tag
func AddTag(tagStore TagStore, name string, color string) error {
	if name == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Normalize tag name (trim spaces)
	name = strings.TrimSpace(name)

	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	// Check if tag already exists
	for _, tag := range tags {
		if tag.Name == name {
			return fmt.Errorf("tag '%s' already exists", name)
		}
	}

	// Create new tag
	newTag := Tag{
		Name:       name,
		Color:      color,
		CreatedAt:  time.Now(),
		UsageCount: 0,
	}

	tags = append(tags, newTag)

	if err := tagStore.Save(tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	logger.Debugf("Tag '%s' created successfully", name)
	fmt.Printf("Tag '%s' created successfully\n", name)
	return nil
}

// DeleteTag removes a tag
func DeleteTag(tagStore TagStore, todoStore TodoStore, name string) error {
	if name == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Normalize tag name
	name = strings.TrimSpace(name)

	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	// Find and remove the tag
	tagFound := false
	newTags := make([]Tag, 0)
	for _, tag := range tags {
		if tag.Name == name {
			tagFound = true
			continue
		}
		newTags = append(newTags, tag)
	}

	if !tagFound {
		return fmt.Errorf("tag '%s' not found", name)
	}

	// Remove the tag from all tasks
	todos, err := todoStore.Load(false)
	if err != nil {
		return fmt.Errorf("failed to load todos: %w", err)
	}

	taskUpdated := false
	for i := range todos {
		newTaskTags := make([]string, 0)
		for _, t := range todos[i].Tags {
			if t != name {
				newTaskTags = append(newTaskTags, t)
			} else {
				taskUpdated = true
			}
		}
		todos[i].Tags = newTaskTags
	}

	// Save updated todos if any task was modified
	if taskUpdated {
		if err := todoStore.Save(todos, false); err != nil {
			return fmt.Errorf("failed to update tasks: %w", err)
		}
	}

	// Save updated tags
	if err := tagStore.Save(newTags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	logger.Debugf("Tag '%s' deleted successfully", name)
	fmt.Printf("Tag '%s' deleted successfully\n", name)
	return nil
}

// RenameTag renames an existing tag
func RenameTag(tagStore TagStore, todoStore TodoStore, oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("tag names cannot be empty")
	}

	// Normalize tag names
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)

	if oldName == newName {
		return fmt.Errorf("old and new tag names are the same")
	}

	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	// Check if old tag exists and new tag doesn't exist
	oldTagFound := false
	newTagExists := false
	for i, tag := range tags {
		if tag.Name == oldName {
			oldTagFound = true
			tags[i].Name = newName
		}
		if tag.Name == newName {
			newTagExists = true
		}
	}

	if !oldTagFound {
		return fmt.Errorf("tag '%s' not found", oldName)
	}

	if newTagExists {
		return fmt.Errorf("tag '%s' already exists", newName)
	}

	// Update the tag in all tasks
	todos, err := todoStore.Load(false)
	if err != nil {
		return fmt.Errorf("failed to load todos: %w", err)
	}

	taskUpdated := false
	for i := range todos {
		for j, t := range todos[i].Tags {
			if t == oldName {
				todos[i].Tags[j] = newName
				taskUpdated = true
			}
		}
	}

	// Save updated todos if any task was modified
	if taskUpdated {
		if err := todoStore.Save(todos, false); err != nil {
			return fmt.Errorf("failed to update tasks: %w", err)
		}
	}

	// Save updated tags
	if err := tagStore.Save(tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	logger.Debugf("Tag renamed from '%s' to '%s'", oldName, newName)
	fmt.Printf("Tag renamed from '%s' to '%s'\n", oldName, newName)
	return nil
}

// AddTagToTask adds a tag to a task
func AddTagToTask(todos *[]TodoItem, taskID int, tagName string, todoStore TodoStore, tagStore TagStore) error {
	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Normalize tag name
	tagName = strings.TrimSpace(tagName)

	// Load tags to check if tag exists
	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	// Check if tag exists
	tagExists := false
	for _, tag := range tags {
		if tag.Name == tagName {
			tagExists = true
			break
		}
	}

	if !tagExists {
		return fmt.Errorf("tag '%s' does not exist. Please create it first using 'tag add %s'", tagName, tagName)
	}

	// Find the task
	var task *TodoItem
	for i := range *todos {
		if (*todos)[i].TaskID == taskID {
			task = &(*todos)[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task with ID %d not found", taskID)
	}

	// Check if task already has this tag
	for _, t := range task.Tags {
		if t == tagName {
			return fmt.Errorf("task %d already has tag '%s'", taskID, tagName)
		}
	}

	// Add tag to task
	if task.Tags == nil {
		task.Tags = make([]string, 0)
	}
	task.Tags = append(task.Tags, tagName)

	// Update tag usage count
	for i := range tags {
		if tags[i].Name == tagName {
			tags[i].UsageCount++
			break
		}
	}

	// Save updates
	if err := todoStore.Save(*todos, false); err != nil {
		return fmt.Errorf("failed to save tasks: %w", err)
	}

	if err := tagStore.Save(tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	logger.Debugf("Tag '%s' added to task %d", tagName, taskID)
	fmt.Printf("Tag '%s' added to task %d: %s\n", tagName, taskID, task.TaskName)
	return nil
}

// RemoveTagFromTask removes a tag from a task
func RemoveTagFromTask(todos *[]TodoItem, taskID int, tagName string, todoStore TodoStore, tagStore TagStore) error {
	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Normalize tag name
	tagName = strings.TrimSpace(tagName)

	// Find the task
	var task *TodoItem
	for i := range *todos {
		if (*todos)[i].TaskID == taskID {
			task = &(*todos)[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task with ID %d not found", taskID)
	}

	// Check if task has this tag
	tagFound := false
	newTags := make([]string, 0)
	for _, t := range task.Tags {
		if t == tagName {
			tagFound = true
			continue
		}
		newTags = append(newTags, t)
	}

	if !tagFound {
		return fmt.Errorf("task %d does not have tag '%s'", taskID, tagName)
	}

	// Update task tags
	task.Tags = newTags

	// Update tag usage count
	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	for i := range tags {
		if tags[i].Name == tagName {
			if tags[i].UsageCount > 0 {
				tags[i].UsageCount--
			}
			break
		}
	}

	// Save updates
	if err := todoStore.Save(*todos, false); err != nil {
		return fmt.Errorf("failed to save tasks: %w", err)
	}

	if err := tagStore.Save(tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	logger.Debugf("Tag '%s' removed from task %d", tagName, taskID)
	fmt.Printf("Tag '%s' removed from task %d: %s\n", tagName, taskID, task.TaskName)
	return nil
}

// RecalculateTagUsage recalculates the usage count for all tags
func RecalculateTagUsage(todoStore TodoStore, tagStore TagStore) error {
	tags, err := tagStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	todos, err := todoStore.Load(false)
	if err != nil {
		return fmt.Errorf("failed to load todos: %w", err)
	}

	// Reset all usage counts
	for i := range tags {
		tags[i].UsageCount = 0
	}

	// Count tag usage from all tasks
	for _, todo := range todos {
		for _, tagName := range todo.Tags {
			for i := range tags {
				if tags[i].Name == tagName {
					tags[i].UsageCount++
					break
				}
			}
		}
	}

	// Save updated tags
	if err := tagStore.Save(tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	logger.Debug("Tag usage counts recalculated successfully")
	return nil
}
