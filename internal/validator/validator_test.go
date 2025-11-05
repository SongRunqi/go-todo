package validator

import (
	"strings"
	"testing"
)

func TestValidateTaskID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{"valid ID", 1, false},
		{"valid large ID", 999, false},
		{"zero ID", 0, true},
		{"negative ID", -1, true},
		{"negative large ID", -999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTaskName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "Task Name", false},
		{"valid with spaces", "  Task Name  ", false},
		{"empty string", "", true},
		{"only spaces", "   ", true},
		{"very long name", strings.Repeat("a", 201), true},
		{"max length", strings.Repeat("a", 200), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"pending", "pending", false},
		{"completed", "completed", false},
		{"invalid status", "in-progress", true},
		{"empty", "", true},
		{"uppercase", "PENDING", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUrgency(t *testing.T) {
	tests := []struct {
		name    string
		urgent  string
		wantErr bool
	}{
		{"low", "low", false},
		{"medium", "medium", false},
		{"high", "high", false},
		{"urgent", "urgent", false},
		{"empty (optional)", "", false},
		{"invalid", "super-urgent", true},
		{"uppercase", "HIGH", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUrgency(tt.urgent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUrgency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		wantErr bool
	}{
		{"short desc", "Short description", false},
		{"empty desc", "", false},
		{"max length", strings.Repeat("a", 5000), false},
		{"too long", strings.Repeat("a", 5001), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    string
		wantErr bool
	}{
		{"valid user", "john", false},
		{"empty (optional)", "", false},
		{"with spaces", "  john  ", false},
		{"max length", strings.Repeat("a", 100), false},
		{"too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAll(t *testing.T) {
	tests := []struct {
		name     string
		taskID   int
		taskName string
		taskDesc string
		status   string
		urgent   string
		user     string
		wantErr  bool
	}{
		{
			name:     "all valid",
			taskID:   1,
			taskName: "Task",
			taskDesc: "Description",
			status:   "pending",
			urgent:   "high",
			user:     "john",
			wantErr:  false,
		},
		{
			name:     "invalid task ID",
			taskID:   0,
			taskName: "Task",
			taskDesc: "Description",
			status:   "pending",
			urgent:   "high",
			user:     "john",
			wantErr:  true,
		},
		{
			name:     "invalid status",
			taskID:   1,
			taskName: "Task",
			taskDesc: "Description",
			status:   "invalid",
			urgent:   "high",
			user:     "john",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAll(tt.taskID, tt.taskName, tt.taskDesc, tt.status, tt.urgent, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
