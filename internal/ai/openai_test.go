package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOpenAIClient(t *testing.T) {
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
			expectedURL:   "https://api.openai.com/v1/chat/completions",
			expectedModel: "gpt-4o-mini",
		},
		{
			name:          "custom values",
			baseURL:       "https://custom.openai.com/v1/chat",
			apiKey:        "custom-key",
			model:         "gpt-4",
			expectedURL:   "https://custom.openai.com/v1/chat",
			expectedModel: "gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.baseURL, tt.apiKey, tt.model)

			if client.BaseURL != tt.expectedURL {
				t.Errorf("BaseURL = %v, want %v", client.BaseURL, tt.expectedURL)
			}
			if client.Model != tt.expectedModel {
				t.Errorf("Model = %v, want %v", client.Model, tt.expectedModel)
			}
			if client.APIKey != tt.apiKey {
				t.Errorf("APIKey = %v, want %v", client.APIKey, tt.apiKey)
			}
		})
	}
}

func TestOpenAIClient_Chat(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Return mock response
		resp := Response{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: `{"intent": "list"}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "test-key", "gpt-4o-mini")

	messages := []Message{
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

func TestOpenAIClient_ChatError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "invalid-key", "gpt-4o-mini")

	messages := []Message{
		{Role: "user", Content: "list tasks"},
	}

	_, err := client.Chat(context.Background(), messages)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestOpenAIClient_ChatEmptyResponse(t *testing.T) {
	// Create a test server that returns empty choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Choices: []Choice{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "test-key", "gpt-4o-mini")

	messages := []Message{
		{Role: "user", Content: "list tasks"},
	}

	_, err := client.Chat(context.Background(), messages)
	if err == nil {
		t.Fatal("Expected error for empty response, got nil")
	}
}
