package notifier

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SongRunqi/go-todo/app"
)

// ReminderService manages background reminder checking
type ReminderService struct {
	store    *app.FileTodoStore
	notifier Notifier
	ticker   *time.Ticker
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	running  bool
}

// NewReminderService creates a new reminder service
func NewReminderService(store *app.FileTodoStore, notifier Notifier) *ReminderService {
	ctx, cancel := context.WithCancel(context.Background())
	return &ReminderService{
		store:    store,
		notifier: notifier,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the reminder service with the specified check interval
func (s *ReminderService) Start(checkInterval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("reminder service is already running")
	}

	s.ticker = time.NewTicker(checkInterval)
	s.running = true

	go s.run()
	return nil
}

// Stop stops the reminder service
func (s *ReminderService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cancel()
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.running = false
}

// IsRunning returns whether the service is currently running
func (s *ReminderService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// run is the main loop that checks for reminders
func (s *ReminderService) run() {
	// Check immediately on start
	s.checkReminders()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.checkReminders()
		}
	}
}

// checkReminders checks all tasks and sends reminders as needed
func (s *ReminderService) checkReminders() {
	tasks, err := s.store.Load(false)
	if err != nil {
		// Log error but continue
		return
	}

	now := time.Now()

	for i := range tasks {
		task := &tasks[i]

		// Skip if reminders are not enabled
		if !task.Reminders.Enabled || len(task.Reminders.ReminderTimes) == 0 {
			continue
		}

		// Skip completed/cancelled tasks
		if task.Status == "completed" || task.Status == "cancelled" {
			continue
		}

		s.checkTaskReminders(task, now)
	}

	// Save updated tasks (with reminder tracking)
	s.store.Save(&tasks, false)
}

// checkTaskReminders checks and sends reminders for a specific task
func (s *ReminderService) checkTaskReminders(task *app.TodoItem, now time.Time) {
	if task.IsRecurring {
		s.checkRecurringTaskReminders(task, now)
	} else {
		s.checkSingleTaskReminders(task, now)
	}
}

// checkSingleTaskReminders handles reminders for non-recurring tasks
func (s *ReminderService) checkSingleTaskReminders(task *app.TodoItem, now time.Time) {
	// Use EndTime as the event time
	if task.EndTime.IsZero() {
		return
	}

	// Check each configured reminder time
	for _, reminderDuration := range task.Reminders.ReminderTimes {
		reminderTime := task.EndTime.Add(reminderDuration)

		// Send reminder if we're past the reminder time but before the event
		if now.After(reminderTime) && now.Before(task.EndTime) {
			// Check if we've already sent this reminder
			// For single tasks, we track at task level
			sent := false
			for _, sentDuration := range task.OccurrenceHistory {
				for _, d := range sentDuration.RemindersSent {
					if d == reminderDuration {
						sent = true
						break
					}
				}
				if sent {
					break
				}
			}

			if !sent {
				s.sendReminder(task, task.EndTime, reminderDuration)
				// Mark as sent
				if len(task.OccurrenceHistory) == 0 {
					task.OccurrenceHistory = []app.OccurrenceRecord{{
						ScheduledTime: task.EndTime,
						Status:        "pending",
						RemindersSent: []time.Duration{reminderDuration},
					}}
				} else {
					task.OccurrenceHistory[0].RemindersSent = append(
						task.OccurrenceHistory[0].RemindersSent,
						reminderDuration,
					)
				}
			}
		}
	}
}

// checkRecurringTaskReminders handles reminders for recurring tasks
func (s *ReminderService) checkRecurringTaskReminders(task *app.TodoItem, now time.Time) {
	// Check each occurrence in history
	for i := range task.OccurrenceHistory {
		occurrence := &task.OccurrenceHistory[i]

		// Skip completed, missed, or skipped occurrences
		if occurrence.Status != "pending" {
			continue
		}

		// Skip if the occurrence is in the past
		if occurrence.ScheduledTime.Before(now) {
			continue
		}

		// Check each configured reminder time
		for _, reminderDuration := range task.Reminders.ReminderTimes {
			reminderTime := occurrence.ScheduledTime.Add(reminderDuration)

			// Send reminder if we're past the reminder time but before the event
			if now.After(reminderTime) && now.Before(occurrence.ScheduledTime) {
				// Check if we've already sent this reminder for this occurrence
				alreadySent := false
				for _, sentDuration := range occurrence.RemindersSent {
					if sentDuration == reminderDuration {
						alreadySent = true
						break
					}
				}

				if !alreadySent {
					s.sendReminder(task, occurrence.ScheduledTime, reminderDuration)
					// Mark as sent
					occurrence.RemindersSent = append(occurrence.RemindersSent, reminderDuration)
				}
			}
		}
	}
}

// sendReminder sends a notification for a task
func (s *ReminderService) sendReminder(task *app.TodoItem, eventTime time.Time, advanceDuration time.Duration) {
	scheduledTimeStr := eventTime.Format("2006-01-02 15:04")
	advanceTimeStr := formatDuration(advanceDuration)

	title, message := FormatReminderMessage(task.TaskName, scheduledTimeStr, advanceTimeStr)

	err := s.notifier.Send(title, message)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to send reminder: %v\n", err)
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	d = -d // Convert negative duration to positive for display

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		if hours > 0 {
			return fmt.Sprintf("%d天%d小时", days, hours)
		}
		return fmt.Sprintf("%d天", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%d小时%d分钟", hours, minutes)
		}
		return fmt.Sprintf("%d小时", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%d分钟", minutes)
	}

	return "即将开始"
}
