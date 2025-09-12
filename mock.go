package llmkit

import "context"

// MockClient implements Client interface for testing
type MockClient struct {
	PromptFunc func(ctx context.Context, opts PromptOptions) (string, error)
}

// Prompt delegates to the mock function
func (m *MockClient) Prompt(ctx context.Context, opts PromptOptions) (string, error) {
	if m.PromptFunc != nil {
		return m.PromptFunc(ctx, opts)
	}
	return "mock response", nil
}

// NewMockClient creates a new mock client with default behavior
func NewMockClient() *MockClient {
	return &MockClient{}
}
