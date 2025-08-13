package google

import (
	"context"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/internal"
)

// Provider implements the internal.Provider interface for Google
type Provider struct{}

// NewProvider creates a new Google provider
func NewProvider() *Provider {
	return &Provider{}
}

// Prompt implements the internal.Provider interface
func (p *Provider) Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, files ...internal.File) (string, error) {
	// Convert internal.File to google.File
	googleFiles := make([]File, len(files))
	for i, f := range files {
		googleFiles[i] = File{
			URI:      f.URI,
			MimeType: f.MimeType,
		}
	}

	return Prompt(systemPrompt, userPrompt, jsonSchema, apiKey, googleFiles...)
}

// Agent implements the internal.Provider interface
func (p *Provider) Agent(apiKey string) (interface{}, error) {
	return nil, &errors.ValidationError{
		Field:   "provider",
		Message: "Google agent is not yet implemented",
	}
}