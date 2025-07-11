package llmkit

import (
	"fmt"

	anthropic "github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/errors"
	openai "github.com/aktagon/llmkit/openai"
)

// Provider represents the LLM provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

// PromptOptions configures the prompt request
type PromptOptions struct {
	Provider     Provider // Which LLM provider to use
	SystemPrompt string   // System prompt for the request
	UserPrompt   string   // User prompt for the request
	JSONSchema   string   // Optional JSON schema for structured output
	APIKey       string   // API key for the provider
}

// Prompt sends a prompt request to the specified LLM provider
func Prompt(opts PromptOptions) (string, error) {
	if opts.APIKey == "" {
		return "", &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	switch opts.Provider {
	case ProviderOpenAI:
		return openai.Prompt(opts.SystemPrompt, opts.UserPrompt, opts.JSONSchema, opts.APIKey)
	case ProviderAnthropic:
		return anthropic.Prompt(opts.SystemPrompt, opts.UserPrompt, opts.JSONSchema, opts.APIKey)
	default:
		return "", &errors.ValidationError{
			Field:   "provider",
			Message: fmt.Sprintf("unsupported provider: %s", opts.Provider),
		}
	}
}

// PromptOpenAI is a convenience function for OpenAI prompts
func PromptOpenAI(systemPrompt, userPrompt, jsonSchema, apiKey string) (string, error) {
	return Prompt(PromptOptions{
		Provider:     ProviderOpenAI,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		JSONSchema:   jsonSchema,
		APIKey:       apiKey,
	})
}

// PromptAnthropic is a convenience function for Anthropic prompts
func PromptAnthropic(systemPrompt, userPrompt, jsonSchema, apiKey string) (string, error) {
	return Prompt(PromptOptions{
		Provider:     ProviderAnthropic,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		JSONSchema:   jsonSchema,
		APIKey:       apiKey,
	})
}
