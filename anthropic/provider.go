package anthropic

import (
	"context"
	"errors"

	"github.com/aktagon/llmkit/anthropic/agents"
	"github.com/aktagon/llmkit/anthropic/types"
	"github.com/aktagon/llmkit/internal"
)

// Provider implements the internal.Provider interface for Anthropic
type Provider struct{}

// NewProvider creates a new Anthropic provider
func NewProvider() *Provider {
	return &Provider{}
}

// Prompt implements the internal.Provider interface
func (p *Provider) Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, settings internal.Settings, files ...internal.File) (string, error) {
	// Validate that max_tokens is set (required for Anthropic)
	if settings.MaxTokens < 1 {
		return "", errors.New("max_tokens is required for Anthropic provider and must be >= 1")
	}

	// Validate temperature range
	if settings.Temperature < 0 || settings.Temperature > 1 {
		return "", errors.New("temperature must be between 0 and 1")
	}

	// Validate top_p range
	if settings.TopP < 0 || settings.TopP > 1 {
		return "", errors.New("top_p must be between 0 and 1")
	}

	// Validate top_k range
	if settings.TopK < 0 {
		return "", errors.New("top_k must be >= 0")
	}

	// Convert internal.File to types.File
	anthropicFiles := make([]types.File, len(files))
	for i, f := range files {
		anthropicFiles[i] = types.File{
			ID: f.ID,
		}
	}

	requestSettings := types.RequestSettings{
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
		TopK:        settings.TopK,
		TopP:        settings.TopP,
	}
	return PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey, requestSettings, anthropicFiles...)
}

// Agent implements the internal.Provider interface
func (p *Provider) Agent(apiKey string) (interface{}, error) {
	return agents.New(apiKey)
}
