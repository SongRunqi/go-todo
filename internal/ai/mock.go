package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// MockClient is a mock AI client for testing
type MockClient struct {
	Response string
	Error    error
	Calls    int
}

// NewMockClient creates a new mock client
func NewMockClient(response string, err error) *MockClient {
	return &MockClient{
		Response: response,
		Error:    err,
		Calls:    0,
	}
}

// Chat implements the Client interface for testing
func (m *MockClient) Chat(ctx context.Context, messages []Message) (string, error) {
	m.Calls++
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

// MockCreateTaskResponse generates a mock "create task" response
func MockCreateTaskResponse(taskName, taskDesc string) string {
	resp := map[string]interface{}{
		"intent": "create",
		"tasks": []map[string]interface{}{
			{
				"taskId":   -1,
				"taskName": taskName,
				"taskDesc": taskDesc,
				"status":   "pending",
				"urgent":   "medium",
			},
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// MockListResponse generates a mock "list" response
func MockListResponse() string {
	return `{"intent": "list"}`
}

// MockCompleteResponse generates a mock "complete" response
func MockCompleteResponse(taskID int) string {
	resp := map[string]interface{}{
		"intent": "complete",
		"tasks": []map[string]interface{}{
			{
				"taskId": taskID,
			},
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// SetResponse sets a new response for the mock client
func (m *MockClient) SetResponse(response string) {
	m.Response = response
}

// SetError sets an error for the mock client
func (m *MockClient) SetError(err error) {
	m.Error = err
}

// Reset resets the call count
func (m *MockClient) Reset() {
	m.Calls = 0
	m.Error = nil
}

// GetCallCount returns the number of times Chat was called
func (m *MockClient) GetCallCount() int {
	return m.Calls
}

// AssertCalled checks if Chat was called exactly n times
func (m *MockClient) AssertCalled(n int) error {
	if m.Calls != n {
		return fmt.Errorf("expected %d calls, got %d", n, m.Calls)
	}
	return nil
}
