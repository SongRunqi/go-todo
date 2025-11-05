package ai

import (
	"context"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client interface for AI providers
type Client interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}

// Response represents AI response structure
type Response struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}
