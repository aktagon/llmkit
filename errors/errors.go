package errors

import "fmt"

// APIError represents an error from an LLM API
type APIError struct {
	Provider   string
	StatusCode int
	Message    string
	Endpoint   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s API error (status %d): %s", e.Provider, e.StatusCode, e.Message)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// RequestError represents an error building or sending a request
type RequestError struct {
	Operation string
	Err       error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("request error during %s: %v", e.Operation, e.Err)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}

// SchemaError represents a JSON schema validation error
type SchemaError struct {
	Field   string
	Message string
}

func (e *SchemaError) Error() string {
	return fmt.Sprintf("schema error for field '%s': %s", e.Field, e.Message)
}
