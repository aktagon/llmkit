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

	// Should use Responses API endpoint
	if capturedPath != "/v1/responses" {
		t.Errorf("path = %q, want /v1/responses", capturedPath)
	}

	// Verify request format uses role/content (new API format)
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	input, ok := body["input"].([]any)
	if !ok {
		t.Fatalf("expected input array, got %T", body["input"])
	}

	// Should have system message and user message with file
	if len(input) != 2 {
		t.Fatalf("expected 2 input messages, got %d: %s", len(input), capturedBody)
	}

	// Check system message
	sysMsg := input[0].(map[string]any)
	if sysMsg["role"] != "system" {
		t.Errorf("first message role = %q, want 'system'", sysMsg["role"])
	}
	if sysMsg["content"] != "Summarize documents" {
		t.Errorf("system content = %q, want 'Summarize documents'", sysMsg["content"])
	}

	// Check user message with file content
	userMsg := input[1].(map[string]any)
	if userMsg["role"] != "user" {
		t.Errorf("second message role = %q, want 'user'", userMsg["role"])
	}
	// Content should be array with file and text parts
	content, ok := userMsg["content"].([]any)
	if !ok {
		t.Fatalf("user content should be array, got %T", userMsg["content"])
	}
	if len(content) != 2 {
		t.Fatalf("expected 2 content parts, got %d", len(content))
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
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
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
		System: "Be friendly",
		User:   "Say hello",
	}

	_, err := Prompt(context.Background(), p, req)
	if err != nil {
		t.Fatalf("Prompt() error = %v", err)
	}

	// Always use Responses API (xAI's preferred endpoint)
	if capturedPath != "/v1/responses" {
		t.Errorf("path = %q, want /v1/responses", capturedPath)
	}

	// Verify request format uses role/content
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	input, ok := body["input"].([]any)
	if !ok {
		t.Fatalf("expected input array, got %T", body["input"])
	}

	// Should have system and user messages
	if len(input) != 2 {
		t.Fatalf("expected 2 input messages, got %d: %s", len(input), capturedBody)
	}

	// Check system message
	sysMsg := input[0].(map[string]any)
	if sysMsg["role"] != "system" {
		t.Errorf("first message role = %q, want 'system'", sysMsg["role"])
	}

	// Check user message - content should be string for text-only
	userMsg := input[1].(map[string]any)
	if userMsg["role"] != "user" {
		t.Errorf("second message role = %q, want 'user'", userMsg["role"])
	}
	if userMsg["content"] != "Say hello" {
		t.Errorf("user content = %q, want 'Say hello'", userMsg["content"])
	}
}
