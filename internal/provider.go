package internal

import "context"

// Settings contains all possible provider-specific settings
type Settings struct {
	MaxTokens   int     // Maximum tokens in response (0 = omit from request)
	Temperature float64 // Sampling temperature (0 = omit from request)
}

// Provider defines the internal interface for LLM providers
type Provider interface {
	Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, settings Settings, files ...File) (string, error)
	Agent(apiKey string) (interface{}, error)
}

// File represents a file attachment for prompts
type File struct {
	ID       string // Provider-specific file ID
	URI      string // For Google (file URI)
	MimeType string // MIME type of the file
}
