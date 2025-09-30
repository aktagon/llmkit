package grok

import (
	"context"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/grok/types"
	"github.com/aktagon/llmkit/internal"
	openaitypes "github.com/aktagon/llmkit/openai/types"
)

// Provider implements the internal.Provider interface for Grok
type Provider struct{}

// NewProvider creates a new Grok provider
func NewProvider() *Provider {
	return &Provider{}
}

// Prompt implements the internal.Provider interface
func (p *Provider) Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, settings internal.Settings, files ...internal.File) (string, error) {
	// Convert internal.File to openaitypes.FileUploadResponse
	grokFiles := make([]openaitypes.FileUploadResponse, len(files))
	for i, f := range files {
		grokFiles[i] = openaitypes.FileUploadResponse{
			ID: f.ID,
		}
	}

	requestSettings := types.RequestSettings{
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
	}
	response, err := PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey, requestSettings, grokFiles...)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", nil
	}

	return response.Choices[0].Message.Content, nil
}

// Agent implements the internal.Provider interface
func (p *Provider) Agent(apiKey string) (interface{}, error) {
	return nil, &errors.ValidationError{
		Field:   "agent",
		Message: "Grok does not support agents",
	}
}