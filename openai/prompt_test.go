package openai

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aktagon/llmkit/openai/types"
)

func TestParseOpenAIResponse(t *testing.T) {
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
				t.Skipf("Test fixture %s not available. Run: go run cmd/tools/capture_response.go openai", tt.file)
			}

			var response types.Response
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Test response structure
			if len(response.Choices) == 0 {
				t.Fatal("Response should have at least one choice")
			}

			content := response.Choices[0].Message.Content
			if content == "" {
				t.Fatal("First choice should have non-empty content")
			}

			// Test metadata
			if response.ID == "" {
				t.Fatal("Response should have an ID")
			}

			if response.Model == "" {
				t.Fatal("Response should have a model")
			}

			// Test usage metadata
			if response.Usage.TotalTokens == 0 {
				t.Fatal("Response should have usage information with total tokens > 0")
			}

			t.Logf("Response text: %q", content)
			t.Logf("Model: %s, ID: %s", response.Model, response.ID)
			t.Logf("Token usage: %d total (%d prompt + %d completion)",
				response.Usage.TotalTokens,
				response.Usage.PromptTokens,
				response.Usage.CompletionTokens)
		})
	}
}