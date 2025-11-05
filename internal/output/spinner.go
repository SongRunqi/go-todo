package output

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Spinner wraps the spinner library with our color scheme
type Spinner struct {
	s *spinner.Spinner
}

// NewSpinner creates a new spinner with default settings
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("cyan")
	return &Spinner{s: s}
}

// NewAISpinner creates a spinner specifically for AI operations
func NewAISpinner() *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = color.CyanString(" Processing with AI...")
	s.Color("cyan", "bold")
	return &Spinner{s: s}
}

// Start starts the spinner
func (sp *Spinner) Start() {
	if !NoColor {
		sp.s.Start()
	}
}

// Stop stops the spinner
func (sp *Spinner) Stop() {
	if !NoColor {
		sp.s.Stop()
	}
}

// UpdateMessage updates the spinner message
func (sp *Spinner) UpdateMessage(message string) {
	sp.s.Suffix = " " + message
}

// Success stops the spinner and shows a success message
func (sp *Spinner) Success(message string) {
	sp.Stop()
	PrintSuccess("%s", message)
}

// Error stops the spinner and shows an error message
func (sp *Spinner) Error(message string) {
	sp.Stop()
	PrintError("%s", message)
}
