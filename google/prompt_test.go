package google

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseGoogleResponse(t *testing.T) {
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
				t.Skipf("Test fixture %s not available. Run scripts/capture_google_response.go first", tt.file)
			}

			var response GoogleResponse
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Test response structure
			if len(response.Candidates) == 0 {
				t.Fatal("Response should have at least one candidate")
			}

			if len(response.Candidates[0].Content.Parts) == 0 {
				t.Fatal("First candidate should have at least one content part")
			}

			text := response.Candidates[0].Content.Parts[0].Text
			if text == "" {
				t.Fatal("First content part should have non-empty text")
			}

			// Test usage metadata
			if response.UsageMetadata.TotalTokenCount == 0 {
				t.Fatal("Response should have usage metadata with total tokens > 0")
			}

			t.Logf("Response text: %q", text)
			t.Logf("Token usage: %d total (%d prompt + %d candidates)", 
				response.UsageMetadata.TotalTokenCount,
				response.UsageMetadata.PromptTokenCount, 
				response.UsageMetadata.CandidatesTokenCount)
		})
	}
}