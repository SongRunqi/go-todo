package ai

import (
	"fmt"
	"strings"

	"github.com/SongRunqi/go-todo/internal/logger"
)

// Provider represents supported AI providers
type Provider string

const (
	ProviderDeepSeek  Provider = "deepseek"
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

// DefaultModels maps providers to their default models
var DefaultModels = map[Provider]string{
	ProviderDeepSeek:  "deepseek-chat",
	ProviderOpenAI:    "gpt-4o-mini",
	ProviderAnthropic: "claude-sonnet-4-20250514",
}

// DefaultBaseURLs maps providers to their default base URLs
var DefaultBaseURLs = map[Provider]string{
	ProviderDeepSeek:  "https://api.deepseek.com/chat/completions",
	ProviderOpenAI:    "https://api.openai.com/v1/chat/completions",
	ProviderAnthropic: "https://api.anthropic.com/v1/messages",
}

// NewClient creates a new AI client based on the provider
func NewClient(provider, baseURL, apiKey, model string) (Client, error) {
	p := ParseProvider(provider)

	logger.Infof("Creating AI client for provider: %s", p)

	switch p {
	case ProviderDeepSeek:
		return NewDeepSeekClient(baseURL, apiKey, model), nil
	case ProviderOpenAI:
		return NewOpenAIClient(baseURL, apiKey, model), nil
	case ProviderAnthropic:
		return NewAnthropicClient(baseURL, apiKey, model), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// ParseProvider parses a provider string and returns the Provider type
// It performs case-insensitive matching and supports common aliases
func ParseProvider(provider string) Provider {
	p := strings.ToLower(strings.TrimSpace(provider))

	switch p {
	case "deepseek", "deep-seek", "ds":
		return ProviderDeepSeek
	case "openai", "open-ai", "gpt", "chatgpt":
		return ProviderOpenAI
	case "anthropic", "claude", "anthropic-claude":
		return ProviderAnthropic
	default:
		// Default to DeepSeek for backward compatibility
		logger.Warnf("Unknown provider '%s', defaulting to DeepSeek", provider)
		return ProviderDeepSeek
	}
}

// GetDefaultModel returns the default model for a provider
func GetDefaultModel(provider Provider) string {
	if model, ok := DefaultModels[provider]; ok {
		return model
	}
	return DefaultModels[ProviderDeepSeek]
}

// GetDefaultBaseURL returns the default base URL for a provider
func GetDefaultBaseURL(provider Provider) string {
	if url, ok := DefaultBaseURLs[provider]; ok {
		return url
	}
	return DefaultBaseURLs[ProviderDeepSeek]
}

// SupportedProviders returns a list of all supported provider names
func SupportedProviders() []string {
	return []string{
		string(ProviderDeepSeek),
		string(ProviderOpenAI),
		string(ProviderAnthropic),
	}
}
