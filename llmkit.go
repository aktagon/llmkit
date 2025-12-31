package llmkit

import (
	"context"
	"os"
	"path/filepath"
)

// optionSupport defines which options each provider supports.
// This mirrors the Option Support Matrix in README.md.
type optionSupport struct {
	temperature      bool
	topP             bool
	topK             bool
	maxTokens        bool
	stopSequences    bool
	seed             bool
	frequencyPenalty bool
	presencePenalty  bool
	thinkingBudget   bool
	reasoningEffort  bool
}

// support maps providers to their supported options.
var support = map[string]optionSupport{
	Anthropic: {
		temperature: true, topP: true, topK: true, maxTokens: true,
		stopSequences: true, thinkingBudget: true,
	},
	OpenAI: {
		temperature: true, topP: true, maxTokens: true, stopSequences: true,
		seed: true, frequencyPenalty: true, presencePenalty: true, reasoningEffort: true,
	},
	Google: {
		temperature: true, topP: true, topK: true, maxTokens: true,
		stopSequences: true, seed: true, thinkingBudget: true, reasoningEffort: true,
	},
	Grok: {
		temperature: true, topP: true, topK: true, maxTokens: true,
		stopSequences: true, seed: true, frequencyPenalty: true, presencePenalty: true,
	},
}

// Prompt sends a one-shot request to an LLM provider.
func Prompt(ctx context.Context, p Provider, req Request, opts ...Option) (Response, error) {
	o := applyOptions(opts...)

	// Before hook
	if o.beforeRequest != nil {
		if err := o.beforeRequest(ctx, &req); err != nil {
			return Response{}, err
		}
	}

	// Validate
	if err := validateProvider(p); err != nil {
		return Response{}, err
	}
	if err := validateRequest(req); err != nil {
		return Response{}, err
	}
	if err := validateOptions(p, o); err != nil {
		return Response{}, err
	}

	// Route to provider
	var resp Response
	var err error
	switch p.Name {
	case Anthropic:
		resp, err = promptAnthropic(ctx, p, req, o)
	case OpenAI:
		resp, err = promptOpenAI(ctx, p, req, o)
	case Google:
		resp, err = promptGoogle(ctx, p, req, o)
	case Grok:
		resp, err = promptGrok(ctx, p, req, o)
	default:
		return Response{}, &ValidationError{Field: "provider", Message: "unknown: " + p.Name}
	}

	// After hook
	if o.afterResponse != nil {
		o.afterResponse(ctx, &resp, err)
	}

	return resp, err
}

// validateProvider checks that provider is properly configured.
func validateProvider(p Provider) error {
	if p.APIKey == "" {
		return &ValidationError{Field: "api_key", Message: "required"}
	}
	return nil
}

// validateRequest checks that required fields are present.
func validateRequest(req Request) error {
	if req.User == "" && len(req.Messages) == 0 {
		return &ValidationError{Field: "user", Message: "required"}
	}
	return nil
}

// validateOptions checks that options are supported by the provider.
func validateOptions(p Provider, o *options) error {
	s := support[p.Name]

	if o.topK != nil && !s.topK {
		return &ValidationError{Field: "top_k", Message: "not supported by " + p.Name}
	}
	if o.seed != nil && !s.seed {
		return &ValidationError{Field: "seed", Message: "not supported by " + p.Name}
	}
	if o.frequencyPenalty != nil && !s.frequencyPenalty {
		return &ValidationError{Field: "frequency_penalty", Message: "not supported by " + p.Name}
	}
	if o.presencePenalty != nil && !s.presencePenalty {
		return &ValidationError{Field: "presence_penalty", Message: "not supported by " + p.Name}
	}
	if o.thinkingBudget != nil && !s.thinkingBudget {
		return &ValidationError{Field: "thinking_budget", Message: "not supported by " + p.Name}
	}
	if o.reasoningEffort != "" && !s.reasoningEffort {
		return &ValidationError{Field: "reasoning_effort", Message: "not supported by " + p.Name}
	}

	// Google only accepts "low" and "high" for reasoning_effort
	if o.reasoningEffort != "" && p.Name == Google {
		if o.reasoningEffort != "low" && o.reasoningEffort != "high" {
			return &ValidationError{Field: "reasoning_effort", Message: "Google only supports 'low' and 'high'"}
		}
	}

	return nil
}

// UploadFile uploads a file to a provider and returns a File reference.
func UploadFile(ctx context.Context, p Provider, path string, opts ...Option) (File, error) {
	if err := validateProvider(p); err != nil {
		return File{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}

	o := applyOptions(opts...)
	mimeType := detectMimeType(path)
	name := filepath.Base(path)

	switch p.Name {
	case Anthropic:
		return uploadAnthropic(ctx, p, data, name, mimeType, o)
	case OpenAI:
		return uploadOpenAI(ctx, p, data, name, o)
	case Google:
		return uploadGoogle(ctx, p, data, name, mimeType, o)
	case Grok:
		return uploadGrok(ctx, p, data, name, o)
	default:
		return File{}, &ValidationError{Field: "provider", Message: "unknown: " + p.Name}
	}
}
