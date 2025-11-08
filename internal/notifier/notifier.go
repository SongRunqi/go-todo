package notifier

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Notifier defines the interface for sending system notifications
type Notifier interface {
	Send(title, message string) error
}

// SystemNotifier implements cross-platform system notifications
type SystemNotifier struct{}

// NewSystemNotifier creates a new system notifier
func NewSystemNotifier() *SystemNotifier {
	return &SystemNotifier{}
}

// Send sends a system notification
func (n *SystemNotifier) Send(title, message string) error {
	switch runtime.GOOS {
	case "linux":
		return n.sendLinux(title, message)
	case "darwin":
		return n.sendDarwin(title, message)
	case "windows":
		return n.sendWindows(title, message)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// sendLinux sends notification on Linux using notify-send
func (n *SystemNotifier) sendLinux(title, message string) error {
	// Try notify-send (most common on Linux)
	cmd := exec.Command("notify-send", title, message, "-u", "normal", "-t", "5000")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send notification: %v, output: %s", err, string(output))
	}
	return nil
}

// sendDarwin sends notification on macOS using osascript
func (n *SystemNotifier) sendDarwin(title, message string) error {
	// Escape quotes in message
	message = strings.ReplaceAll(message, `"`, `\"`)
	title = strings.ReplaceAll(title, `"`, `\"`)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send notification: %v, output: %s", err, string(output))
	}
	return nil
}

// sendWindows sends notification on Windows using PowerShell
func (n *SystemNotifier) sendWindows(title, message string) error {
	// Escape quotes in message
	message = strings.ReplaceAll(message, `"`, `'`)
	title = strings.ReplaceAll(title, `"`, `'`)

	// Use PowerShell to create a toast notification
	script := fmt.Sprintf(`
		[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

		$APP_ID = 'GoTodo'

		$template = @"
<toast>
	<visual>
		<binding template="ToastText02">
			<text id="1">%s</text>
			<text id="2">%s</text>
		</binding>
	</visual>
</toast>
"@

		$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
		$xml.LoadXml($template)
		$toast = New-Object Windows.UI.Notifications.ToastNotification $xml
		[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($APP_ID).Show($toast)
	`, title, message)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send notification: %v, output: %s", err, string(output))
	}
	return nil
}

// FormatReminderMessage formats a reminder message for a task
func FormatReminderMessage(taskName string, scheduledTime string, advanceTime string) (title string, message string) {
	title = "📅 任务提醒"
	message = fmt.Sprintf("任务: %s\n时间: %s\n提前: %s", taskName, scheduledTime, advanceTime)
	return title, message
}
