package google

import (
	"context"

	"github.com/aktagon/llmkit/google/agents"
	"github.com/aktagon/llmkit/internal"
)

// Provider implements the internal.Provider interface for Google
type Provider struct{}

// NewProvider creates a new Google provider
func NewProvider() *Provider {
	return &Provider{}
}

// Prompt implements the internal.Provider interface
func (p *Provider) Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, settings internal.Settings, files ...internal.File) (string, error) {
	// Convert internal.File to google.File
	googleFiles := make([]File, len(files))
	for i, f := range files {
		googleFiles[i] = File{
			URI:      f.URI,
			MimeType: f.MimeType,
		}
	}

	requestSettings := RequestSettings{
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
	}
	response, err := PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey, requestSettings, googleFiles...)
	if err != nil {
		return "", err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// Agent implements the internal.Provider interface
func (p *Provider) Agent(apiKey string) (interface{}, error) {
	return agents.New(apiKey)
}
