package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/parser"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/validator"
	"github.com/SongRunqi/go-todo/internal/output"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

const cmd = `
<System>
You are a todo helper agent. Your task is to analyze user input and determine their intent along with any tasks they want to create.

Context Format:
You will receive user context in XML format:
<context>
	<current_time>ISO 8601 timestamp</current_time>
	<weekday>Day of the week</weekday>
	<user_preferred_language>Chinese or English</user_preferred_language>
	<user_input>The actual user input</user_input>
</context>

Key behaviors:
1. Identify the user's primary intent from the <ability> tag options based on <user_input>
2. IMPORTANT: ONLY semicolon ';' is used to separate multiple tasks. Commas (,), periods (.), and other punctuation within a sentence are NOT task separators. Only split on ';' character.
3. For a single sentence without semicolon, create ONLY ONE task regardless of commas or other punctuation
4. Return intent as a separate, independent attribute
5. Return tasks array only when user wants to create tasks (intent="create")
6. Use <current_time> to calculate task times and deadlines
7. Use <user_preferred_language> to generate taskName and taskDesc in the appropriate language

<ability>
<item>
	<name>create</name>
	<desc>user wants to create one or more tasks</desc>
</item>
<item>
	<name>delete</name>
	<desc>user wants to delete a task</desc>
</item>
<item>
	<name>list</name>
	<desc>user wants to see all the todolist</desc>
</item>
<item>
	<name>complete</name>
	<desc>user wants to complete a task</desc>
</item>
</ability>

Return format (remove markdown code fence):
{
	"intent": "create|delete|list|complete",
	"tasks": [
		{
			"taskId": -1,
			"user": "if not mentioned, You is default",
			"createTime": "use current time",
			"eventDuration": "IMPORTANT - Duration in nanoseconds for events with time ranges. Examples: '2pm-3pm' -> 3600000000000 (1 hour), '2pm-4:30pm' -> 9000000000000 (2.5 hours), '10:00-11:00' -> 3600000000000. Leave 0 or omit if no end time specified. Calculate: (end_time - start_time) in nanoseconds. 1 hour = 3600000000000ns, 1 minute = 60000000000ns",
			"endTime": "CRITICAL - Use START time for EVENTS, deadline time for TASKS, first occurrence time for RECURRING tasks:

			Use START time for these EVENT types (time-sensitive, must attend at specific time):
			- Meetings (会议): '3pm meeting' -> endTime=3pm START time, '3pm-5pm meeting' -> endTime=3pm (START not end)
			- Classes/Training (课程/培训): '2pm training session' -> endTime=2pm START time, '2pm-4pm training' -> endTime=2pm
			- Driving (开车): '3:00到5:00开车' -> endTime=3:00 (START time)
			- Appointments (预约): 'doctor appointment at 10am' -> endTime=10am
			- Interviews (面试): 'job interview at 9am' -> endTime=9am
			- Transportation (交通): 'flight departs 8am', 'train at 3pm' -> use departure time
			- Entertainment (娱乐): 'movie at 7pm', 'concert at 8pm' -> use start time
			- Exams (考试): 'exam starts at 2pm' -> endTime=2pm
			- Social events (社交): 'dinner at 6pm', 'party at 8pm' -> use start time
			- Live events (直播): 'webinar starts 3pm' -> endTime=3pm
			- Pick up/Drop off (接送): 'pick up kids at 4pm' -> endTime=4pm

			Use DEADLINE time for TASK types (flexible, can complete anytime before deadline):
			- Reports/Documents: 'submit report by Friday' -> use Friday as deadline
			- Projects: 'finish project by month end' -> use deadline
			- General todos: 'buy groceries' -> estimate reasonable deadline

			Key rule: If it has a specific time range (e.g., '3pm-5pm'), it's an EVENT -> use START time (3pm).
			If it only mentions 'by/before date', it's a TASK -> use deadline.",
			"taskName": "CRITICAL - Use <user_preferred_language> from context: Generate the task name in the language specified in <user_preferred_language> tag. If Chinese, create Chinese task name. If English, create English task name. Extract a clear, concise title from <user_input> without adding creative interpretations.",
			"taskDesc": "CRITICAL - Use <user_preferred_language> from context: Generate the task description in the language specified in <user_preferred_language> tag. If Chinese, write description in Chinese. If English, write description in English. Summarize <user_input> directly and factually. Keep it concise (1-2 sentences) and preserve the original meaning.",
			"dueDate": "give a clear due date",
			"urgent": "low, medium, high, urgent, select one, default is medium, calculate this by time left",
			"isRecurring": "true or false - Detect if this is a recurring/repeating task. Keywords: 每天 (daily), 每周 (weekly), 每月 (monthly), 每年 (yearly), daily, weekly, monthly, yearly, every day, every week, 定期 (regularly), 例行 (routine)",
			"recurringType": "Only set if isRecurring=true. Values: 'daily', 'weekly', 'monthly', 'yearly'. Examples: 每天->daily, 每周->weekly, 每月->monthly, 每年->yearly",
			"recurringInterval": "Only set if isRecurring=true. Integer for interval. Default 1. Examples: 每天->1, 每两天->2, 每周->1, 每两周->2",
			"recurringWeekdays": "Only set if isRecurring=true AND recurringType='weekly' AND task specifies specific weekdays. Array of integers where 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday. Examples: 周一周三周五->[1,3,5], 周二周四->[2,4], Mon/Wed/Fri->[1,3,5], Tue/Thu->[2,4]. Leave empty for simple weekly (every week same day).",
			"recurringMaxCount": "Only set if isRecurring=true AND user specifies a limited number of repetitions. Integer value for maximum repetitions (periods, not individual occurrences). 0 or omitted = infinite. IMPORTANT: For weekday-specific tasks, count means number of WEEKS, not individual days. Examples: 每天跑步30次->30, 每周健身12次->12, 连续10天打卡->10, 连续7周->7, 共8周->8, 连续4个月->4, daily exercise for 30 days->30, weekly meeting 12 times->12, for 12 weeks->12, Mon/Wed/Fri driving for 7 weeks->7. If no count specified, omit this field or use 0."
		}
	]
}

Note: Only include "tasks" array when intent is "create". For other intents, omit the tasks field or return empty array.

Examples:
EVENT types (use START time + eventDuration):
- "明天下午3点到5点开会" -> endTime=tomorrow 3pm (meeting START, not 5pm), eventDuration=7200000000000 (2 hours)
- "周三上午10点医生预约" -> endTime=Wed 10am (appointment time), eventDuration=0 (no end time specified)
- "下午2点到3点培训课程" -> endTime=today 2pm (class START), eventDuration=3600000000000 (1 hour)
- "明天早上9点面试" -> endTime=tomorrow 9am (interview time), eventDuration=0
- "晚上7点到9点看电影" -> endTime=today 7pm (movie START), eventDuration=7200000000000 (2 hours)
- "下午4点接孩子放学" -> endTime=today 4pm (pickup time), eventDuration=0
- "周一、周三、周五 3:00到5:00开车" -> isRecurring=true, recurringWeekdays=[1,3,5], endTime=next Monday 3:00 (START time), eventDuration=7200000000000 (2 hours)
- "周三、周五下午2点到3点上课" -> isRecurring=true, recurringWeekdays=[3,5], endTime=next Wed 2pm, eventDuration=3600000000000 (1 hour)

TASK types (use DEADLINE):
- "周五前提交报告" -> endTime=Friday end of day (deadline)
- "买牛奶，面包，鸡蛋" -> ONE task, estimate reasonable deadline
- "月底前完成项目" -> endTime=end of month (deadline)

Separator examples:
- "买牛奶，面包，鸡蛋" -> ONE task (commas are content)
- "买牛奶; 写报告; 开会" -> THREE tasks (semicolon separates)

RECURRING task examples:
- "每天早上9点站会" -> isRecurring=true, recurringType="daily", recurringInterval=1, endTime=tomorrow 9am
- "每周一下午2点周会" -> isRecurring=true, recurringType="weekly", recurringInterval=1, endTime=next Monday 2pm
- "每两周写周报" -> isRecurring=true, recurringType="weekly", recurringInterval=2
- "每月1号交房租" -> isRecurring=true, recurringType="monthly", recurringInterval=1
- "daily standup at 9am" -> isRecurring=true, recurringType="daily", recurringInterval=1
- "weekly report every Friday" -> isRecurring=true, recurringType="weekly", recurringInterval=1
- "周一、周三、周五去上课" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[1,3,5], endTime=next matching day
- "周二周四晚上健身" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[2,4]
- "Mon/Wed/Fri team meeting" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[1,3,5]
- "Tuesday and Thursday gym" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[2,4]
- "每天跑步30次" -> isRecurring=true, recurringType="daily", recurringInterval=1, recurringMaxCount=30
- "每周健身12次" -> isRecurring=true, recurringType="weekly", recurringInterval=1, recurringMaxCount=12
- "连续10天打卡" -> isRecurring=true, recurringType="daily", recurringInterval=1, recurringMaxCount=10
- "daily exercise for 30 days" -> isRecurring=true, recurringType="daily", recurringInterval=1, recurringMaxCount=30
- "weekly meeting 12 times" -> isRecurring=true, recurringType="weekly", recurringInterval=1, recurringMaxCount=12

COMPLEX recurring task examples (combining weekdays + time + count + duration):
- "周一、周三、周五 3:00到5:00开车，连续7周" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[1,3,5], recurringMaxCount=7, endTime=next Monday 3:00 (START time), eventDuration=7200000000000 (2 hours)
- "周二周四上午10点到11点培训，共8周" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[2,4], recurringMaxCount=8, endTime=next matching day 10:00, eventDuration=3600000000000 (1 hour)
- "Mon/Wed/Fri 2pm-4pm team meeting, 12 weeks" -> isRecurring=true, recurringType="weekly", recurringWeekdays=[1,3,5], recurringMaxCount=12, endTime=next Monday 2pm, eventDuration=7200000000000 (2 hours)
- "连续4个月每月1号交房租" -> isRecurring=true, recurringType="monthly", recurringInterval=1, recurringMaxCount=4
- "连续6周每周五写周报" -> isRecurring=true, recurringType="weekly", recurringInterval=1, recurringMaxCount=6

Pattern recognition for "连续X周/月/年" (consecutive periods):
- "连续7周" = recurringMaxCount=7, recurringType="weekly"
- "连续10天" = recurringMaxCount=10, recurringType="daily"
- "连续4个月" = recurringMaxCount=4, recurringType="monthly"
- "连续2年" = recurringMaxCount=2, recurringType="yearly"
- "共8周" = recurringMaxCount=8, recurringType="weekly"
- "for 12 weeks" = recurringMaxCount=12, recurringType="weekly"
- "for 30 days" = recurringMaxCount=30, recurringType="daily"

- "例行检查设备" (without specific frequency) -> isRecurring=false (not specific enough)
- "买牛奶" (one-time task) -> isRecurring=false

Language preference examples with XML context:

Example 1 - Chinese user with English input:
Input context:
<context>
	<current_time>2025-01-15T10:00:00Z</current_time>
	<weekday>Monday</weekday>
	<user_preferred_language>Chinese</user_preferred_language>
	<user_input>meeting tomorrow at 3pm</user_input>
</context>
Expected output: taskName: "明天下午3点开会", taskDesc: "明天下午3点参加会议"

Example 2 - English user with Chinese input:
Input context:
<context>
	<current_time>2025-01-15T10:00:00Z</current_time>
	<weekday>Monday</weekday>
	<user_preferred_language>English</user_preferred_language>
	<user_input>明天下午3点开会</user_input>
</context>
Expected output: taskName: "Meeting Tomorrow at 3 PM", taskDesc: "Attend meeting tomorrow at 3 PM"

`

func DoI(todoStr string, todos *[]TodoItem, store *FileTodoStore) error {

	var intentResponse IntentResponse
	removedata := removeJsonTag(todoStr)
	err := json.Unmarshal([]byte(removedata), &intentResponse)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to parse intent response")
		return fmt.Errorf("failed to parse intent response: %w", err)
	}

	logger.Infof("Intent: %s, Number of tasks: %d", intentResponse.Intent, len(intentResponse.Tasks))

	switch intentResponse.Intent {
	case "create":
		// Handle multiple tasks separated by semicolons
		for i := range intentResponse.Tasks {
			task := &intentResponse.Tasks[i]
			if err := CreateTask(todos, task); err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}
			output.PrintTaskCreated(task.TaskID, task.TaskName)
		}
		// Save all tasks at once after creating them
		err := store.Save(todos, false)
		if err != nil {
			return fmt.Errorf("failed to save todos batch: %w", err)
		}
	case "list":
		if err := List(todos); err != nil {
			return fmt.Errorf("failed to list todos: %w", err)
		}
	case "complete":
		// For complete and delete, we might need additional logic
		// to extract task ID from the user input or tasks array
		if len(intentResponse.Tasks) > 0 {
			if err := Complete(todos, &intentResponse.Tasks[0], store); err != nil {
				return fmt.Errorf("failed to complete task: %w", err)
			}
		}
	case "delete":
		if len(intentResponse.Tasks) > 0 {
			if err := DeleteTask(todos, intentResponse.Tasks[0].TaskID, store); err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}
		}
	default:
		logger.Warnf("Unknown intent: %s", intentResponse.Intent)
		return fmt.Errorf("unknown intent: %s", intentResponse.Intent)
	}
	return nil
}

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

func RestoreTask(todos *[]TodoItem, backupTodos *[]TodoItem, id int, store *FileTodoStore) error {
	if err := validator.ValidateTaskID(id); err != nil {
		return err
	}

	// Find the task in backup
	var taskToRestore *TodoItem
	var backupIndex int = -1
	for i := 0; i < len(*backupTodos); i++ {
		if (*backupTodos)[i].TaskID == id {
			taskToRestore = &(*backupTodos)[i]
			backupIndex = i
			break
		}
	}

	if taskToRestore == nil {
		return fmt.Errorf("task with ID %d not found in backup", id)
	}

	logger.Debugf("Found task to restore - ID %d: %s", id, taskToRestore.TaskName)

	// Change status back to pending
	restoredTask := *taskToRestore
	restoredTask.Status = "pending"

	// Add to active todos
	*todos = append(*todos, restoredTask)

	// Save updated active todos
	err := store.Save(todos, false)
	if err != nil {
		return fmt.Errorf("failed to save active todos: %w", err)
	}

	// Remove from backup
	newBackupTodos := make([]TodoItem, 0)
	for i := 0; i < len(*backupTodos); i++ {
		if i != backupIndex {
			newBackupTodos = append(newBackupTodos, (*backupTodos)[i])
		}
	}
	*backupTodos = newBackupTodos

	// Save updated backup
	err = store.Save(backupTodos, true)
	if err != nil {
		return fmt.Errorf("failed to update backup: %w", err)
	}

	logger.Debug("Task restored successfully")
	output.PrintTaskRestored(id, restoredTask.TaskName)
	return nil
}

func CopyCompletedTasks(todos *[]TodoItem, store *FileTodoStore, weekOnly bool) error {
	// Collect completed tasks from both main list and backup
	completedTasks := make([]TodoItem, 0)

	// Get completed tasks from main list
	for _, task := range *todos {
		if task.Status == "completed" {
			completedTasks = append(completedTasks, task)
		}
	}

	// Get completed tasks from backup
	backupTodos, err := store.Load(true)
	if err != nil {
		logger.Warnf("Failed to load backup todos: %v", err)
	} else {
		for _, task := range backupTodos {
			if task.Status == "completed" {
				completedTasks = append(completedTasks, task)
			}
		}
	}

	if len(completedTasks) == 0 {
		fmt.Println("No completed tasks found")
		return nil
	}

	// Group tasks by week
	tasksByWeek := make(map[string][]string)
	now := time.Now()

	for _, task := range completedTasks {
		// Use EndTime to determine the week
		year, week := task.EndTime.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		// If weekOnly is true, only include current week
		if weekOnly {
			currentYear, currentWeek := now.ISOWeek()
			if year != currentYear || week != currentWeek {
				continue
			}
		}

		if _, exists := tasksByWeek[weekKey]; !exists {
			tasksByWeek[weekKey] = make([]string, 0)
		}
		tasksByWeek[weekKey] = append(tasksByWeek[weekKey], task.TaskName)
	}

	if len(tasksByWeek) == 0 {
		fmt.Println("No completed tasks found for the specified time period")
		return nil
	}

	// Sort weeks
	weeks := make([]string, 0, len(tasksByWeek))
	for week := range tasksByWeek {
		weeks = append(weeks, week)
	}
	sort.Strings(weeks)

	// Format output
	output := ""
	for _, week := range weeks {
		output += fmt.Sprintf("=== %s ===\n", week)
		for i, taskName := range tasksByWeek[week] {
			output += fmt.Sprintf("%d. %s\n", i+1, taskName)
		}
		output += "\n"
	}

	// Print to stdout (can be piped to clipboard tools like pbcopy or xclip)
	fmt.Print(output)

	logger.Infof("Copied %d completed tasks from %d week(s)", len(completedTasks), len(weeks))
	return nil
}

func CompactTasks(store *FileTodoStore, period string) error {
	// Validate period
	if period != "week" && period != "month" {
		return fmt.Errorf("invalid period: %s (must be 'week' or 'month')", period)
	}

	// Load backup tasks
	backupTodos, err := store.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load backup: %w", err)
	}

	// Group tasks by period
	type PeriodTasks struct {
		Tasks      []TodoItem
		PeriodKey  string
		StartTime  time.Time
		EndTime    time.Time
	}

	tasksByPeriod := make(map[string]*PeriodTasks)
	tasksToRemove := make(map[int]bool) // Track task indices to remove

	// Group completed and deleted tasks by period
	for i, task := range backupTodos {
		if task.Status != "completed" && task.Status != "deleted" {
			continue
		}

		periodKey := getPeriodKey(task.EndTime, period)
		if _, exists := tasksByPeriod[periodKey]; !exists {
			tasksByPeriod[periodKey] = &PeriodTasks{
				Tasks:     make([]TodoItem, 0),
				PeriodKey: periodKey,
			}
		}
		tasksByPeriod[periodKey].Tasks = append(tasksByPeriod[periodKey].Tasks, task)
		tasksToRemove[i] = true

		// Track time range
		if tasksByPeriod[periodKey].StartTime.IsZero() || task.EndTime.Before(tasksByPeriod[periodKey].StartTime) {
			tasksByPeriod[periodKey].StartTime = task.EndTime
		}
		if tasksByPeriod[periodKey].EndTime.IsZero() || task.EndTime.After(tasksByPeriod[periodKey].EndTime) {
			tasksByPeriod[periodKey].EndTime = task.EndTime
		}
	}

	if len(tasksByPeriod) == 0 {
		fmt.Println("No completed or deleted tasks found in backup")
		return nil
	}

	// Sort periods
	periods := make([]string, 0, len(tasksByPeriod))
	for p := range tasksByPeriod {
		periods = append(periods, p)
	}
	sort.Strings(periods)

	fmt.Println("==============================================")
	fmt.Printf("Compacting tasks by %s using AI...\n", period)
	fmt.Println("==============================================")

	// Remove original tasks from backup first to prepare the base list
	newBackupTodos := make([]TodoItem, 0)
	for i, task := range backupTodos {
		if !tasksToRemove[i] {
			newBackupTodos = append(newBackupTodos, task)
		}
	}

	// Generate summaries for each period and assign unique IDs
	totalCompacted := 0

	for _, periodKey := range periods {
		periodData := tasksByPeriod[periodKey]
		tasks := periodData.Tasks

		fmt.Printf("📅 Processing %s (%d tasks)...\n", periodKey, len(tasks))

		// Prepare task list for AI
		taskList := ""
		completedCount := 0
		deletedCount := 0
		for _, task := range tasks {
			taskList += fmt.Sprintf("- %s (status: %s)\n", task.TaskName, task.Status)
			if task.Status == "completed" {
				completedCount++
			} else if task.Status == "deleted" {
				deletedCount++
			}
		}

		// Call AI to generate summary
		prompt := fmt.Sprintf(`Please create a concise and friendly summary for the following tasks from %s.

Time period: %s
Total tasks: %d (completed: %d, deleted: %d)

Tasks:
%s

Please provide:
1. A brief title/name for this period (max 50 characters)
2. A friendly summary paragraph (2-3 sentences) describing the main accomplishments and activities

Format your response as:
Title: [your title here]
Summary: [your summary here]`, periodKey, periodKey, len(tasks), completedCount, deletedCount, taskList)

		req := OpenAIRequest{
			Model: "deepseek-chat",
			Messages: []Msg{
				{Role: "user", Content: prompt},
			},
		}

		// Show spinner during AI request
		spin := output.NewAISpinner()
		spin.Start()

		aiResponse, err := Chat(req)
		spin.Stop()

		if err != nil {
			logger.Warnf("Failed to generate AI summary for %s: %v", periodKey, err)
			// Create a simple summary without AI
			aiResponse = fmt.Sprintf("Title: Tasks Summary - %s\nSummary: Completed %d tasks and managed %d items during this period.",
				periodKey, completedCount, len(tasks))
		}

		// Parse AI response
		title, summary := parseAISummaryResponse(aiResponse)
		if title == "" {
			title = fmt.Sprintf("Tasks Summary - %s", periodKey)
		}
		if summary == "" {
			summary = fmt.Sprintf("Period: %s. Completed %d tasks, deleted %d tasks.", periodKey, completedCount, deletedCount)
		}

		// Format task name with AI-generated title and date range
		taskName := fmt.Sprintf("%s %s ~ %s",
			title,
			periodData.StartTime.Format("1.2"),
			periodData.EndTime.Format("1.2"))

		fmt.Printf("   ✅ Generated task name: %s\n\n", taskName)

		// Create summary task with unique ID
		summaryTask := TodoItem{
			TaskID:     GetLastId(&newBackupTodos),
			CreateTime: periodData.StartTime,
			EndTime:    periodData.EndTime,
			User:       "System",
			TaskName:   taskName,
			TaskDesc:   summary,
			Status:     "completed",
			DueDate:    periodKey,
			Urgent:     "low",
		}

		// Add summary task to backup immediately so next ID is unique
		newBackupTodos = append(newBackupTodos, summaryTask)
		totalCompacted += len(tasks)
	}

	// Save updated backup
	err = store.Save(&newBackupTodos, true)
	if err != nil {
		return fmt.Errorf("failed to save compacted backup: %w", err)
	}

	summaryCount := len(periods)
	fmt.Println("==============================================")
	fmt.Printf("✅ Successfully compacted %d tasks into %d summaries\n", totalCompacted, summaryCount)
	fmt.Println("==============================================")

	logger.Infof("Compacted %d tasks into %d summaries by %s", totalCompacted, summaryCount, period)
	return nil
}

// parseAISummaryResponse extracts title and summary from AI response
func parseAISummaryResponse(response string) (title, summary string) {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		} else if strings.HasPrefix(line, "Summary:") {
			summary = strings.TrimSpace(strings.TrimPrefix(line, "Summary:"))
		}
	}
	return title, summary
}

func getPeriodKey(t time.Time, period string) string {
	if period == "week" {
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	} else { // month
		return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
	}
}
