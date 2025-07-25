package openai

import (
	"testing"

	"github.com/aktagon/llmkit/errors"
)

func TestValidateTTSInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		options     *TTSOptions
		expectError bool
		errorType   string
	}{
		{
			name:        "valid input",
			input:       "Hello world",
			options:     nil,
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "whitespace only input",
			input:       "   ",
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "input too long",
			input:       string(make([]byte, 4097)),
			options:     nil,
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "invalid speed too low",
			input:       "Hello world",
			options:     &TTSOptions{Speed: 0.1},
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "invalid speed too high",
			input:       "Hello world",
			options:     &TTSOptions{Speed: 5.0},
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name:        "valid speed range",
			input:       "Hello world",
			options:     &TTSOptions{Speed: 1.5},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTTSInput(tt.input, tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				switch tt.errorType {
				case "ValidationError":
					if _, ok := err.(*errors.ValidationError); !ok {
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

func TestBuildTTSRequest(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *TTSOptions
		expected TTSRequest
	}{
		{
			name:    "default options",
			input:   "Hello world",
			options: nil,
			expected: TTSRequest{
				Model: ModelTTS1,
				Input: "Hello world",
				Voice: VoiceAlloy,
			},
		},
		{
			name:  "custom options",
			input: "Custom test",
			options: &TTSOptions{
				Model:          ModelTTS1HD,
				Voice:          VoiceNova,
				ResponseFormat: FormatWAV,
				Speed:          1.5,
			},
			expected: TTSRequest{
				Model:          ModelTTS1HD,
				Input:          "Custom test",
				Voice:          VoiceNova,
				ResponseFormat: FormatWAV,
				Speed:          1.5,
			},
		},
		{
			name:  "partial options",
			input: "Partial test",
			options: &TTSOptions{
				Voice: VoiceEcho,
				Speed: 2.0,
			},
			expected: TTSRequest{
				Model: ModelTTS1,
				Input: "Partial test",
				Voice: VoiceEcho,
				Speed: 2.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := buildTTSRequest(tt.input, tt.options)
			if err != nil {
				t.Fatalf("buildTTSRequest failed: %v", err)
			}

			if len(data) == 0 {
				t.Error("expected non-empty request data")
			}

			// Basic validation that JSON was created
			if data[0] != '{' {
				t.Error("expected JSON object to start with '{'")
			}
		})
	}
}

func TestText2Speech_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		apiKey    string
		options   *TTSOptions
		errorType string
	}{
		{
			name:      "missing API key",
			input:     "Hello world",
			apiKey:    "",
			options:   nil,
			errorType: "ValidationError",
		},
		{
			name:      "empty input",
			input:     "",
			apiKey:    "test-key",
			options:   nil,
			errorType: "ValidationError",
		},
		{
			name:      "invalid speed",
			input:     "Hello world",
			apiKey:    "test-key",
			options:   &TTSOptions{Speed: 10.0},
			errorType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Text2Speech(tt.input, tt.apiKey, tt.options)

			if err == nil {
				t.Error("expected error but got none")
				return
			}

			switch tt.errorType {
			case "ValidationError":
				if _, ok := err.(*errors.ValidationError); !ok {
					t.Errorf("expected ValidationError but got %T: %v", err, err)
				}
			}
		})
	}
}
