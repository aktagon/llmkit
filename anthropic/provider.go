package anthropic

import (
	"context"

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
func (p *Provider) Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, files ...internal.File) (string, error) {
	// Convert internal.File to types.File
	anthropicFiles := make([]types.File, len(files))
	for i, f := range files {
		anthropicFiles[i] = types.File{
			ID: f.ID,
		}
	}

	return Prompt(systemPrompt, userPrompt, jsonSchema, apiKey, anthropicFiles...)
}

// Agent implements the internal.Provider interface
func (p *Provider) Agent(apiKey string) (interface{}, error) {
	return agents.New(apiKey)
}