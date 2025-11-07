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
			"endTime": "CRITICAL - Use START time for EVENTS, deadline time for TASKS, first occurrence time for RECURRING tasks:

			Use START time for these EVENT types (time-sensitive, must attend at specific time):
			- Meetings (会议): '3pm meeting' -> endTime=3pm START time
			- Classes/Training (课程/培训): '2pm training session' -> endTime=2pm START time
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
			"recurringInterval": "Only set if isRecurring=true. Integer for interval. Default 1. Examples: 每天->1, 每两天->2, 每周->1, 每两周->2"
		}
	]
}

Note: Only include "tasks" array when intent is "create". For other intents, omit the tasks field or return empty array.

Examples:
EVENT types (use START time):
- "明天下午3点到5点开会" -> endTime=tomorrow 3pm (meeting START)
- "周三上午10点医生预约" -> endTime=Wed 10am (appointment time)
- "下午2点培训课程" -> endTime=today 2pm (class START)
- "明天早上9点面试" -> endTime=tomorrow 9am (interview time)
- "晚上7点看电影" -> endTime=today 7pm (movie START)
- "下午4点接孩子放学" -> endTime=today 4pm (pickup time)

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

			// Check if this is a recurring task
			if task.IsRecurring {
				// Increment completion count
				task.CompletionCount++

				// Calculate next occurrence
				nextTime := calculateNextOccurrence(task.EndTime, task.RecurringType, task.RecurringInterval)
				task.EndTime = nextTime

				// Update DueDate to reflect next occurrence
				task.DueDate = nextTime.Format("2006-01-02")

				// Keep status as pending for next occurrence
				task.Status = "pending"

				// Save updated todos
				err := store.Save(todos, false)
				if err != nil {
					return fmt.Errorf("failed to save updated todos: %w", err)
				}

				logger.Infof("Recurring task completed. Count: %d, Next occurrence: %s", task.CompletionCount, nextTime.Format("2006-01-02 15:04"))
				fmt.Printf("✅ Task completed! (Count: %d) Next occurrence: %s\n", task.CompletionCount, nextTime.Format("2006-01-02 15:04"))
				return nil
			}

			// Non-recurring task: mark as completed (keep in main list)
			task.Status = "completed"

			// Save updated todos to original file
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

// calculateNextOccurrence calculates the next occurrence time based on recurring type and interval
func calculateNextOccurrence(current time.Time, recurringType string, interval int) time.Time {
	switch recurringType {
	case "daily":
		return current.AddDate(0, 0, interval)
	case "weekly":
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
		// Set default interval if not specified
		if todo.RecurringInterval == 0 {
			todo.RecurringInterval = 1
		}
		// Initialize completion count
		todo.CompletionCount = 0
	}

	// Generate a unique TaskID
	id := GetLastId(todos)
	// Set the Status field to "pending"
	todo.TaskID = id
	todo.Status = "pending"
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

			md := fmt.Sprintf(`# %s

- **%s:** %d
- **%s:** %s
- **%s:** %s
- **%s:** %s
- **%s:** %s
- **%s:** %s%s%s

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
	fmt.Println("==============================================\n")

	// Generate summaries for each period
	summaryTasks := make([]TodoItem, 0)
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

		fmt.Printf("   ✅ Generated summary: %s\n\n", title)

		// Create summary task
		summaryTask := TodoItem{
			TaskID:     0, // Will be set when added to main list if needed
			CreateTime: periodData.StartTime,
			EndTime:    periodData.EndTime,
			User:       "System",
			TaskName:   title,
			TaskDesc:   summary,
			Status:     "completed",
			DueDate:    periodKey,
			Urgent:     "low",
		}

		summaryTasks = append(summaryTasks, summaryTask)
		totalCompacted += len(tasks)
	}

	// Remove original tasks from backup and add summaries
	newBackupTodos := make([]TodoItem, 0)
	for i, task := range backupTodos {
		if !tasksToRemove[i] {
			newBackupTodos = append(newBackupTodos, task)
		}
	}

	// Add summary tasks to backup
	newBackupTodos = append(newBackupTodos, summaryTasks...)

	// Save updated backup
	err = store.Save(&newBackupTodos, true)
	if err != nil {
		return fmt.Errorf("failed to save compacted backup: %w", err)
	}

	fmt.Println("==============================================")
	fmt.Printf("✅ Successfully compacted %d tasks into %d summaries\n", totalCompacted, len(summaryTasks))
	fmt.Println("==============================================")

	logger.Infof("Compacted %d tasks into %d summaries by %s", totalCompacted, len(summaryTasks), period)
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
