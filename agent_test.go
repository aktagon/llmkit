package llmkit

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestNewAgent(t *testing.T) {
	p := Provider{
		Name:   Anthropic,
		APIKey: "test-key",
	}

	agent := NewAgent(p)
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
}

func TestAgent_Chat(t *testing.T) {
	rec, stop := newRecorder(t, "anthropic-agent-chat")
	defer stop()

	p := Provider{
		Name:   Anthropic,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	agent := NewAgent(p, WithHTTPClient(&http.Client{Transport: rec}))

	resp, err := agent.Chat(context.Background(), "Say hello in exactly 3 words")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if resp.Text == "" {
		t.Error("expected non-empty response text")
	}
}

func TestAgent_MultiTurn(t *testing.T) {
	rec, stop := newRecorder(t, "anthropic-agent-multiturn")
	defer stop()

	p := Provider{
		Name:   Anthropic,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	agent := NewAgent(p, WithHTTPClient(&http.Client{Transport: rec}))
	agent.SetSystem("You are a helpful assistant. Always respond concisely.")

	// First turn - introduce name
	resp1, err := agent.Chat(context.Background(), "My name is Alice.")
	if err != nil {
		t.Fatalf("Chat() turn 1 error = %v", err)
	}
	if resp1.Text == "" {
		t.Error("expected non-empty response for turn 1")
	}

	// Second turn - should remember context and return the name
	resp2, err := agent.Chat(context.Background(), "What is my name? Reply with just the name.")
	if err != nil {
		t.Fatalf("Chat() turn 2 error = %v", err)
	}
	if resp2.Text == "" {
		t.Error("expected non-empty response for turn 2")
	}

	// Verify the model remembers the name from turn 1
	if !containsIgnoreCase(resp2.Text, "Alice") {
		t.Errorf("expected response to contain 'Alice', got: %q", resp2.Text)
	}
}

func TestAgent_Reset(t *testing.T) {
	p := Provider{
		Name:   Anthropic,
		APIKey: "test-key",
	}

	agent := NewAgent(p)
	// Add some state (simulated by adding a tool and then resetting)
	agent.AddTool(Tool{Name: "test"})
	agent.Reset()

	// After reset, tools should be cleared
	// No assertion needed - just verify it doesn't panic
}

func TestAgent_AddTool(t *testing.T) {
	p := Provider{
		Name:   Anthropic,
		APIKey: "test-key",
	}

	agent := NewAgent(p)
	tool := Tool{
		Name:        "get_weather",
		Description: "Get current weather",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{"type": "string"},
			},
			"required": []string{"location"},
		},
		Run: func(args map[string]any) (string, error) {
			return "sunny", nil
		},
	}

	agent.AddTool(tool)
	// No assertion needed - just verify it doesn't panic
}

// testWeatherTool returns a test tool for weather lookups.
func testWeatherTool() Tool {
	return Tool{
		Name:        "get_weather",
		Description: "Get current weather for a city",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{
					"type":        "string",
					"description": "The city name",
				},
			},
			"required": []string{"city"},
		},
		Run: func(input map[string]any) (string, error) {
			city, _ := input["city"].(string)
			return "72°F and sunny in " + city, nil
		},
	}
}

func TestAgent_ChatWithTool(t *testing.T) {
	rec, stop := newRecorder(t, "anthropic-tool-use")
	defer stop()

	p := Provider{
		Name:   Anthropic,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	agent := NewAgent(p, WithHTTPClient(&http.Client{Transport: rec}))
	agent.AddTool(testWeatherTool())

	resp, err := agent.Chat(context.Background(), "What's the weather in Paris?")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	// Response should mention the weather from our tool
	if !containsIgnoreCase(resp.Text, "72") || !containsIgnoreCase(resp.Text, "sunny") {
		t.Errorf("expected response to contain tool result (72, sunny), got: %q", resp.Text)
	}
}

func TestAgent_ChatWithTool_UnknownTool(t *testing.T) {
	// This test verifies error handling when LLM tries to call an unknown tool
	// For now, just verify the agent handles tools gracefully
	p := Provider{
		Name:   Anthropic,
		APIKey: "test-key",
	}

	agent := NewAgent(p)
	// Don't add any tools

	// findTool should return nil for unknown tools
	tool := agent.findTool("unknown_tool")
	if tool != nil {
		t.Error("expected nil for unknown tool")
	}
}

func TestAgent_ChatWithTool_OpenAI(t *testing.T) {
	rec, stop := newRecorder(t, "openai-tool-use")
	defer stop()

	p := Provider{
		Name:   OpenAI,
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	agent := NewAgent(p, WithHTTPClient(&http.Client{Transport: rec}))
	agent.AddTool(testWeatherTool())

	resp, err := agent.Chat(context.Background(), "What's the weather in Paris?")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if !containsIgnoreCase(resp.Text, "72") || !containsIgnoreCase(resp.Text, "sunny") {
		t.Errorf("expected response to contain tool result (72, sunny), got: %q", resp.Text)
	}
}

func TestAgent_ChatWithTool_Google(t *testing.T) {
	rec, stop := newRecorder(t, "google-tool-use")
	defer stop()

	p := Provider{
		Name:   Google,
		APIKey: googleAPIKey(),
	}

	agent := NewAgent(p, WithHTTPClient(&http.Client{Transport: rec}))
	agent.AddTool(testWeatherTool())

	resp, err := agent.Chat(context.Background(), "What's the weather in Paris?")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if !containsIgnoreCase(resp.Text, "72") || !containsIgnoreCase(resp.Text, "sunny") {
		t.Errorf("expected response to contain tool result (72, sunny), got: %q", resp.Text)
	}
}

func TestAgent_ChatWithTool_Grok(t *testing.T) {
	rec, stop := newRecorder(t, "grok-tool-use")
	defer stop()

	p := Provider{
		Name:   Grok,
		APIKey: os.Getenv("XAI_API_KEY"),
	}
	if p.APIKey == "" {
		p.APIKey = "test-key"
	}

	agent := NewAgent(p, WithHTTPClient(&http.Client{Transport: rec}))
	agent.AddTool(testWeatherTool())

	resp, err := agent.Chat(context.Background(), "What's the weather in Paris?")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if !containsIgnoreCase(resp.Text, "72") || !containsIgnoreCase(resp.Text, "sunny") {
		t.Errorf("expected response to contain tool result (72, sunny), got: %q", resp.Text)
	}
}

// mockToolTransport always returns a tool_use response to test max iterations.
type mockToolTransport struct {
	calls int
}

func (m *mockToolTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.calls++
	body := `{
		"content": [{"type": "tool_use", "id": "toolu_123", "name": "get_weather", "input": {"city": "Paris"}}],
		"stop_reason": "tool_use",
		"usage": {"input_tokens": 10, "output_tokens": 5}
	}`
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}

func TestAgent_ChatWithTool_MaxIterations(t *testing.T) {
	mock := &mockToolTransport{}
	p := Provider{
		Name:   Anthropic,
		APIKey: "test-key",
	}

	agent := NewAgent(p,
		WithHTTPClient(&http.Client{Transport: mock}),
		WithMaxToolIterations(3),
	)
	agent.AddTool(testWeatherTool())

	_, err := agent.Chat(context.Background(), "What's the weather?")
	if err == nil {
		t.Fatal("expected error for max iterations exceeded")
	}

	if !strings.Contains(err.Error(), "exceeded max tool iterations") {
		t.Errorf("expected 'exceeded max tool iterations' error, got: %v", err)
	}

	// Should have made exactly 3 requests (maxToolIterations)
	if mock.calls != 3 {
		t.Errorf("expected 3 API calls, got %d", mock.calls)
	}
}
