package openai

import (
	"strings"
	"testing"

	"github.com/aktagon/llmkit"
)

func TestValidateSTTInput(t *testing.T) {
	tests := []struct {
		name        string
		audioData   []byte
		filename    string
		options     *STTOptions
		expectError bool
		errorType   string
	}{
		{
			name:        "valid input",
			audioData:   []byte("fake audio data"),
			filename:    "test.mp3",
			options:     nil,
			expectError: false,
		},
		{
			name:        "empty audio data",
			audioData:   []byte{},
			filename:    "test.mp3",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "nil audio data",
			audioData:   nil,
			filename:    "test.mp3",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "empty filename",
			audioData:   []byte("fake audio data"),
			filename:    "",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "whitespace only filename",
			audioData:   []byte("fake audio data"),
			filename:    "   ",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "unsupported file extension",
			audioData:   []byte("fake audio data"),
			filename:    "test.txt",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "invalid temperature too low",
			audioData:   []byte("fake audio data"),
			filename:    "test.wav",
			options:     &STTOptions{Temperature: -0.1},
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "invalid temperature too high",
			audioData:   []byte("fake audio data"),
			filename:    "test.wav",
			options:     &STTOptions{Temperature: 1.1},
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "valid temperature range",
			audioData:   []byte("fake audio data"),
			filename:    "test.flac",
			options:     &STTOptions{Temperature: 0.5},
			expectError: false,
		},
		{
			name:        "supported extensions case insensitive",
			audioData:   []byte("fake audio data"),
			filename:    "test.MP3",
			options:     nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSTTInput(tt.audioData, tt.filename, tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				switch tt.errorType {
				case "ValidationError":
					if _, ok := err.(*llmkit.ValidationError); !ok {
						t.Errorf("expected ValidationError but got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildSTTRequest(t *testing.T) {
	tests := []struct {
		name      string
		audioData []byte
		filename  string
		options   *STTOptions
	}{
		{
			name:      "default options",
			audioData: []byte("fake audio data"),
			filename:  "test.mp3",
			options:   nil,
		},
		{
			name:      "custom options",
			audioData: []byte("fake audio data"),
			filename:  "test.wav",
			options: &STTOptions{
				Model:          ModelGPT4OTranscribe,
				Language:       "en",
				Prompt:         "This is a test prompt",
				ResponseFormat: STTFormatVerboseJSON,
				Temperature:    0.3,
				TimestampGranularities: []TimestampGranularity{
					GranularityWord,
					GranularitySegment,
				},
			},
		},
		{
			name:      "partial options",
			audioData: []byte("fake audio data"),
			filename:  "recording.flac",
			options: &STTOptions{
				Language: "es",
				Prompt:   "Spanish audio",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, contentType, err := buildSTTRequest(tt.audioData, tt.filename, tt.options)
			if err != nil {
				t.Fatalf("buildSTTRequest failed: %v", err)
			}

			if body == nil || body.Len() == 0 {
				t.Error("expected non-empty request body")
			}

			if contentType == "" {
				t.Error("expected non-empty content type")
			}

			// Verify it's multipart form data
			if !containsString(contentType, "multipart/form-data") {
				t.Errorf("expected multipart/form-data content type, got: %s", contentType)
			}
		})
	}
}

func TestSpeech2Text_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		audioData []byte
		filename  string
		apiKey    string
		options   *STTOptions
		errorType string
	}{
		{
			name:      "missing API key",
			audioData: []byte("fake audio data"),
			filename:  "test.mp3",
			apiKey:    "",
			options:   nil,
			errorType: "ValidationError",
		},
		{
			name:      "empty audio data",
			audioData: []byte{},
			filename:  "test.mp3",
			apiKey:    "test-key",
			options:   nil,
			errorType: "ValidationError",
		},
		{
			name:      "invalid filename",
			audioData: []byte("fake audio data"),
			filename:  "test.txt",
			apiKey:    "test-key",
			options:   nil,
			errorType: "ValidationError",
		},
		{
			name:      "invalid temperature",
			audioData: []byte("fake audio data"),
			filename:  "test.mp3",
			apiKey:    "test-key",
			options:   &STTOptions{Temperature: 2.0},
			errorType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Speech2Text(tt.audioData, tt.filename, tt.apiKey, tt.options)

			if err == nil {
				t.Error("expected error but got none")
				return
			}

			switch tt.errorType {
			case "ValidationError":
				if _, ok := err.(*llmkit.ValidationError); !ok {
					t.Errorf("expected ValidationError but got %T: %v", err, err)
				}
			}
		})
	}
}

func TestSpeech2TextDetailed_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		audioData []byte
		filename  string
		apiKey    string
		options   *STTOptions
		errorType string
	}{
		{
			name:      "missing API key",
			audioData: []byte("fake audio data"),
			filename:  "test.mp3",
			apiKey:    "",
			options:   nil,
			errorType: "ValidationError",
		},
		{
			name:      "empty audio data",
			audioData: []byte{},
			filename:  "test.mp3",
			apiKey:    "test-key",
			options:   nil,
			errorType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Speech2TextDetailed(tt.audioData, tt.filename, tt.apiKey, tt.options)

			if err == nil {
				t.Error("expected error but got none")
				return
			}

			switch tt.errorType {
			case "ValidationError":
				if _, ok := err.(*llmkit.ValidationError); !ok {
					t.Errorf("expected ValidationError but got %T: %v", err, err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
