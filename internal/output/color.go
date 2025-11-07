package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/SongRunqi/go-todo/internal/i18n"
)

var (
	// Success colors (green)
	Success = color.New(color.FgGreen)
	SuccessBold = color.New(color.FgGreen, color.Bold)

	// Error colors (red)
	Error = color.New(color.FgRed)
	ErrorBold = color.New(color.FgRed, color.Bold)

	// Warning colors (yellow)
	Warning = color.New(color.FgYellow)
	WarningBold = color.New(color.FgYellow, color.Bold)

	// Info colors (blue)
	Info = color.New(color.FgCyan)
	InfoBold = color.New(color.FgCyan, color.Bold)

	// Task title colors
	TaskTitle = color.New(color.FgCyan, color.Bold)
	TaskID = color.New(color.FgMagenta, color.Bold)

	// Disabled flag for testing
	NoColor = false
)

func init() {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
		NoColor = true
	}
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Success.Printf("✓ %s\n", msg)
}

// PrintError prints an error message with X mark
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Error.Fprintf(os.Stderr, "✗ %s\n", msg)
}

// PrintWarning prints a warning message with warning sign
func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Warning.Printf("⚠ %s\n", msg)
}

// PrintInfo prints an info message with info icon
func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Info.Printf("ℹ %s\n", msg)
}

// PrintTaskCreated prints a formatted task creation message
func PrintTaskCreated(taskID int, taskName string) {
	Success.Printf("✓ %s: ", i18n.T("output.task_created"))
	TaskID.Printf("#%d ", taskID)
	TaskTitle.Println(taskName)
}

// PrintTaskCompleted prints a formatted task completion message
func PrintTaskCompleted(taskID int, taskName string) {
	Success.Printf("✓ %s: ", i18n.T("output.task_completed"))
	TaskID.Printf("#%d ", taskID)
	TaskTitle.Println(taskName)
}

// PrintTaskUpdated prints a formatted task update message
func PrintTaskUpdated(taskID int, taskName string) {
	Success.Printf("✓ %s: ", i18n.T("output.task_updated"))
	TaskID.Printf("#%d ", taskID)
	TaskTitle.Println(taskName)
}

// PrintTaskDeleted prints a formatted task deletion message
func PrintTaskDeleted(taskID int) {
	Success.Printf("✓ %s: #%d\n", i18n.T("output.task_deleted"), taskID)
}

// PrintTaskRestored prints a formatted task restoration message
func PrintTaskRestored(taskID int, taskName string) {
	Success.Printf("✓ %s: ", i18n.T("output.task_restored"))
	TaskID.Printf("#%d ", taskID)
	TaskTitle.Println(taskName)
}

// PrintErrorWithSuggestion prints an error with an actionable suggestion
func PrintErrorWithSuggestion(errorMsg string, suggestion string) {
	Error.Fprintf(os.Stderr, "✗ Error: %s\n", errorMsg)
	if suggestion != "" {
		Info.Fprintf(os.Stderr, "  💡 Suggestion: %s\n", suggestion)
	}
}

// PrintUsageExample prints a usage example
func PrintUsageExample(command string, description string) {
	Info.Printf("  Example: ")
	color.New(color.FgWhite, color.Bold).Printf("%s", command)
	fmt.Printf(" - %s\n", description)
}
