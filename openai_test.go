package llmkit

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestPromptOpenAI_Chat(t *testing.T) {
	rec, stop := newRecorder(t, "openai-chat")
	defer stop()

	p := Provider{
		Name:   OpenAI,
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
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

func TestPromptOpenAI_WithSystem(t *testing.T) {
	rec, stop := newRecorder(t, "openai-chat-system")
	defer stop()

	p := Provider{
		Name:   OpenAI,
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
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

func TestBuildOpenAIContent(t *testing.T) {
	tests := []struct {
		name      string
		req       Request
		wantLen   int
		wantTypes []string
	}{
		{
			name:      "text only",
			req:       Request{User: "hello"},
			wantLen:   1,
			wantTypes: []string{"text"},
		},
		{
			name: "image and text",
			req: Request{
				User:   "describe this",
				Images: []Image{{URL: "https://example.com/img.jpg", MimeType: "image/jpeg"}},
			},
			wantLen:   2,
			wantTypes: []string{"image_url", "text"},
		},
		{
			name: "image with detail",
			req: Request{
				User:   "describe",
				Images: []Image{{URL: "https://example.com/img.jpg", Detail: "high"}},
			},
			wantLen:   2,
			wantTypes: []string{"image_url", "text"},
		},
		{
			name: "file and text",
			req: Request{
				User:  "summarize this",
				Files: []File{{ID: "file-123", MimeType: "application/pdf"}},
			},
			wantLen:   2,
			wantTypes: []string{"file", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := buildOpenAIContent(tt.req)
			if len(content) != tt.wantLen {
				t.Errorf("got %d parts, want %d", len(content), tt.wantLen)
			}
			for i, wantType := range tt.wantTypes {
				if i < len(content) && content[i].Type != wantType {
					t.Errorf("part %d: got type %q, want %q", i, content[i].Type, wantType)
				}
			}
		})
	}
}
