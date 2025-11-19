package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/output"
)

const Cmd = `
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
<item>
<name>update</name>
<desc>user wants to update tasks, you should be careful about the update filed, if user want to update the task deadline, you should also update the endTime and the dueDate, because they are relevant
		and please keep the same format as the format list below. 请返回一个task所有的字段，只修改用户希望修改的字段，这很重要，关系到用户体验，因此，请务必小心.
</desc>
</item>
</ability>

Return format (remove markdown code fence):
{
	"intent": "create|delete|list|complete|update",
	"tasks": [
		{
			"taskId": if the user specifies some task id and user want to update the task, and note the Id is int,
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
			"status": "pending/completed/missed/skipped",
			"taskName": "CRITICAL - Use <user_preferred_language> from context: Generate the task name in the language specified in <user_preferred_language> tag. If Chinese, create Chinese task name. If English, create English task name. Extract a clear, concise title from <user_input> ,不要有任何的时间信息",
			"taskDesc": "CRITICAL - Use <user_preferred_language> from context: Generate the task description in the language specified in <user_preferred_language> tag. If Chinese, write description in Chinese. If English, write description in English. List <user_input>, make it readable, so user can know what tasks it needs todo. Keep it concise (1-2 sentences) and preserve the original meaning, only remove meaningless words",
			"dueDate": "give a clear due date, format is: yyyy-MM-dd",
			"urgent": "low, medium, high, urgent, select one, default is medium, calculate this by time left",
			"isRecurring": "true or false - Detect if this is a recurring/repeating task. Keywords: 每天 (daily), 每周 (weekly), 每月 (monthly), 每年 (yearly), daily, weekly, monthly, yearly, every day, every week, 定期 (regularly), 例行 (routine)",
			"recurringType": "Only set if isRecurring=true. Values: 'daily', 'weekly', 'monthly', 'yearly'. Examples: 每天->daily, 每周->weekly, 每月->monthly, 每年->yearly",
			"recurringInterval": "Only set if isRecurring=true. Integer for interval. Default 1. Examples: 每天->1, 每两天->2, 每周->1, 每两周->2",
			"recurringWeekdays": "Only set if isRecurring=true AND recurringType='weekly' AND task specifies specific weekdays. Array of integers where 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday. Examples: 周一周三周五->[1,3,5], 周二周四->[2,4], Mon/Wed/Fri->[1,3,5], Tue/Thu->[2,4]. Leave empty for simple weekly (every week same day).",
			"recurringMaxCount": "Only set if isRecurring=true AND user specifies a limited number of repetitions. Integer value for maximum repetitions (periods, not individual occurrences). 0 or omitted = infinite. IMPORTANT: For weekday-specific tasks, count means number of WEEKS, not individual days. Examples: 每天跑步30次->30, 每周健身12次->12, 连续10天打卡->10, 连续7周->7, 共8周->8, 连续4个月->4, daily exercise for 30 days->30, weekly meeting 12 times->12, for 12 weeks->12, Mon/Wed/Fri driving for 7 weeks->7. If no count specified, omit this field or use 0."
		}
	]
}

Note: Only include "tasks" array when intent is "create/update". For other intents, omit the tasks field or return empty array. If user intent is "update", only update the specified filed, the others keep their original values.

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
		err := store.Save(*todos, false)
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
	case "update":
		if len(intentResponse.Tasks) > 0 {

			for _, u := range intentResponse.Tasks {
				bytes, _ := json.Marshal(u)
				err := UpdateTask(todos, string(bytes), store)
				if err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}
			}
		}
	default:
		logger.Warnf("Unknown intent: %s", intentResponse.Intent)
		return fmt.Errorf("unknown intent: %s", intentResponse.Intent)
	}
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
		Tasks     []TodoItem
		PeriodKey string
		StartTime time.Time
		EndTime   time.Time
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
		for i, task := range tasks {
			taskList += fmt.Sprintf("%d. %s: %s (status: %s)\n", i+1, task.TaskName, task.TaskDesc, task.Status)
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

		// Build compact format task description with numbered list and summary
		taskDesc := ""
		for i, task := range tasks {
			taskDesc += fmt.Sprintf("%d. %s: %s\n", i+1, task.TaskName, task.TaskDesc)
		}
		taskDesc += fmt.Sprintf("\nSummary: %s", summary)

		// Create summary task with unique ID
		summaryTask := TodoItem{
			TaskID:     GetLastId(&newBackupTodos),
			CreateTime: periodData.StartTime,
			EndTime:    periodData.EndTime,
			User:       "System",
			TaskName:   taskName,
			TaskDesc:   taskDesc,
			Status:     "completed",
			DueDate:    periodKey,
			Urgent:     "low",
		}

		// Add summary task to backup immediately so next ID is unique
		newBackupTodos = append(newBackupTodos, summaryTask)
		totalCompacted += len(tasks)
	}

	// Save updated backup
	err = store.Save(newBackupTodos, true)
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
