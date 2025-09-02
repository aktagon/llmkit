package llmkit

import (
	"context"
	"fmt"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/google"
	"github.com/aktagon/llmkit/internal"
	"github.com/aktagon/llmkit/openai"
)

// Provider represents the LLM provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
)

// PromptOptions configures the prompt request
type PromptOptions struct {
	Provider     Provider        // Which LLM provider to use
	SystemPrompt string          // System prompt for the request
	UserPrompt   string          // User prompt for the request
	JSONSchema   string          // Optional JSON schema for structured output
	APIKey       string          // API key for the provider
	MaxTokens    int             // Maximum tokens in response (0 = omit from request)
	Temperature  float64         // Sampling temperature (0 = omit from request)
	Files        []internal.File // Optional file attachments
}

// Prompt sends a prompt request to the specified LLM provider
func Prompt(opts PromptOptions) (string, error) {
	if opts.APIKey == "" {
		return "", &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	ctx := context.Background()
	var provider internal.Provider

	switch opts.Provider {
	case ProviderOpenAI:
		provider = openai.NewProvider()
	case ProviderAnthropic:
		provider = anthropic.NewProvider()
	case ProviderGoogle:
		provider = google.NewProvider()
	default:
		return "", &errors.ValidationError{
			Field:   "provider",
			Message: fmt.Sprintf("unsupported provider: %s", opts.Provider),
		}
	}

	settings := internal.Settings{
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	}
	return provider.Prompt(ctx, opts.SystemPrompt, opts.UserPrompt, opts.JSONSchema, opts.APIKey, settings, opts.Files...)
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

// PromptGoogle is a convenience function for Google prompts
func PromptGoogle(systemPrompt, userPrompt, jsonSchema, apiKey string) (string, error) {
	return Prompt(PromptOptions{
		Provider:     ProviderGoogle,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		JSONSchema:   jsonSchema,
		APIKey:       apiKey,
	})
}
