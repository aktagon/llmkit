package llmkit

import (
	"net/http"
	"testing"
	"time"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Provider:   "anthropic",
		StatusCode: 429,
		Message:    "rate limit exceeded",
	}
	want := "anthropic: rate limit exceeded (429)"
	if got := err.Error(); got != want {
		t.Errorf("APIError.Error() = %q, want %q", got, want)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "user",
		Message: "required",
	}
	want := "validation: user - required"
	if got := err.Error(); got != want {
		t.Errorf("ValidationError.Error() = %q, want %q", got, want)
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name       string
		provider   string
		statusCode int
		body       string
		wantType   string
		wantMsg    string
		retryable  bool
	}{
		{
			name:       "anthropic rate limit",
			provider:   Anthropic,
			statusCode: 429,
			body:       `{"type":"error","error":{"type":"rate_limit_error","message":"Rate limit exceeded"}}`,
			wantType:   "rate_limit_error",
			wantMsg:    "Rate limit exceeded",
			retryable:  true,
		},
		{
			name:       "anthropic invalid request",
			provider:   Anthropic,
			statusCode: 400,
			body:       `{"type":"error","error":{"type":"invalid_request_error","message":"Invalid model"}}`,
			wantType:   "invalid_request_error",
			wantMsg:    "Invalid model",
			retryable:  false,
		},
		{
			name:       "openai rate limit",
			provider:   OpenAI,
			statusCode: 429,
			body:       `{"error":{"message":"Rate limit exceeded","type":"rate_limit_error","code":"rate_limit_exceeded"}}`,
			wantType:   "rate_limit_error",
			wantMsg:    "Rate limit exceeded",
			retryable:  true,
		},
		{
			name:       "openai auth error",
			provider:   OpenAI,
			statusCode: 401,
			body:       `{"error":{"message":"Invalid API key","type":"invalid_api_key"}}`,
			wantType:   "invalid_api_key",
			wantMsg:    "Invalid API key",
			retryable:  false,
		},
		{
			name:       "google rate limit",
			provider:   Google,
			statusCode: 429,
			body:       `{"error":{"code":429,"message":"Resource exhausted","status":"RESOURCE_EXHAUSTED"}}`,
			wantType:   "RESOURCE_EXHAUSTED",
			wantMsg:    "Resource exhausted",
			retryable:  true,
		},
		{
			name:       "grok server error",
			provider:   Grok,
			statusCode: 500,
			body:       `{"error":{"message":"Internal server error","type":"server_error"}}`,
			wantType:   "server_error",
			wantMsg:    "Internal server error",
			retryable:  true,
		},
		{
			name:       "502 gateway error",
			provider:   OpenAI,
			statusCode: 502,
			body:       `Bad Gateway`,
			wantType:   "",
			wantMsg:    "Bad Gateway",
			retryable:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			got := parseError(tt.provider, tt.statusCode, []byte(tt.body), headers)

			if got.Provider != tt.provider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.provider)
			}
			if got.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", got.StatusCode, tt.statusCode)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", got.Message, tt.wantMsg)
			}
			if got.Retryable != tt.retryable {
				t.Errorf("Retryable = %v, want %v", got.Retryable, tt.retryable)
			}
		})
	}
}

func TestExtractRetryAfter(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		want    time.Duration
	}{
		{
			name:    "no header",
			headers: http.Header{},
			want:    0,
		},
		{
			name: "retry-after seconds",
			headers: http.Header{
				"Retry-After": []string{"30"},
			},
			want: 30 * time.Second,
		},
		{
			name: "invalid retry-after",
			headers: http.Header{
				"Retry-After": []string{"invalid"},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRetryAfter(tt.headers)
			if got != tt.want {
				t.Errorf("extractRetryAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}
