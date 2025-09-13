package internal

import "context"

// Settings contains all possible provider-specific settings
type Settings struct {
	Model       string  // Model to use for the request (provider-specific default if empty)
	MaxTokens   int     // Maximum tokens in response (0 = omit from request)
	Temperature float64 // Sampling temperature (0 = omit from request)
	TopK        int     // Only sample from the top K options for each subsequent token (0 = omit from request)
	TopP        float64 // Use nucleus sampling (0 = omit from request)
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
