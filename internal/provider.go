package internal

import "context"

// Provider defines the internal interface for LLM providers
type Provider interface {
	Prompt(ctx context.Context, systemPrompt, userPrompt, jsonSchema, apiKey string, files ...File) (string, error)
	Agent(apiKey string) (interface{}, error)
}

// File represents a file attachment for prompts
type File struct {
	ID       string // Provider-specific file ID
	URI      string // For Google (file URI)  
	MimeType string // MIME type of the file
}