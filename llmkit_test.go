package llmkit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantErr bool
		field   string
	}{
		{
			name:    "valid minimal request",
			req:     Request{User: "Hello"},
			wantErr: false,
		},
		{
			name:    "valid with system",
			req:     Request{System: "You are helpful", User: "Hello"},
			wantErr: false,
		},
		{
			name:    "missing user",
			req:     Request{},
			wantErr: true,
			field:   "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequest(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if valErr.Field != tt.field {
					t.Errorf("Field = %q, want %q", valErr.Field, tt.field)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPrompt_UnknownProvider(t *testing.T) {
	p := Provider{Name: "unknown", APIKey: "key"}
	req := Request{User: "Hello"}

	_, err := Prompt(context.Background(), p, req)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if valErr.Field != "provider" {
		t.Errorf("Field = %q, want provider", valErr.Field)
	}
}

func TestPrompt_ValidationError(t *testing.T) {
	p := Provider{Name: Anthropic, APIKey: "key"}
	req := Request{} // missing User

	_, err := Prompt(context.Background(), p, req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestPrompt_MissingAPIKey(t *testing.T) {
	p := Provider{Name: Anthropic, APIKey: ""} // missing APIKey
	req := Request{User: "Hello"}

	_, err := Prompt(context.Background(), p, req)
	if err == nil {
		t.Fatal("expected validation error for missing API key")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if valErr.Field != "api_key" {
		t.Errorf("Field = %q, want api_key", valErr.Field)
	}
}

func TestPrompt_ThinkingBudget_UnsupportedProvider(t *testing.T) {
	// ThinkingBudget is only supported by Anthropic and Google
	unsupportedProviders := []string{OpenAI, Grok}

	for _, providerName := range unsupportedProviders {
		t.Run(providerName, func(t *testing.T) {
			p := Provider{Name: providerName, APIKey: "test-key"}
			req := Request{User: "Hello"}

			_, err := Prompt(context.Background(), p, req, WithThinkingBudget(4096))
			if err == nil {
				t.Fatal("expected error for unsupported provider")
			}

			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected ValidationError, got %T: %v", err, err)
			}
			if valErr.Field != "thinking_budget" {
				t.Errorf("Field = %q, want thinking_budget", valErr.Field)
			}
		})
	}
}

func TestPrompt_ReasoningEffort_UnsupportedProvider(t *testing.T) {
	// ReasoningEffort is only supported by OpenAI and Google
	unsupportedProviders := []string{Anthropic, Grok}

	for _, providerName := range unsupportedProviders {
		t.Run(providerName, func(t *testing.T) {
			p := Provider{Name: providerName, APIKey: "test-key"}
			req := Request{User: "Hello"}

			_, err := Prompt(context.Background(), p, req, WithReasoningEffort("high"))
			if err == nil {
				t.Fatal("expected error for unsupported provider")
			}

			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected ValidationError, got %T: %v", err, err)
			}
			if valErr.Field != "reasoning_effort" {
				t.Errorf("Field = %q, want reasoning_effort", valErr.Field)
			}
		})
	}
}

func TestPrompt_ReasoningEffort_InvalidValueForGoogle(t *testing.T) {
	// Google only supports "low" and "high", not "medium"
	p := Provider{Name: Google, APIKey: "test-key"}
	req := Request{User: "Hello"}

	_, err := Prompt(context.Background(), p, req, WithReasoningEffort("medium"))
	if err == nil {
		t.Fatal("expected error for invalid reasoning_effort value")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if valErr.Field != "reasoning_effort" {
		t.Errorf("Field = %q, want reasoning_effort", valErr.Field)
	}
}

func TestPrompt_Structured(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		cassette string
		apiKey   string
	}{
		{"anthropic", Anthropic, "anthropic-structured", "ANTHROPIC_API_KEY"},
		{"openai", OpenAI, "openai-structured", "OPENAI_API_KEY"},
		{"google", Google, "google-structured", "GEMINI_API_KEY"},
		{"grok", Grok, "grok-structured", "XAI_API_KEY"},
	}

	schema := `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"],"additionalProperties":false}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, stop := newRecorder(t, tt.cassette)
			defer stop()

			apiKey := os.Getenv(tt.apiKey)
			if apiKey == "" && tt.provider == Google {
				apiKey = googleAPIKey()
			}
			if apiKey == "" {
				apiKey = "test-key"
			}
			p := Provider{
				Name:   tt.provider,
				APIKey: apiKey,
			}

			req := Request{
				User:   "Extract: John is 30 years old.",
				Schema: schema,
			}

			resp, err := Prompt(context.Background(), p, req,
				WithHTTPClient(&http.Client{Transport: rec}),
			)
			if err != nil {
				t.Fatalf("Prompt failed: %v", err)
			}

			var result struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}
			if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
				t.Fatalf("response not valid JSON: %v\ngot: %s", err, resp.Text)
			}
			if result.Name == "" || result.Age == 0 {
				t.Errorf("schema not applied: name=%q age=%d", result.Name, result.Age)
			}
		})
	}
}
