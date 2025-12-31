package llmkit

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func googleAPIKey() string {
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key
	}
	return "test-key"
}

func TestPromptGoogle_Chat(t *testing.T) {
	rec, stop := newRecorder(t, "google-chat")
	defer stop()

	p := Provider{
		Name:   Google,
		APIKey: googleAPIKey(),
	}

	req := Request{
		User: "Say hello in exactly 3 words",
	}

	resp, err := Prompt(context.Background(), p, req,
		WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}

	if resp.Text == "" {
		t.Error("expected non-empty response text")
	}
	if resp.Tokens.Input == 0 {
		t.Error("expected non-zero input tokens")
	}
	if resp.Tokens.Output == 0 {
		t.Error("expected non-zero output tokens")
	}
}

func TestPromptGoogle_WithSystem(t *testing.T) {
	rec, stop := newRecorder(t, "google-chat-system")
	defer stop()

	p := Provider{
		Name:   Google,
		APIKey: googleAPIKey(),
	}

	req := Request{
		System: "You are a pirate. Respond in pirate speak.",
		User:   "Hello",
	}

	resp, err := Prompt(context.Background(), p, req,
		WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}

	if resp.Text == "" {
		t.Error("expected non-empty response text")
	}
}

func TestBuildGoogleParts(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantLen int
	}{
		{
			name:    "text only",
			req:     Request{User: "hello"},
			wantLen: 1,
		},
		{
			name: "image and text",
			req: Request{
				User:   "describe this",
				Images: []Image{{URL: "data:image/png;base64,abc123", MimeType: "image/png"}},
			},
			wantLen: 2,
		},
		{
			name: "file and text",
			req: Request{
				User:  "summarize this",
				Files: []File{{URI: "https://example.com/files/abc", MimeType: "application/pdf"}},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := buildGoogleParts(tt.req)
			if len(parts) != tt.wantLen {
				t.Errorf("got %d parts, want %d", len(parts), tt.wantLen)
			}
		})
	}
}
