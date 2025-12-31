package llmkit

import (
	"context"
	"net/http"
	"os"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

func newRecorder(t *testing.T, name string) (*recorder.Recorder, func()) {
	t.Helper()

	mode := recorder.ModeReplayOnly
	if os.Getenv("LLMKIT_RECORD") == "1" {
		mode = recorder.ModeRecordOnly
	}

	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       "testdata/cassettes/" + name,
		Mode:               mode,
		SkipRequestLatency: true,
	})
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}

	// Custom matcher that ignores API key in URL
	r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		// Compare method
		if r.Method != i.Method {
			return false
		}
		// Compare URL with redacted key
		reqURL := redactURLKey(r.URL.String())
		return reqURL == i.URL
	})

	// Whitelist headers - only keep what's needed for replay
	r.AddHook(func(i *cassette.Interaction) error {
		reqAllow := map[string]bool{
			"Content-Type":      true,
			"Anthropic-Version": true,
		}
		respAllow := map[string]bool{
			"Content-Type": true,
		}

		for h := range i.Request.Headers {
			if !reqAllow[h] {
				delete(i.Request.Headers, h)
			}
		}
		for h := range i.Response.Headers {
			if !respAllow[h] {
				delete(i.Response.Headers, h)
			}
		}

		// Redact Google API key from URL
		i.Request.URL = redactURLKey(i.Request.URL)
		return nil
	}, recorder.BeforeSaveHook)

	return r, func() { r.Stop() }
}

func redactURLKey(u string) string {
	const key = "key="
	idx := -1
	for i := 0; i <= len(u)-len(key); i++ {
		if u[i:i+len(key)] == key {
			idx = i
			break
		}
	}
	if idx == -1 {
		return u
	}
	end := idx + len(key)
	for end < len(u) && u[end] != '&' {
		end++
	}
	return u[:idx+len(key)] + "REDACTED" + u[end:]
}

func TestPromptAnthropic_Chat(t *testing.T) {
	rec, stop := newRecorder(t, "anthropic-chat")
	defer stop()

	p := Provider{
		Name:   Anthropic,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key" // for replay mode
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

func TestPromptAnthropic_WithSystem(t *testing.T) {
	rec, stop := newRecorder(t, "anthropic-chat-system")
	defer stop()

	p := Provider{
		Name:   Anthropic,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
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

func TestBuildAnthropicContent(t *testing.T) {
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
			name: "image URL and text",
			req: Request{
				User:   "describe this",
				Images: []Image{{URL: "https://example.com/img.jpg", MimeType: "image/jpeg"}},
			},
			wantLen:   2,
			wantTypes: []string{"image", "text"},
		},
		{
			name: "base64 image and text",
			req: Request{
				User:   "describe this",
				Images: []Image{{URL: "data:image/png;base64,iVBORw0KGgo=", MimeType: "image/png"}},
			},
			wantLen:   2,
			wantTypes: []string{"image", "text"},
		},
		{
			name: "file and text",
			req: Request{
				User:  "summarize this",
				Files: []File{{ID: "file_123", MimeType: "application/pdf"}},
			},
			wantLen:   2,
			wantTypes: []string{"document", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := buildAnthropicContent(tt.req)
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

func TestExtractBase64Data(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"data:image/png;base64,abc123", "abc123"},
		{"data:image/jpeg;base64,xyz", "xyz"},
		{"abc123", "abc123"}, // no prefix
	}

	for _, tt := range tests {
		got := extractBase64Data(tt.input)
		if got != tt.want {
			t.Errorf("extractBase64Data(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
