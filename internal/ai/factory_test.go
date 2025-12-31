package ai

import (
	"testing"
)

func TestParseProvider(t *testing.T) {
	tests := []struct {
		input    string
		expected Provider
	}{
		// DeepSeek
		{"deepseek", ProviderDeepSeek},
		{"DeepSeek", ProviderDeepSeek},
		{"DEEPSEEK", ProviderDeepSeek},
		{"deep-seek", ProviderDeepSeek},
		{"ds", ProviderDeepSeek},

		// OpenAI
		{"openai", ProviderOpenAI},
		{"OpenAI", ProviderOpenAI},
		{"OPENAI", ProviderOpenAI},
		{"open-ai", ProviderOpenAI},
		{"gpt", ProviderOpenAI},
		{"chatgpt", ProviderOpenAI},

		// Anthropic
		{"anthropic", ProviderAnthropic},
		{"Anthropic", ProviderAnthropic},
		{"ANTHROPIC", ProviderAnthropic},
		{"claude", ProviderAnthropic},
		{"anthropic-claude", ProviderAnthropic},

		// Unknown (defaults to DeepSeek)
		{"unknown", ProviderDeepSeek},
		{"", ProviderDeepSeek},
		{"  ", ProviderDeepSeek},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseProvider(tt.input)
			if result != tt.expected {
				t.Errorf("ParseProvider(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		expectError bool
	}{
		{"deepseek provider", "deepseek", false},
		{"openai provider", "openai", false},
		{"anthropic provider", "anthropic", false},
		{"claude alias", "claude", false},
		{"gpt alias", "gpt", false},
		{"unknown defaults to deepseek", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.provider, "", "test-key", "")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("Expected client, got nil")
			}
		})
	}
}

func TestNewClientTypes(t *testing.T) {
	tests := []struct {
		provider     string
		expectedType string
	}{
		{"deepseek", "*ai.DeepSeekClient"},
		{"openai", "*ai.OpenAIClient"},
		{"anthropic", "*ai.AnthropicClient"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			client, err := NewClient(tt.provider, "", "test-key", "")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			switch tt.provider {
			case "deepseek":
				if _, ok := client.(*DeepSeekClient); !ok {
					t.Errorf("Expected *DeepSeekClient, got %T", client)
				}
			case "openai":
				if _, ok := client.(*OpenAIClient); !ok {
					t.Errorf("Expected *OpenAIClient, got %T", client)
				}
			case "anthropic":
				if _, ok := client.(*AnthropicClient); !ok {
					t.Errorf("Expected *AnthropicClient, got %T", client)
				}
			}
		})
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		provider Provider
		expected string
	}{
		{ProviderDeepSeek, "deepseek-chat"},
		{ProviderOpenAI, "gpt-4o-mini"},
		{ProviderAnthropic, "claude-sonnet-4-20250514"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetDefaultModel(tt.provider)
			if result != tt.expected {
				t.Errorf("GetDefaultModel(%v) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultBaseURL(t *testing.T) {
	tests := []struct {
		provider Provider
		expected string
	}{
		{ProviderDeepSeek, "https://api.deepseek.com/chat/completions"},
		{ProviderOpenAI, "https://api.openai.com/v1/chat/completions"},
		{ProviderAnthropic, "https://api.anthropic.com/v1/messages"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetDefaultBaseURL(tt.provider)
			if result != tt.expected {
				t.Errorf("GetDefaultBaseURL(%v) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestSupportedProviders(t *testing.T) {
	providers := SupportedProviders()

	expected := []string{"deepseek", "openai", "anthropic"}

	if len(providers) != len(expected) {
		t.Errorf("Expected %d providers, got %d", len(expected), len(providers))
	}

	for i, p := range expected {
		if providers[i] != p {
			t.Errorf("Expected provider[%d] = %s, got %s", i, p, providers[i])
		}
	}
}
