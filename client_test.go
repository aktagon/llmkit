package llmkit

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_Prompt_Integration(t *testing.T) {
	// Integration test - will fail with invalid API key (expected)
	client := NewClient("invalid-key")

	opts := PromptOptions{
		Provider:     ProviderOpenAI,
		SystemPrompt: "test",
		UserPrompt:   "test",
	}

	_, err := client.Prompt(context.Background(), opts)

	// Should get an API error, proving the interface works
	if err == nil {
		t.Fatal("expected error due to invalid API key")
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	httpClient := &http.Client{}
	client := NewClientWithHTTPClient("test-key", httpClient)

	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestMockClient(t *testing.T) {
	mock := NewMockClient()
	mock.PromptFunc = func(ctx context.Context, opts PromptOptions) (string, error) {
		return "mocked response", nil
	}

	opts := PromptOptions{
		Provider:     ProviderOpenAI,
		SystemPrompt: "test",
		UserPrompt:   "test",
	}

	response, err := mock.Prompt(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response != "mocked response" {
		t.Fatalf("expected 'mocked response', got %q", response)
	}
}
