package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/i18n"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/SongRunqi/go-todo/internal/validator"
	"github.com/SongRunqi/go-todo/parser"
)

func CreateTask(todos *[]TodoItem, todo *TodoItem) error {
	// Validate task fields
	if err := validator.ValidateTaskName(todo.TaskName); err != nil {
		return err
	}
	if err := validator.ValidateStatus("pending"); err != nil {
		return err
	}
	if todo.Urgent != "" {
		if err := validator.ValidateUrgency(todo.Urgent); err != nil {
			return err
		}
	}
	if todo.TaskDesc != "" {
		if err := validator.ValidateDescription(todo.TaskDesc); err != nil {
			return err
		}
	}
	if todo.User != "" {
		if err := validator.ValidateUser(todo.User); err != nil {
			return err
		}
	}

	// Validate recurring task fields
	if todo.IsRecurring {
		if err := validator.ValidateRecurringType(todo.RecurringType); err != nil {
			return err
		}
		if err := validator.ValidateRecurringInterval(todo.RecurringInterval, todo.IsRecurring); err != nil {
			return err
		}
		if err := validator.ValidateRecurringWeekdays(todo.RecurringWeekdays); err != nil {
			return err
		}
		if err := validator.ValidateRecurringMaxCount(todo.RecurringMaxCount, todo.IsRecurring); err != nil {
			return err
		}
		// Set default interval if not specified
		if todo.RecurringInterval == 0 {
			todo.RecurringInterval = 1
		}
		// Initialize completion count
		todo.CompletionCount = 0

		// Initialize occurrence history for recurring tasks
		todo.OccurrenceHistory = initializeOccurrenceHistory(todo)

		// Set status to "active" for recurring tasks
		todo.Status = "active"
	} else {
		// Set status to "pending" for non-recurring tasks
		todo.Status = "pending"
	}

	// Generate a unique TaskID
	id := GetLastId(todos)
	todo.TaskID = id
	// Add the new todo to the todos slice (but don't save yet)
	*todos = append(*todos, *todo)
	return nil
}

func GetLastId(todos *[]TodoItem) int {
	todoList := *todos
	length := len(todoList)
	if length < 1 {
		return 1
	}

	// Find the maximum TaskID to ensure uniqueness
	maxID := 0
	for _, todo := range todoList {
		if todo.TaskID > maxID {
			maxID = todo.TaskID
		}
	}
	return maxID + 1
}

func List(todos *[]TodoItem) error {
	newTodos := sortedList(todos)
	alfredItems := TransToAlfredItem(&newTodos)
	response := AlfredResponse{Items: *alfredItems}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func GetTask(todos *[]TodoItem, id int) error {
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			task := &(*todos)[i]
			logger.Debugf("Found task ID %d: %s", id, task.TaskName)

			// Format task as markdown
			// Only show Created and End Time if they have valid values
			createdTime := ""
			if !task.CreateTime.IsZero() {
				createdTime = task.CreateTime.Format("2006-01-02 15:04:05")
			}
			endTime := ""
			if !task.EndTime.IsZero() {
				endTime = task.EndTime.Format("2006-01-02 15:04:05")
			}

			// Build recurring task info if applicable
			recurringInfo := ""
			if task.IsRecurring {
				recurringInfo = "\n\n## 🔄 Recurring Task Details\n\n"
				recurringInfo += fmt.Sprintf("- **Type:** %s\n", task.RecurringType)
				recurringInfo += fmt.Sprintf("- **Interval:** Every %d %s\n", task.RecurringInterval, task.RecurringType)

				// Show event duration if specified
				if task.EventDuration > 0 {
					hours := int(task.EventDuration.Hours())
					minutes := int(task.EventDuration.Minutes()) % 60
					if hours > 0 && minutes > 0 {
						recurringInfo += fmt.Sprintf("- **Duration:** %dh %dm\n", hours, minutes)
					} else if hours > 0 {
						recurringInfo += fmt.Sprintf("- **Duration:** %dh\n", hours)
					} else if minutes > 0 {
						recurringInfo += fmt.Sprintf("- **Duration:** %dm\n", minutes)
					}
				}

				if len(task.RecurringWeekdays) > 0 {
					weekdayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
					weekdayNamesZh := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
					days := []string{}
					for _, wd := range task.RecurringWeekdays {
						if wd >= 0 && wd <= 6 {
							if i18n.T("field.task_id") == "Task ID" {
								days = append(days, weekdayNames[wd])
							} else {
								days = append(days, weekdayNamesZh[wd])
							}
						}
					}
					recurringInfo += fmt.Sprintf("- **Weekdays:** %s\n", strings.Join(days, ", "))
				}

				// Show occurrence history if using new model
				if len(task.OccurrenceHistory) > 0 {
					now := time.Now()
					weekStart := now
					for weekStart.Weekday() != time.Sunday {
						weekStart = weekStart.AddDate(0, 0, -1)
					}
					weekEnd := weekStart.AddDate(0, 0, 7)

					// Count occurrences in current week
					pendingThisWeek := 0
					completedThisWeek := 0
					missedThisWeek := 0

					for _, occ := range task.OccurrenceHistory {
						if !occ.ScheduledTime.Before(weekStart) && occ.ScheduledTime.Before(weekEnd) {
							switch occ.Status {
							case "pending":
								pendingThisWeek++
							case "completed":
								completedThisWeek++
							case "missed":
								missedThisWeek++
							}
						}
					}

					// Show current week progress for weekday-specific tasks
					if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 {
						recurringInfo += fmt.Sprintf("- **Current Week:** %d completed", completedThisWeek)
						if missedThisWeek > 0 {
							recurringInfo += fmt.Sprintf(", %d missed", missedThisWeek)
						}
						if pendingThisWeek > 0 {
							recurringInfo += fmt.Sprintf(", %d pending", pendingThisWeek)
						}
						recurringInfo += fmt.Sprintf(" (out of %d)\n", len(task.RecurringWeekdays))
					}

					// Show recent completed occurrences (last 3)
					completedOccs := []OccurrenceRecord{}
					for _, occ := range task.OccurrenceHistory {
						if occ.Status == "completed" {
							completedOccs = append(completedOccs, occ)
						}
					}
					if len(completedOccs) > 0 {
						recurringInfo += "- **Recent Completions:**\n"
						start := len(completedOccs) - 3
						if start < 0 {
							start = 0
						}
						for i := len(completedOccs) - 1; i >= start && i >= 0; i-- {
							occ := completedOccs[i]
							recurringInfo += fmt.Sprintf("  - ✅ %s", occ.ScheduledTime.Format("2006-01-02 15:04"))
							if !occ.CompletedAt.IsZero() && occ.CompletedAt.Format("2006-01-02") != occ.ScheduledTime.Format("2006-01-02") {
								recurringInfo += fmt.Sprintf(" (completed on %s)", occ.CompletedAt.Format("2006-01-02"))
							}
							recurringInfo += "\n"
						}
					}

					// Show upcoming occurrences (next 3 pending)
					pendingOccs := []OccurrenceRecord{}
					for _, occ := range task.OccurrenceHistory {
						if occ.Status == "pending" {
							pendingOccs = append(pendingOccs, occ)
						}
					}
					if len(pendingOccs) > 0 {
						recurringInfo += "- **Upcoming:**\n"
						count := 3
						if len(pendingOccs) < count {
							count = len(pendingOccs)
						}
						for i := 0; i < count; i++ {
							occ := pendingOccs[i]
							recurringInfo += fmt.Sprintf("  - 📅 %s", occ.ScheduledTime.Format("2006-01-02 15:04"))
							if task.EventDuration > 0 {
								endTime := occ.ScheduledTime.Add(task.EventDuration)
								recurringInfo += fmt.Sprintf(" - %s", endTime.Format("15:04"))
							}
							recurringInfo += "\n"
						}
					}

					// Show missed occurrences if any
					missedOccs := []OccurrenceRecord{}
					for _, occ := range task.OccurrenceHistory {
						if occ.Status == "missed" {
							missedOccs = append(missedOccs, occ)
						}
					}
					if len(missedOccs) > 0 {
						recurringInfo += fmt.Sprintf("- **Missed:** %d occurrence(s)\n", len(missedOccs))
					}
				} else {
					// Legacy format - show old progress tracking
					if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 && len(task.CurrentPeriodCompletions) > 0 {
						periodProgress := fmt.Sprintf("%d/%d", len(task.CurrentPeriodCompletions), len(task.RecurringWeekdays))
						recurringInfo += fmt.Sprintf("- **Current Week Progress:** %s days completed\n", periodProgress)
						recurringInfo += "- **Completed This Week:** " + strings.Join(task.CurrentPeriodCompletions, ", ") + "\n"
					}
				}

				// Show total progress
				if task.RecurringMaxCount > 0 {
					recurringInfo += fmt.Sprintf("- **Total Progress:** %d/%d periods completed\n", task.CompletionCount, task.RecurringMaxCount)
					remaining := task.RecurringMaxCount - task.CompletionCount
					recurringInfo += fmt.Sprintf("- **Remaining:** %d periods\n", remaining)
				} else {
					if task.CompletionCount > 0 {
						recurringInfo += fmt.Sprintf("- **Total Completed:** %d periods\n", task.CompletionCount)
					}
					recurringInfo += "- **Max Count:** Infinite ♾️\n"
				}
			}

			md := fmt.Sprintf(`# %s

- **%s:** %d
- **%s:** %s
- **%s:** %s
- **%s:** %s
- **%s:** %s
- **%s:** %s%s%s%s

## %s

%s

---

**%s:** %s`,
				task.TaskName,
				i18n.T("field.task_id"), task.TaskID,
				i18n.T("field.task_name"), task.TaskName,
				i18n.T("field.status"), task.Status,
				i18n.T("field.user"), task.User,
				i18n.T("field.due_date"), task.DueDate,
				i18n.T("field.urgency"), task.Urgent,
				func() string {
					if createdTime != "" {
						return "\n- **" + i18n.T("field.created") + ":** " + createdTime
					}
					return ""
				}(),
				func() string {
					if endTime != "" {
						return "\n- **" + i18n.T("field.end_time") + ":** " + endTime
					}
					return ""
				}(),
				recurringInfo,
				i18n.T("field.description"),
				task.TaskDesc,
				i18n.T("field.tips"), i18n.T("tip.edit_markdown"))

			fmt.Println(md)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

func UpdateTask(todos *[]TodoItem, todoMD string, store *FileTodoStore) error {
	logger.Debugf("Updating task with content: %s", todoMD)

	// Parse the input using the parser package
	parsedTask, err := parser.Parse(todoMD)
	if err != nil {
		return fmt.Errorf("failed to parse task update: %w", err)
	}

	// Normalize status first to handle Chinese/English variations
	normalizedStatus := validator.NormalizeStatus(parsedTask.Status)

	// Convert parser.TodoItem to main.TodoItem
	updatedTask := TodoItem{
		TaskID:     parsedTask.TaskID,
		CreateTime: parsedTask.CreateTime,
		EndTime:    parsedTask.EndTime,
		User:       parsedTask.User,
		TaskName:   parsedTask.TaskName,
		TaskDesc:   parsedTask.TaskDesc,
		Status:     normalizedStatus,
		DueDate:    parsedTask.DueDate,
		Urgent:     parsedTask.Urgent,
	}

	// Validate task ID
	if err := validator.ValidateTaskID(updatedTask.TaskID); err != nil {
		return err
	}

	// Validate other fields
	if err := validator.ValidateTaskName(updatedTask.TaskName); err != nil {
		return err
	}
	if err := validator.ValidateStatus(updatedTask.Status); err != nil {
		return err
	}
	if updatedTask.Urgent != "" {
		if err := validator.ValidateUrgency(updatedTask.Urgent); err != nil {
			return err
		}
	}
	if updatedTask.TaskDesc != "" {
		if err := validator.ValidateDescription(updatedTask.TaskDesc); err != nil {
			return err
		}
	}
	if updatedTask.User != "" {
		if err := validator.ValidateUser(updatedTask.User); err != nil {
			return err
		}
	}

	// Find and update the task
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == updatedTask.TaskID {
			logger.Debugf("Updating task ID %d: %s", updatedTask.TaskID, updatedTask.TaskName)

			// Preserve CreateTime and EndTime from original task if not provided
			if updatedTask.CreateTime.IsZero() {
				updatedTask.CreateTime = (*todos)[i].CreateTime
			}
			if updatedTask.EndTime.IsZero() {
				updatedTask.EndTime = (*todos)[i].EndTime
			}

			// Update the task in place
			(*todos)[i] = updatedTask

			// Save to file
			err := store.Save(todos, false)
			if err != nil {
				return fmt.Errorf("failed to save task: %w", err)
			}

			logger.Debug("Task updated and saved successfully")
			output.PrintTaskUpdated(updatedTask.TaskID, updatedTask.TaskName)

			// Return the updated task as JSON
			data, err := json.MarshalIndent(&updatedTask, "", "  ")
			if err != nil {
				logger.ErrorWithErr(err, "Failed to marshal updated task")
			} else {
				fmt.Println(string(data))
			}
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", updatedTask.TaskID)
}

func DeleteTask(todos *[]TodoItem, id int, store *FileTodoStore) error {
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	var deletedTask *TodoItem
	taskIndex := -1
	for i := 0; i < len(*todos); i++ {
		if (*todos)[i].TaskID == id {
			deletedTask = &(*todos)[i]
			taskIndex = i
			break
		}
	}

	if deletedTask == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	taskName := deletedTask.TaskName
	logger.Debugf("Deleting task ID %d: %s", id, taskName)

	// Mark task as deleted
	deletedTask.Status = "deleted"

	// Load existing backup todos
	backupTodos, err := store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup: %w", err)
	}

	// Add deleted task to backup
	backupTodos = append(backupTodos, *deletedTask)

	// Save deleted task to backup file
	err = store.Save(&backupTodos, true)
	if err != nil {
		return fmt.Errorf("failed to save to backup: %w", err)
	}

	// Remove task from main todos
	newTodos := make([]TodoItem, 0)
	for i := 0; i < len(*todos); i++ {
		if i != taskIndex {
			newTodos = append(newTodos, (*todos)[i])
		}
	}
	*todos = newTodos

	// Save updated todos
	err = store.Save(todos, false)
	if err != nil {
		return fmt.Errorf("failed to save after deletion: %w", err)
	}

	logger.Debug("Task moved to backup with 'deleted' status")
	output.PrintTaskDeleted(id)
	return nil
}
