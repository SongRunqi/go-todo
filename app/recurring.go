package app

import (
	"fmt"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/SongRunqi/go-todo/internal/validator"
)

func Complete(todos *[]TodoItem, todo *TodoItem, store *FileTodoStore) error {
	id := todo.TaskID
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			task := &(*todos)[i]
			taskName := task.TaskName
			logger.Debugf("Completing task ID %d: %s - %s", id, task.TaskName, task.TaskDesc)

			// Handle recurring tasks with new occurrence-based model
			if task.IsRecurring && len(task.OccurrenceHistory) > 0 {
				// Find the current occurrence to complete
				currentOcc, _ := GetCurrentOccurrence(task)

				// If no current due occurrence, try to find next pending (allow early completion)
				if currentOcc == nil {
					currentOcc, _ = GetNextPendingOccurrence(task)
				}

				if currentOcc == nil {
					return fmt.Errorf("no pending occurrence found to complete")
				}

				// Mark this occurrence as completed
				currentOcc.Status = "completed"
				currentOcc.CompletedAt = time.Now()
				logger.Infof("Marked occurrence at %s as completed", currentOcc.ScheduledTime.Format("2006-01-02 15:04"))

				// For weekday-specific weekly tasks, check if the period is complete
				if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 {
					if IsPeriodCompletedNew(task) {
						// Period completed! Increment completion count
						task.CompletionCount++

						// Check if max count is reached
						if task.RecurringMaxCount > 0 && task.CompletionCount >= task.RecurringMaxCount {
							task.Status = "completed"
							err := store.Save(todos, false)
							if err != nil {
								return fmt.Errorf("failed to save updated todos: %w", err)
							}

							logger.Infof("Recurring task completed for the final time. Total periods: %d/%d", task.CompletionCount, task.RecurringMaxCount)
							fmt.Printf("✅ Period completed! (%d/%d - Final period) 🎉\n", task.CompletionCount, task.RecurringMaxCount)
							return nil
						}

						// Create occurrences for next period
						nextPeriodOccurrences := CreateNextPeriodOccurrences(task)
						task.OccurrenceHistory = append(task.OccurrenceHistory, nextPeriodOccurrences...)

						// Update EndTime to first occurrence of next period
						if len(nextPeriodOccurrences) > 0 {
							task.EndTime = nextPeriodOccurrences[0].ScheduledTime
							task.DueDate = nextPeriodOccurrences[0].ScheduledTime.Format("2006-01-02")
						}

						err := store.Save(todos, false)
						if err != nil {
							return fmt.Errorf("failed to save updated todos: %w", err)
						}

						// Show count with max if specified
						countDisplay := fmt.Sprintf("%d", task.CompletionCount)
						if task.RecurringMaxCount > 0 {
							countDisplay = fmt.Sprintf("%d/%d", task.CompletionCount, task.RecurringMaxCount)
						}

						logger.Infof("Period completed. Count: %s, Next period starts: %s", countDisplay, task.EndTime.Format("2006-01-02 15:04"))
						fmt.Printf("✅ Period completed! (Count: %s) Next period starts: %s\n", countDisplay, task.EndTime.Format("2006-01-02 15:04"))
						return nil
					}

					// Period not complete, find next pending in current period
					nextOcc, _ := GetNextPendingOccurrence(task)
					if nextOcc != nil {
						task.EndTime = nextOcc.ScheduledTime
						task.DueDate = nextOcc.ScheduledTime.Format("2006-01-02")

						// Count completed occurrences in current week
						now := time.Now()
						weekStart := now
						for weekStart.Weekday() != time.Sunday {
							weekStart = weekStart.AddDate(0, 0, -1)
						}
						weekEnd := weekStart.AddDate(0, 0, 7)

						completedInWeek := 0
						for _, occ := range task.OccurrenceHistory {
							if !occ.ScheduledTime.Before(weekStart) && occ.ScheduledTime.Before(weekEnd) && occ.Status == "completed" {
								completedInWeek++
							}
						}

						err := store.Save(todos, false)
						if err != nil {
							return fmt.Errorf("failed to save updated todos: %w", err)
						}

						progressDisplay := fmt.Sprintf("%d/%d in this period", completedInWeek, len(task.RecurringWeekdays))
						logger.Infof("Sub-task completed. Progress: %s, Next: %s", progressDisplay, nextOcc.ScheduledTime.Format("2006-01-02 15:04"))
						fmt.Printf("✅ Sub-task completed! (%s) Next: %s\n", progressDisplay, nextOcc.ScheduledTime.Format("2006-01-02 15:04"))
						return nil
					}
				}

				// For other recurring types (daily, simple weekly, monthly, yearly)
				// Each completion counts as one period
				task.CompletionCount++

				// Check if max count is reached
				if task.RecurringMaxCount > 0 && task.CompletionCount >= task.RecurringMaxCount {
					task.Status = "completed"
					err := store.Save(todos, false)
					if err != nil {
						return fmt.Errorf("failed to save updated todos: %w", err)
					}

					logger.Infof("Recurring task completed for the final time. Total completions: %d/%d", task.CompletionCount, task.RecurringMaxCount)
					fmt.Printf("✅ Task completed! (%d/%d - Final completion) 🎉\n", task.CompletionCount, task.RecurringMaxCount)
					return nil
				}

				// Create next occurrence
				nextOccurrences := CreateNextPeriodOccurrences(task)
				task.OccurrenceHistory = append(task.OccurrenceHistory, nextOccurrences...)

				if len(nextOccurrences) > 0 {
					task.EndTime = nextOccurrences[0].ScheduledTime
					task.DueDate = nextOccurrences[0].ScheduledTime.Format("2006-01-02")
				}

				err := store.Save(todos, false)
				if err != nil {
					return fmt.Errorf("failed to save updated todos: %w", err)
				}

				// Show count with max if specified
				countDisplay := fmt.Sprintf("%d", task.CompletionCount)
				if task.RecurringMaxCount > 0 {
					countDisplay = fmt.Sprintf("%d/%d", task.CompletionCount, task.RecurringMaxCount)
				}

				logger.Infof("Recurring task completed. Count: %s, Next occurrence: %s", countDisplay, task.EndTime.Format("2006-01-02 15:04"))
				fmt.Printf("✅ Task completed! (Count: %s) Next occurrence: %s\n", countDisplay, task.EndTime.Format("2006-01-02 15:04"))
				return nil
			}

			// Handle legacy recurring tasks (with CurrentPeriodCompletions) - migrate to new model
			if task.IsRecurring && len(task.CurrentPeriodCompletions) > 0 {
				// TODO: Migration logic for old format
				return fmt.Errorf("please recreate this recurring task to use the new occurrence tracking system")
			}

			// Non-recurring task: mark as completed
			task.Status = "completed"

			err := store.Save(todos, false)
			if err != nil {
				return fmt.Errorf("failed to save updated todos: %w", err)
			}

			logger.Debug("Task marked as completed")
			output.PrintTaskCompleted(id, taskName)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

// GetCurrentOccurrence returns the current occurrence that should be completed
// Returns the occurrence and its index in the history, or -1 if not found
func GetCurrentOccurrence(task *TodoItem) (*OccurrenceRecord, int) {
	if !task.IsRecurring || len(task.OccurrenceHistory) == 0 {
		return nil, -1
	}

	now := time.Now()

	// Find the first pending occurrence that is due (scheduled time has passed or is today)
	for i := range task.OccurrenceHistory {
		occ := &task.OccurrenceHistory[i]
		if occ.Status == "pending" {
			// Check if it's due (scheduled time has passed or is within today)
			if !occ.ScheduledTime.After(now) {
				return occ, i
			}
		}
	}

	return nil, -1
}

// GetNextPendingOccurrence returns the next pending occurrence
func GetNextPendingOccurrence(task *TodoItem) (*OccurrenceRecord, int) {
	if !task.IsRecurring || len(task.OccurrenceHistory) == 0 {
		return nil, -1
	}

	for i := range task.OccurrenceHistory {
		occ := &task.OccurrenceHistory[i]
		if occ.Status == "pending" {
			return occ, i
		}
	}

	return nil, -1
}

// IsPeriodCompletedNew checks if the current period is completed based on OccurrenceHistory
func IsPeriodCompletedNew(task *TodoItem) bool {
	if !task.IsRecurring {
		return false
	}

	// For weekday-specific weekly tasks, check if all occurrences in current week are completed
	if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 {
		now := time.Now()
		weekStart := now
		for weekStart.Weekday() != time.Sunday {
			weekStart = weekStart.AddDate(0, 0, -1)
		}
		weekEnd := weekStart.AddDate(0, 0, 7)

		pendingInCurrentWeek := 0
		completedInCurrentWeek := 0

		for _, occ := range task.OccurrenceHistory {
			if !occ.ScheduledTime.Before(weekStart) && occ.ScheduledTime.Before(weekEnd) {
				if occ.Status == "pending" {
					pendingInCurrentWeek++
				} else if occ.Status == "completed" {
					completedInCurrentWeek++
				}
			}
		}

		// Period is completed if no pending occurrences left in current week
		// and we have completed at least some occurrences
		return pendingInCurrentWeek == 0 && completedInCurrentWeek > 0
	}

	// For other types, a single completion marks the period as complete
	return false
}

// CreateNextPeriodOccurrences creates occurrence records for the next period
func CreateNextPeriodOccurrences(task *TodoItem) []OccurrenceRecord {
	newOccurrences := []OccurrenceRecord{}

	if !task.IsRecurring {
		return newOccurrences
	}

	// For weekday-specific weekly tasks, create occurrences for next week
	if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 {
		// Find the start of next week
		now := time.Now()
		nextWeekStart := now
		for nextWeekStart.Weekday() != time.Sunday {
			nextWeekStart = nextWeekStart.AddDate(0, 0, -1)
		}
		nextWeekStart = nextWeekStart.AddDate(0, 0, 7) // Move to next week

		// Create occurrences for each required weekday
		for _, weekday := range task.RecurringWeekdays {
			scheduledTime := nextWeekStart.AddDate(0, 0, weekday)

			// Preserve the time of day from the task's EndTime
			scheduledTime = time.Date(
				scheduledTime.Year(), scheduledTime.Month(), scheduledTime.Day(),
				task.EndTime.Hour(), task.EndTime.Minute(), task.EndTime.Second(),
				0, task.EndTime.Location(),
			)

			newOccurrences = append(newOccurrences, OccurrenceRecord{
				ScheduledTime: scheduledTime,
				Status:        "pending",
			})
		}
	} else {
		// For other recurring types, create a single next occurrence
		nextTime := calculateNextOccurrence(task)
		newOccurrences = append(newOccurrences, OccurrenceRecord{
			ScheduledTime: nextTime,
			Status:        "pending",
		})
	}

	return newOccurrences
}

// MarkMissedOccurrences marks overdue pending occurrences as missed
func MarkMissedOccurrences(task *TodoItem) int {
	if !task.IsRecurring {
		return 0
	}

	now := time.Now()
	missedCount := 0

	for i := range task.OccurrenceHistory {
		occ := &task.OccurrenceHistory[i]
		if occ.Status == "pending" {
			// If scheduled time + event duration has passed, mark as missed
			endTime := occ.ScheduledTime.Add(task.EventDuration)
			if endTime.Before(now) {
				occ.Status = "missed"
				missedCount++
			}
		}
	}

	return missedCount
}

// initializeOccurrenceHistory creates initial occurrence records for a new recurring task
func initializeOccurrenceHistory(task *TodoItem) []OccurrenceRecord {
	history := []OccurrenceRecord{}

	if !task.IsRecurring {
		return history
	}

	// For weekday-specific weekly tasks, create records for all days in the current period (week)
	if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 {
		currentDate := task.EndTime // EndTime is set to the first scheduled occurrence

		// Find the start of the current week (Sunday)
		weekStart := currentDate
		for weekStart.Weekday() != time.Sunday {
			weekStart = weekStart.AddDate(0, 0, -1)
		}

		// Create an occurrence for each required weekday in the current period
		for _, weekday := range task.RecurringWeekdays {
			scheduledTime := weekStart.AddDate(0, 0, weekday)

			// Preserve the time of day from EndTime
			scheduledTime = time.Date(
				scheduledTime.Year(), scheduledTime.Month(), scheduledTime.Day(),
				task.EndTime.Hour(), task.EndTime.Minute(), task.EndTime.Second(),
				0, task.EndTime.Location(),
			)

			// Only add if it's in the future or today
			if !scheduledTime.Before(time.Now().Truncate(24 * time.Hour)) {
				history = append(history, OccurrenceRecord{
					ScheduledTime: scheduledTime,
					Status:        "pending",
				})
			}
		}
	} else {
		// For other recurring types (daily, simple weekly, monthly, yearly)
		// Create just the first occurrence
		history = append(history, OccurrenceRecord{
			ScheduledTime: task.EndTime,
			Status:        "pending",
		})
	}

	return history
}

// calculateNextOccurrence calculates the next occurrence time based on recurring type and interval
func calculateNextOccurrence(task *TodoItem) time.Time {
	current := task.EndTime
	recurringType := task.RecurringType
	interval := task.RecurringInterval

	switch recurringType {
	case "daily":
		return current.AddDate(0, 0, interval)

	case "weekly":
		// Check if specific weekdays are set
		if len(task.RecurringWeekdays) > 0 {
			return calculateNextWeekday(current, task.RecurringWeekdays)
		}
		// Default weekly behavior: add interval weeks
		return current.AddDate(0, 0, interval*7)

	case "monthly":
		return current.AddDate(0, interval, 0)

	case "yearly":
		return current.AddDate(interval, 0, 0)

	default:
		// Default to daily if type is unknown
		logger.Warnf("Unknown recurring type: %s, defaulting to daily", recurringType)
		return current.AddDate(0, 0, 1)
	}
}

// calculateNextWeekday finds the next occurrence for specific weekdays
// weekdays is an array of integers (0=Sunday, 1=Monday, ..., 6=Saturday)
func calculateNextWeekday(current time.Time, weekdays []int) time.Time {
	if len(weekdays) == 0 {
		return current.AddDate(0, 0, 7) // Default to next week same day
	}

	// Convert weekdays slice to map for quick lookup
	weekdaySet := make(map[int]bool)
	for _, day := range weekdays {
		weekdaySet[day] = true
	}

	// Start from next day
	next := current.AddDate(0, 0, 1)

	// Search for the next matching weekday (max 7 days)
	for i := 0; i < 7; i++ {
		currentWeekday := int(next.Weekday())
		if weekdaySet[currentWeekday] {
			return next
		}
		next = next.AddDate(0, 0, 1)
	}

	// Fallback (should never reach here if weekdays is not empty)
	return current.AddDate(0, 0, 7)
}

// findNextInCurrentPeriod finds the next date to complete in the current period
// Returns the next date, or zero time if all dates in period are completed
func findNextInCurrentPeriod(task *TodoItem, currentDate time.Time) (time.Time, bool) {
	if len(task.RecurringWeekdays) == 0 {
		return time.Time{}, false
	}

	// Build set of completed weekdays in current period
	completedDates := make(map[string]bool)
	for _, dateStr := range task.CurrentPeriodCompletions {
		completedDates[dateStr] = true
	}

	// Get current week's start (Sunday)
	weekStart := currentDate
	for weekStart.Weekday() != time.Sunday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}

	// Check each required weekday in current week
	for _, weekday := range task.RecurringWeekdays {
		targetDate := weekStart.AddDate(0, 0, weekday)
		dateStr := targetDate.Format("2006-01-02")

		// If this date is not completed and is today or in the future
		if !completedDates[dateStr] && !targetDate.Before(currentDate) {
			return targetDate, true
		}
	}

	return time.Time{}, false
}

// isPeriodCompleted checks if all required dates in the current period are completed
func isPeriodCompleted(task *TodoItem) bool {
	if len(task.RecurringWeekdays) == 0 {
		// For non-weekday tasks, each completion is a period
		return false
	}

	// Check if we have completions for all required weekdays
	return len(task.CurrentPeriodCompletions) >= len(task.RecurringWeekdays)
}
