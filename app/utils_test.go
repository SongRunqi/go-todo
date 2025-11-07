package app

import (
	"testing"
	"time"
)

func TestSortedListWithSameEndTime(t *testing.T) {
	// Create test todos with the same end time
	now := time.Now()
	endTime := now.Add(3 * 24 * time.Hour) // 3 days from now

	todos := []TodoItem{
		{
			TaskID:     1,
			TaskName:   "Task 1",
			TaskDesc:   "Description 1",
			EndTime:    endTime,
			CreateTime: now,
			Status:     "pending",
		},
		{
			TaskID:     2,
			TaskName:   "Task 2",
			TaskDesc:   "Description 2",
			EndTime:    endTime, // Same end time as Task 1
			CreateTime: now,
			Status:     "pending",
		},
		{
			TaskID:     3,
			TaskName:   "Task 3",
			TaskDesc:   "Description 3",
			EndTime:    endTime, // Same end time as Task 1 and 2
			CreateTime: now,
			Status:     "pending",
		},
	}

	// Call sortedList
	result := sortedList(&todos)

	// Verify all tasks are in the result
	if len(result) != 3 {
		t.Errorf("Expected 3 tasks in result, got %d", len(result))
	}

	// Verify all task IDs are present
	taskIDs := make(map[int]bool)
	for _, task := range result {
		taskIDs[task.TaskID] = true
	}

	for i := 1; i <= 3; i++ {
		if !taskIDs[i] {
			t.Errorf("Task ID %d is missing from the result", i)
		}
	}
}

func TestSortedListWithDifferentEndTimes(t *testing.T) {
	// Create test todos with different end times
	now := time.Now()

	todos := []TodoItem{
		{
			TaskID:     1,
			TaskName:   "Task 1",
			TaskDesc:   "Description 1",
			EndTime:    now.Add(1 * 24 * time.Hour), // 1 day
			CreateTime: now,
			Status:     "pending",
		},
		{
			TaskID:     2,
			TaskName:   "Task 2",
			TaskDesc:   "Description 2",
			EndTime:    now.Add(3 * 24 * time.Hour), // 3 days
			CreateTime: now,
			Status:     "pending",
		},
		{
			TaskID:     3,
			TaskName:   "Task 3",
			TaskDesc:   "Description 3",
			EndTime:    now.Add(2 * 24 * time.Hour), // 2 days
			CreateTime: now,
			Status:     "pending",
		},
	}

	// Call sortedList
	result := sortedList(&todos)

	// Verify all tasks are in the result
	if len(result) != 3 {
		t.Errorf("Expected 3 tasks in result, got %d", len(result))
	}

	// Verify tasks are sorted by end time (earliest first)
	if result[0].TaskID != 1 {
		t.Errorf("Expected Task 1 first (1 day), got Task %d", result[0].TaskID)
	}
	if result[1].TaskID != 3 {
		t.Errorf("Expected Task 3 second (2 days), got Task %d", result[1].TaskID)
	}
	if result[2].TaskID != 2 {
		t.Errorf("Expected Task 2 third (3 days), got Task %d", result[2].TaskID)
	}
}
