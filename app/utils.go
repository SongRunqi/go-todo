package app

import (
	"sort"
	"strconv"
	"time"

	"github.com/SongRunqi/go-todo/internal/i18n"
)

func TransToAlfredItem(todos *[]TodoItem) *[]AlfredItem {
	var items = make([]AlfredItem, 0)
	for i := 0; i < len(*todos); i++ {
		task := &(*todos)[i]
		item := AlfredItem{}

		// Add recurring indicator
		recurringIndicator := ""
		if task.IsRecurring {
			recurringIndicator = "🔄 "

			// For weekday-specific recurring tasks, show period progress
			if task.RecurringType == "weekly" && len(task.RecurringWeekdays) > 0 {
				periodProgress := strconv.Itoa(len(task.CurrentPeriodCompletions)) + "/" + strconv.Itoa(len(task.RecurringWeekdays))

				// Show period count
				if task.RecurringMaxCount > 0 {
					recurringIndicator += "(" + periodProgress + " week, " + strconv.Itoa(task.CompletionCount) + "/" + strconv.Itoa(task.RecurringMaxCount) + " periods) "
				} else if task.CompletionCount > 0 {
					recurringIndicator += "(" + periodProgress + " week, " + strconv.Itoa(task.CompletionCount) + " periods) "
				} else {
					recurringIndicator += "(" + periodProgress + " this week) "
				}
			} else {
				// For other recurring types, show simple count
				if task.CompletionCount > 0 || task.RecurringMaxCount > 0 {
					// Show count/max format if max is set, otherwise just count
					if task.RecurringMaxCount > 0 {
						recurringIndicator += "(" + strconv.Itoa(task.CompletionCount) + "/" + strconv.Itoa(task.RecurringMaxCount) + ") "
					} else {
						recurringIndicator += "(" + strconv.Itoa(task.CompletionCount) + "x) "
					}
				}
			}
		}

		item.Title = "[" + strconv.Itoa(task.TaskID) + "] " + recurringIndicator + "🎯" + task.TaskName + " " + task.Urgent

		completed := task.Status == "completed"
		var prefix string = ""
		if completed {
			prefix = "✅"
		} else {
			prefix = "⌛️"
		}
		item.Subtitle = prefix + task.TaskDesc
		item.Arg = strconv.Itoa(task.TaskID)
		item.Autocomplete = task.TaskName
		items = append(items, item)
	}
	return &items
}

func sortedList(todos *[]TodoItem) []TodoItem {
	// Separate completed and non-completed tasks
	completedTasks := make([]TodoItem, 0)
	activeTasks := make([]TodoItem, 0)

	for _, task := range *todos {
		if task.Status == "completed" {
			completedTasks = append(completedTasks, task)
		} else {
			activeTasks = append(activeTasks, task)
		}
	}

	// Sort active tasks by end time
	sortedActive := sortTasksByTime(&activeTasks)

	// Sort completed tasks by end time (for consistency)
	sortedCompleted := sortTasksByTime(&completedTasks)

	// Combine: active tasks first, then completed tasks
	result := make([]TodoItem, 0)
	result = append(result, sortedActive...)
	result = append(result, sortedCompleted...)

	return result
}

// sortTasksByTime sorts tasks by their end time
func sortTasksByTime(todos *[]TodoItem) []TodoItem {
	// Use map[int64][]int to handle multiple tasks with the same end time
	score := make(map[int64][]int)
	now := time.Now().Unix()
	// assign score with task id, the less score, the higher priority
	for i, v := range *todos {
		s := v.EndTime.Unix() - now
		score[s] = append(score[s], i)
	}

	times := make([]int64, 0)
	for k := range score {
		times = append(times, k)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	var newTodos []TodoItem = make([]TodoItem, 0)
	for _, v := range times {
		// Process all tasks with the same time score
		for _, idx := range score[v] {
			item := &(*todos)[idx]
			if v < 0 {
				item.Urgent = i18n.T("time.expired")
			} else {
				days := v / 86400
				hours := (v % 86400) / 3600
				minutes := (v % 3600) / 60

				tip := ""
				if days > 0 {
					tip = tip + i18n.T("time.days", days) + " "
				} else if hours > 0 {
					tip = tip + i18n.T("time.hours", hours) + " "
				} else if minutes > 0 {
					tip = tip + i18n.T("time.minutes", minutes) + " "
				}

				if tip != "" {
					item.Urgent = i18n.T("time.remaining", tip)
				} else {
					item.Urgent = i18n.T("time.expired")
				}
			}
			newTodos = append(newTodos, (*todos)[idx])
		}
	}
	return newTodos
}
