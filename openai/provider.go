package openai

import (
	"context"

	"github.com/aktagon/llmkit/internal"
	"github.com/aktagon/llmkit/openai/agents"
	"github.com/aktagon/llmkit/openai/types"
)

// Provider implements the internal.Provider interface for OpenAI
type Provider struct{}

// NewProvider creates a new OpenAI provider
func NewProvider() *Provider {
	return &Provider{}
}

// Prompt implements the internal.Provider interface
func (p *Provider) Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, files ...internal.File) (string, error) {
	// Convert internal.File to types.FileUploadResponse
	openaiFiles := make([]types.FileUploadResponse, len(files))
	for i, f := range files {
		openaiFiles[i] = types.FileUploadResponse{
			ID: f.ID,
		}
	}

	return Prompt(systemPrompt, userPrompt, jsonSchema, apiKey, openaiFiles...)
}

// Agent implements the internal.Provider interface
func (p *Provider) Agent(apiKey string) (interface{}, error) {
	return agents.New(apiKey)
}