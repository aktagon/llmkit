package llmkit

import (
	"fmt"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/google"
	"github.com/aktagon/llmkit/internal"
	"github.com/aktagon/llmkit/openai"
)

// ConversationalAgent represents a conversational agent
type ConversationalAgent interface {
	Chat(message string, opts ...interface{}) (interface{}, error)
	RegisterTool(tool interface{}) error
	Reset(clearMemory bool) error
	Remember(key, value string) error
	Recall(key string) (string, bool)
}

// Agent creates a new conversational agent for the specified provider
func Agent(provider Provider, apiKey string) (ConversationalAgent, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	var internalProvider internal.Provider

	switch provider {
	case ProviderOpenAI:
		internalProvider = openai.NewProvider()
	case ProviderAnthropic:
		internalProvider = anthropic.NewProvider()
	case ProviderGoogle:
		internalProvider = google.NewProvider()
	default:
		return nil, &errors.ValidationError{
			Field:   "provider",
			Message: fmt.Sprintf("unsupported provider: %s", provider),
		}
	}

	agent, err := internalProvider.Agent(apiKey)
	if err != nil {
		return nil, err
	}

	return agent.(ConversationalAgent), nil
}