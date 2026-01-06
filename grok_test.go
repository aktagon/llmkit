package llmkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPromptGrok_ResponsesAPI_WithFiles(t *testing.T) {
	var capturedPath string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		// Responses API format
		w.Write([]byte(`{
			"id": "resp_123",
			"output": [{"type": "message", "content": [{"type": "output_text", "text": "PDF summary"}]}],
			"usage": {"input_tokens": 100, "output_tokens": 50}
		}`))
	}))
	defer server.Close()

	p := Provider{
		Name:    Grok,
		APIKey:  "test-key",
		BaseURL: server.URL,
	}

	req := Request{
		System: "Summarize documents",
		User:   "Summarize this PDF",
		Files:  []File{{ID: "file-123", MimeType: "application/pdf"}},
	}

	resp, err := Prompt(context.Background(), p, req)
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}

	// Should use Responses API endpoint when files attached
	if capturedPath != "/v1/responses" {
		t.Errorf("path = %q, want /v1/responses", capturedPath)
	}

	// Verify request format uses input_file (not nested file object)
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	input, ok := body["input"].([]any)
	if !ok {
		t.Fatalf("expected input array, got %T", body["input"])
	}

	// Find the file input
	var foundFile bool
	for _, item := range input {
		m := item.(map[string]any)
		if m["type"] == "input_file" && m["file_id"] == "file-123" {
			foundFile = true
		}
	}
	if !foundFile {
		t.Errorf("expected input_file with file_id, got: %s", capturedBody)
	}

	if resp.Text != "PDF summary" {
		t.Errorf("text = %q, want 'PDF summary'", resp.Text)
	}
	if resp.Tokens.Input != 100 || resp.Tokens.Output != 50 {
		t.Errorf("tokens = %+v, want input=100, output=50", resp.Tokens)
	}
}

func TestPromptGrok_ResponsesAPI_TextOnly(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		// Responses API format
		w.Write([]byte(`{
			"id": "resp_456",
			"output": [{"type": "message", "content": [{"type": "output_text", "text": "Hello!"}]}],
			"usage": {"input_tokens": 10, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	p := Provider{
		Name:    Grok,
		APIKey:  "test-key",
		BaseURL: server.URL,
	}

	req := Request{
		User: "Say hello",
	}

	_, err := Prompt(context.Background(), p, req)
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}

	// Always use Responses API (xAI's preferred endpoint)
	if capturedPath != "/v1/responses" {
		t.Errorf("path = %q, want /v1/responses", capturedPath)
	}
}
