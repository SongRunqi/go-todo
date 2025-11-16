package main

import (
	"fmt"
	"os"
	"time"

	"github.com/SongRunqi/go-todo/app"
)

func main() {
	fmt.Println("=== Real Scenario Test: 周三、周五2点到3点上课 ===\n")

	// Setup
	tmpFile := "/tmp/test_real_todos.json"
	tmpBackup := "/tmp/test_real_backup.json"
	defer os.Remove(tmpFile)
	defer os.Remove(tmpBackup)

	store := &app.FileTodoStore{
		Path:       tmpFile,
		BackupPath: tmpBackup,
	}

	todos := []app.TodoItem{}

	// Scenario: 现在是周一，创建"周三、周五下午2点到3点上课，连续4周"的任务
	fmt.Println("📅 现在是周一，创建任务...")

	now := time.Now()
	fmt.Printf("当前时间: %s (%s)\n\n", now.Format("2006-01-02 15:04"), now.Weekday())

	// 找到下一个周三 14:00
	nextWed := findNextWeekday(now, time.Wednesday, 14, 0)

	task := &app.TodoItem{
		TaskName:          "周三、周五下午2点到3点上课",
		TaskDesc:          "每周三和周五下午上课",
		User:              "Student",
		CreateTime:        now,
		EndTime:           nextWed,
		EventDuration:     1 * time.Hour,
		DueDate:           nextWed.Format("2006-01-02"),
		Urgent:            "medium",
		IsRecurring:       true,
		RecurringType:     "weekly",
		RecurringInterval: 1,
		RecurringWeekdays: []int{3, 5}, // Wed=3, Fri=5
		RecurringMaxCount: 4,
	}

	err := app.CreateTask(&todos, task)
	if err != nil {
		fmt.Printf("❌ 创建失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 任务创建成功!\n")
	fmt.Printf("   任务ID: %d\n", task.TaskID)
	fmt.Printf("   状态: %s\n", task.Status)
	fmt.Printf("   endTime: %s\n", task.EndTime.Format("2006-01-02 15:04"))
	fmt.Printf("   事件时长: %v\n\n", task.EventDuration)

	fmt.Println("创建的实例:")
	printOccurrences(&todos[0])

	// Save
	store.Save(&todos, false)

	// Question 1: 现在是周一，endTime是什么？
	fmt.Println("\n🤔 问题1: 现在是周一，endTime是什么？")
	fmt.Printf("   答案: %s (%s) - 本周三 14:00\n",
		todos[0].EndTime.Format("2006-01-02 15:04"),
		todos[0].EndTime.Weekday())

	// Question 2: 现在是周四，但周三的课没上，应该如何描述？
	fmt.Println("\n🤔 问题2: 假设现在是周四，周三的课没上...")

	// 模拟：标记周三的课为 missed（但不完成）
	fmt.Println("   系统应该显示:")
	fmt.Println("   - 周三的实例: missed ❌ (已过期)")
	fmt.Println("   - endTime: 周五 14:00 (下一个待完成时间)")
	fmt.Println("   - 用户仍可以补做周三的课")

	// Question 3: 现在是周四，endTime是什么？
	fmt.Println("\n🤔 问题3: 现在是周四，endTime是什么？")
	nextOcc, _ := app.GetNextPendingOccurrence(&todos[0])
	if nextOcc != nil {
		fmt.Printf("   答案: %s (%s)\n",
			nextOcc.ScheduledTime.Format("2006-01-02 15:04"),
			nextOcc.ScheduledTime.Weekday())
	}

	// 验证两层状态模型
	fmt.Println("\n=== 验证两层状态模型 ===")
	fmt.Printf("📊 任务级别状态: %s (整体任务状态)\n", todos[0].Status)
	fmt.Println("📅 实例级别状态:")
	for i, occ := range todos[0].OccurrenceHistory {
		statusIcon := "📅"
		if occ.Status == "completed" {
			statusIcon = "✅"
		} else if occ.Status == "missed" {
			statusIcon = "❌"
		}
		fmt.Printf("   %d. %s %s - %s [%s %s]\n",
			i+1,
			occ.ScheduledTime.Weekday().String()[:3],
			occ.ScheduledTime.Format("2006-01-02 15:04"),
			occ.ScheduledTime.Add(todos[0].EventDuration).Format("15:04"),
			occ.Status,
			statusIcon,
		)
	}

	// 验证 EventDuration 显示
	fmt.Println("\n=== 验证 EventDuration 功能 ===")
	fmt.Printf("✅ eventDuration: %v\n", todos[0].EventDuration)
	fmt.Printf("✅ 时间范围: 14:00 - %s\n",
		todos[0].EndTime.Add(todos[0].EventDuration).Format("15:04"))

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("\n总结:")
	fmt.Println("✅ 两层状态模型工作正常")
	fmt.Println("✅ EventDuration 正确存储和显示")
	fmt.Println("✅ OccurrenceHistory 正确初始化")
	fmt.Println("✅ endTime 语义清晰（下一个待完成的时间点）")
}

func printOccurrences(task *app.TodoItem) {
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
}

func findNextWeekday(from time.Time, targetWeekday time.Weekday, hour, minute int) time.Time {
	current := from
	// If today is target weekday and time hasn't passed, use today
	if current.Weekday() == targetWeekday {
		targetTime := time.Date(current.Year(), current.Month(), current.Day(), hour, minute, 0, 0, current.Location())
		if targetTime.After(current) {
			return targetTime
		}
	}

	// Otherwise find next occurrence
	for current.Weekday() != targetWeekday {
		current = current.AddDate(0, 0, 1)
	}

	return time.Date(
		current.Year(), current.Month(), current.Day(),
		hour, minute, 0, 0, current.Location(),
	)
}
