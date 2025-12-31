package llmkit

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFile_OpenAI(t *testing.T) {
	rec, stop := newRecorder(t, "openai-upload")
	defer stop()

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	p := Provider{
		Name:   OpenAI,
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	file, err := UploadFile(context.Background(), p, tmpFile,
		WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}

	if file.ID == "" {
		t.Error("expected non-empty file ID")
	}
}

func TestUploadFile_Google(t *testing.T) {
	rec, stop := newRecorder(t, "google-upload")
	defer stop()

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	p := Provider{
		Name:   Google,
		APIKey: googleAPIKey(),
	}

	file, err := UploadFile(context.Background(), p, tmpFile,
		WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}

	if file.URI == "" {
		t.Error("expected non-empty file URI")
	}
}

func TestUploadFile_Grok(t *testing.T) {
	rec, stop := newRecorder(t, "grok-upload")
	defer stop()

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	p := Provider{
		Name:   Grok,
		APIKey: os.Getenv("XAI_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	file, err := UploadFile(context.Background(), p, tmpFile,
		WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}

	if file.ID == "" {
		t.Error("expected non-empty file ID")
	}
}

func TestUploadFile_Anthropic(t *testing.T) {
	rec, stop := newRecorder(t, "anthropic-upload")
	defer stop()

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	p := Provider{
		Name:   Anthropic,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	file, err := UploadFile(context.Background(), p, tmpFile,
		WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}

	if file.ID == "" {
		t.Error("expected non-empty file ID")
	}
}

func TestUploadFile_UnknownProvider(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	p := Provider{
		Name:   "unknown",
		APIKey: "test-key",
	}

	_, err := UploadFile(context.Background(), p, tmpFile)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestUploadFile_MissingFile(t *testing.T) {
	p := Provider{
		Name:   OpenAI,
		APIKey: "test-key",
	}

	_, err := UploadFile(context.Background(), p, "/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestUploadFile_MissingAPIKey(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	p := Provider{
		Name:   OpenAI,
		APIKey: "",
	}

	_, err := UploadFile(context.Background(), p, tmpFile)
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"test.txt", "text/plain"},
		{"test.pdf", "application/pdf"},
		{"test.json", "application/json"},
		{"test.png", "image/png"},
		{"test.jpg", "image/jpeg"},
		{"test.jpeg", "image/jpeg"},
		{"test.gif", "image/gif"},
		{"test.webp", "image/webp"},
		{"test.md", "text/markdown"},
		{"test.csv", "text/csv"},
		{"unknown", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := detectMimeType(tt.filename)
			if got != tt.want {
				t.Errorf("detectMimeType(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
