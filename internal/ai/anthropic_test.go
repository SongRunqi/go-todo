package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAnthropicClient(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		apiKey          string
		model           string
		expectedURL     string
		expectedModel   string
	}{
		{
			name:          "default values",
			baseURL:       "",
			apiKey:        "test-key",
			model:         "",
			expectedURL:   "https://api.anthropic.com/v1/messages",
			expectedModel: "claude-sonnet-4-20250514",
		},
		{
			name:          "custom values",
			baseURL:       "https://custom.anthropic.com/v1/messages",
			apiKey:        "custom-key",
			model:         "claude-3-opus-20240229",
			expectedURL:   "https://custom.anthropic.com/v1/messages",
			expectedModel: "claude-3-opus-20240229",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAnthropicClient(tt.baseURL, tt.apiKey, tt.model)

			if client.BaseURL != tt.expectedURL {
				t.Errorf("BaseURL = %v, want %v", client.BaseURL, tt.expectedURL)
			}
			if client.Model != tt.expectedModel {
				t.Errorf("Model = %v, want %v", client.Model, tt.expectedModel)
			}
			if client.APIKey != tt.apiKey {
				t.Errorf("APIKey = %v, want %v", client.APIKey, tt.apiKey)
			}
			if client.MaxTokens != 4096 {
				t.Errorf("MaxTokens = %v, want 4096", client.MaxTokens)
			}
			if client.APIVersion != "2023-06-01" {
				t.Errorf("APIVersion = %v, want '2023-06-01'", client.APIVersion)
			}
		})
	}
}

func TestAnthropicClient_Chat(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key header 'test-key', got '%s'", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version header '2023-06-01', got '%s'", r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Verify request body includes system message handling
		var req AnthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Return mock response
		resp := AnthropicResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []AnthropicContent{
				{
					Type: "text",
					Text: `{"intent": "list"}`,
				},
			},
			Model:      "claude-sonnet-4-20250514",
			StopReason: "end_turn",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "test-key", "claude-sonnet-4-20250514")

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "list tasks"},
	}

	result, err := client.Chat(context.Background(), messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"intent": "list"}`
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestAnthropicClient_ChatError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := AnthropicResponse{
			Error: &AnthropicError{
				Type:    "authentication_error",
				Message: "Invalid API key",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "invalid-key", "claude-sonnet-4-20250514")

	messages := []Message{
		{Role: "user", Content: "list tasks"},
	}

	_, err := client.Chat(context.Background(), messages)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestAnthropicClient_ChatEmptyResponse(t *testing.T) {
	// Create a test server that returns empty content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AnthropicResponse{
			ID:      "msg_123",
			Type:    "message",
			Role:    "assistant",
			Content: []AnthropicContent{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "test-key", "claude-sonnet-4-20250514")

	messages := []Message{
		{Role: "user", Content: "list tasks"},
	}

	_, err := client.Chat(context.Background(), messages)
	if err == nil {
		t.Fatal("Expected error for empty response, got nil")
	}
}

func TestAnthropicClient_SystemMessageHandling(t *testing.T) {
	var capturedRequest AnthropicRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		resp := AnthropicResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []AnthropicContent{
				{Type: "text", Text: "response"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "test-key", "claude-sonnet-4-20250514")

	messages := []Message{
		{Role: "system", Content: "You are a task parser."},
		{Role: "user", Content: "add task"},
	}

	_, err := client.Chat(context.Background(), messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify system message was extracted
	if capturedRequest.System != "You are a task parser." {
		t.Errorf("Expected system = 'You are a task parser.', got '%s'", capturedRequest.System)
	}

	// Verify only user message remains in messages array
	if len(capturedRequest.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(capturedRequest.Messages))
	}
	if capturedRequest.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", capturedRequest.Messages[0].Role)
	}
}
