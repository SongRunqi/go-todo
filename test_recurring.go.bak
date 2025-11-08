package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/SongRunqi/go-todo/app"
)

func main() {
	fmt.Println("=== Testing Recurring Tasks with EventDuration ===\n")

	// Test 1: Create a recurring task with event duration
	fmt.Println("Test 1: Creating recurring task 'Monday/Wednesday 2pm-3pm class for 4 weeks'")

	task := &app.TodoItem{
		TaskName:          "周三、周五下午2点到3点上课",
		TaskDesc:          "每周三和周五下午上课，共4周",
		User:              "TestUser",
		CreateTime:        time.Now(),
		EndTime:           getNextWeekday(time.Now(), 3, 14, 0), // Next Wednesday at 14:00
		EventDuration:     1 * time.Hour, // 1 hour duration (2pm-3pm)
		Urgent:            "medium",
		IsRecurring:       true,
		RecurringType:     "weekly",
		RecurringInterval: 1,
		RecurringWeekdays: []int{3, 5}, // Wednesday=3, Friday=5
		RecurringMaxCount: 4,            // 4 weeks
	}

	todos := []app.TodoItem{}
	err := app.CreateTask(&todos, task)
	if err != nil {
		fmt.Printf("❌ Error creating task: %v\n", err)
		return
	}

	fmt.Printf("✅ Task created with ID: %d\n", task.TaskID)
	fmt.Printf("   Status: %s\n", task.Status)
	fmt.Printf("   EndTime: %s\n", task.EndTime.Format("2006-01-02 15:04"))
	fmt.Printf("   EventDuration: %v\n", task.EventDuration)
	fmt.Printf("   OccurrenceHistory count: %d\n", len(task.OccurrenceHistory))

	// Display occurrence history
	fmt.Println("\n   Occurrence History:")
	for i, occ := range task.OccurrenceHistory {
		endTime := occ.ScheduledTime.Add(task.EventDuration)
		fmt.Printf("   %d. %s %s - %s [%s]\n",
			i+1,
			occ.ScheduledTime.Weekday().String()[:3],
			occ.ScheduledTime.Format("2006-01-02 15:04"),
			endTime.Format("15:04"),
			occ.Status,
		)
	}

	// Test 2: Check task detail display
	fmt.Println("\n=== Test 2: Task Detail Display ===")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling task: %v\n", err)
		return
	}
	fmt.Println(string(data))

	// Test 3: Verify helper functions
	fmt.Println("\n=== Test 3: Helper Functions ===")

	currentOcc, idx := app.GetCurrentOccurrence(task)
	if currentOcc != nil {
		fmt.Printf("✅ GetCurrentOccurrence: Found occurrence at index %d\n", idx)
		fmt.Printf("   Scheduled: %s\n", currentOcc.ScheduledTime.Format("2006-01-02 15:04"))
	} else {
		fmt.Println("ℹ️  GetCurrentOccurrence: No current occurrence (expected if all are in the future)")
	}

	nextOcc, idx := app.GetNextPendingOccurrence(task)
	if nextOcc != nil {
		fmt.Printf("✅ GetNextPendingOccurrence: Found occurrence at index %d\n", idx)
		fmt.Printf("   Scheduled: %s\n", nextOcc.ScheduledTime.Format("2006-01-02 15:04"))
	} else {
		fmt.Println("❌ GetNextPendingOccurrence: No pending occurrence found")
	}

	isPeriodComplete := app.IsPeriodCompletedNew(task)
	fmt.Printf("✅ IsPeriodCompletedNew: %v (expected: false)\n", isPeriodComplete)

	fmt.Println("\n=== All Tests Completed ===")
}

// Helper function to get next occurrence of a specific weekday at a specific time
func getNextWeekday(from time.Time, weekday int, hour, minute int) time.Time {
	// Start from tomorrow to avoid same-day issues
	current := from.AddDate(0, 0, 1)

	// Find the next occurrence of the target weekday
	for int(current.Weekday()) != weekday {
		current = current.AddDate(0, 0, 1)
	}

	// Set the specific time
	return time.Date(
		current.Year(), current.Month(), current.Day(),
		hour, minute, 0, 0, current.Location(),
	)
}
