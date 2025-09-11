package anthropic

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aktagon/llmkit/anthropic/types"
)

func TestParseAnthropicResponse(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"simple math response", "testdata/simple_response.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Skipf("Test fixture %s not available. Run scripts/capture_anthropic_response.go first", tt.file)
			}

			var response types.AnthropicResponse
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Test response structure
			if len(response.Content) == 0 {
				t.Fatal("Response should have at least one content item")
			}

			text := response.Content[0].Text
			if text == "" {
				t.Fatal("First content item should have non-empty text")
			}

			// Test metadata
			if response.ID == "" {
				t.Fatal("Response should have an ID")
			}

			if response.Model == "" {
				t.Fatal("Response should have a model")
			}

			if response.Role != "assistant" {
				t.Fatalf("Response role should be 'assistant', got %q", response.Role)
			}

			// Test usage metadata
			if response.Usage.OutputTokens == 0 {
				t.Fatal("Response should have usage information with output tokens > 0")
			}

			t.Logf("Response text: %q", text)
			t.Logf("Model: %s, ID: %s", response.Model, response.ID)
			t.Logf("Token usage: %d input + %d output tokens",
				response.Usage.InputTokens,
				response.Usage.OutputTokens)
		})
	}
}