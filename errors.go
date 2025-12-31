package llmkit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// APIError represents a provider API error.
type APIError struct {
	Provider   string
	StatusCode int
	Type       string
	Message    string
	Retryable  bool
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s (%d)", e.Provider, e.Message, e.StatusCode)
}

// ValidationError represents a request validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s - %s", e.Field, e.Message)
}

// parseError parses provider-specific error responses into APIError.
func parseError(provider string, statusCode int, body []byte, headers http.Header) *APIError {
	apiErr := &APIError{
		Provider:   provider,
		StatusCode: statusCode,
		Retryable:  statusCode == 429 || statusCode >= 500,
		RetryAfter: extractRetryAfter(headers),
	}

	switch provider {
	case Anthropic:
		var resp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &resp) == nil {
			apiErr.Type = resp.Error.Type
			apiErr.Message = resp.Error.Message
		}

	case OpenAI, Grok:
		var resp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &resp) == nil {
			apiErr.Type = resp.Error.Type
			apiErr.Message = resp.Error.Message
		}

	case Google:
		var resp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &resp) == nil {
			apiErr.Type = resp.Error.Status
			apiErr.Message = resp.Error.Message
		}
	}

	if apiErr.Message == "" {
		apiErr.Message = string(body)
	}

	return apiErr
}

// extractRetryAfter parses the Retry-After header.
func extractRetryAfter(headers http.Header) time.Duration {
	if v := headers.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return 0
}
