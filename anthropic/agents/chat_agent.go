package agents

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aktagon/llmkit"
	"github.com/aktagon/llmkit/anthropic"
)

// ChatAgent maintains conversation state and handles tool execution
type ChatAgent struct {
	client   *http.Client
	apiKey   string
	model    string
	messages []anthropic.Message
	tools    map[string]anthropic.Tool
}

// New creates a new ChatAgent
func New(apiKey string) *ChatAgent {
	if apiKey == "" {
		return nil
	}

	return &ChatAgent{
		client:   &http.Client{},
		apiKey:   apiKey,
		model:    anthropic.Model,
		messages: make([]anthropic.Message, 0),
		tools:    make(map[string]anthropic.Tool),
	}
}

// RegisterTool adds a tool that Claude can use
func (ca *ChatAgent) RegisterTool(tool anthropic.Tool) error {
	if tool.Name == "" {
		return &llmkit.ValidationError{
			Field:   "name",
			Message: "tool name is required",
		}
	}
	if tool.Description == "" {
		return &llmkit.ValidationError{
			Field:   "description",
			Message: "tool description is required",
		}
	}
	if tool.InputSchema == nil {
		return &llmkit.ValidationError{
			Field:   "input_schema",
			Message: "tool input schema is required",
		}
	}
	if tool.Handler == nil {
		return &llmkit.ValidationError{
			Field:   "handler",
			Message: "tool handler is required",
		}
	}

	ca.tools[tool.Name] = tool
	return nil
}

// Chat sends a message and handles tool execution automatically
func (ca *ChatAgent) Chat(message string) (string, error) {
	if ca.apiKey == "" {
		return "", &llmkit.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}
	if message == "" {
		return "", &llmkit.ValidationError{
			Field:   "message",
			Message: "message cannot be empty",
		}
	}

	// Add user message to conversation history
	userMessage := anthropic.Message{
		Role: "user",
		Content: []anthropic.Content{{
			Type: "text",
			Text: message,
		}},
	}
	ca.messages = append(ca.messages, userMessage)

	// Continue conversation until we get a final response
	for {
		// Send request to Claude
		response, err := ca.sendRequest()
		if err != nil {
			return "", fmt.Errorf("sending request: %w", err)
		}

		// Add Claude's response to conversation history
		assistantMessage := anthropic.Message{
			Role:    "assistant",
			Content: response.Content,
		}
		ca.messages = append(ca.messages, assistantMessage)

		// Check if Claude wants to use tools
		toolCalls := ca.extractToolCalls(response.Content)
		if len(toolCalls) == 0 {
			// No tool calls - return the text response
			return ca.extractTextResponse(response.Content), nil
		}

		// Execute tools and add results to conversation
		err = ca.executeToolCalls(toolCalls)
		if err != nil {
			return "", fmt.Errorf("executing tools: %w", err)
		}
	}
}

// sendRequest sends the current conversation to Claude
func (ca *ChatAgent) sendRequest() (*anthropic.AnthropicResponse, error) {
	requestBody := map[string]interface{}{
		"model":      ca.model,
		"max_tokens": anthropic.MaxTokens,
		"messages":   ca.messages,
	}

	// Add tools if any are registered
	if len(ca.tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(ca.tools))
		for _, tool := range ca.tools {
			tools = append(tools, map[string]interface{}{
				"name":         tool.Name,
				"description":  tool.Description,
				"input_schema": tool.InputSchema,
			})
		}
		requestBody["tools"] = tools
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "marshaling request body",
			Err:       err,
		}
	}

	req, err := http.NewRequest("POST", anthropic.Endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("x-api-key", ca.apiKey)
	req.Header.Set("anthropic-version", anthropic.AnthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := ca.client.Do(req)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "sending request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "reading response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &llmkit.APIError{
			Provider:   "Anthropic",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   anthropic.Endpoint,
		}
	}

	var anthropicResp anthropic.AnthropicResponse
	if err := json.Unmarshal(bodyText, &anthropicResp); err != nil {
		return nil, &llmkit.RequestError{
			Operation: "parsing response",
			Err:       err,
		}
	}

	return &anthropicResp, nil
}

// extractToolCalls extracts tool calls from Claude's response content
func (ca *ChatAgent) extractToolCalls(content []anthropic.Content) []anthropic.ToolCall {
	var toolCalls []anthropic.ToolCall
	for _, c := range content {
		if c.Type == "tool_use" {
			toolCalls = append(toolCalls, anthropic.ToolCall{
				ID:    c.ID,
				Name:  c.Name,
				Input: c.Input,
			})
		}
	}
	return toolCalls
}

// extractTextResponse extracts text content from Claude's response
func (ca *ChatAgent) extractTextResponse(content []anthropic.Content) string {
	var textParts []string
	for _, c := range content {
		if c.Type == "text" && c.Text != "" {
			textParts = append(textParts, c.Text)
		}
	}
	return strings.Join(textParts, " ")
}

// executeToolCalls executes tool calls and adds results to conversation
func (ca *ChatAgent) executeToolCalls(toolCalls []anthropic.ToolCall) error {
	var toolResults []anthropic.Content

	for _, toolCall := range toolCalls {
		tool, exists := ca.tools[toolCall.Name]
		if !exists {
			return &llmkit.ValidationError{
				Field:   "tool",
				Message: fmt.Sprintf("tool '%s' not found", toolCall.Name),
			}
		}

		// Execute the tool
		result, err := tool.Handler(toolCall.Input)
		if err != nil {
			return fmt.Errorf("executing tool '%s': %w", toolCall.Name, err)
		}

		// Add tool result to results
		toolResults = append(toolResults, anthropic.Content{
			Type:      "tool_result",
			ToolUseID: toolCall.ID,
			Content:   result,
		})
	}

	// Add tool results as a user message
	toolMessage := anthropic.Message{
		Role:    "user",
		Content: toolResults,
	}
	ca.messages = append(ca.messages, toolMessage)

	return nil
}
