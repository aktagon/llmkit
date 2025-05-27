package agents

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aktagon/llmkit"
	"github.com/aktagon/llmkit/openai"
)

// ChatAgent maintains conversation state and handles tool execution
type ChatAgent struct {
	client   *http.Client
	apiKey   string
	model    string
	messages []openai.Message
	tools    map[string]openai.Tool
}

// New creates a new ChatAgent
func New(apiKey string) *ChatAgent {
	if apiKey == "" {
		return nil
	}

	return &ChatAgent{
		client:   &http.Client{},
		apiKey:   apiKey,
		model:    openai.Model,
		messages: make([]openai.Message, 0),
		tools:    make(map[string]openai.Tool),
	}
}

// RegisterTool adds a tool that GPT can use
func (ca *ChatAgent) RegisterTool(tool openai.Tool) error {
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
	if tool.Parameters == nil {
		return &llmkit.ValidationError{
			Field:   "parameters",
			Message: "tool parameters schema is required",
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
	userMessage := openai.Message{
		Role:    "user",
		Content: message,
	}
	ca.messages = append(ca.messages, userMessage)

	// Continue conversation until we get a final response
	for {
		// Send request to GPT
		response, err := ca.sendRequest()
		if err != nil {
			return "", fmt.Errorf("sending request: %w", err)
		}

		// Add GPT's response to conversation history
		assistantMessage := openai.Message{
			Role:         "assistant",
			Content:      response.Choices[0].Message.Content,
			FunctionCall: response.Choices[0].Message.FunctionCall,
		}
		ca.messages = append(ca.messages, assistantMessage)

		// Check if GPT wants to use tools
		toolCalls := ca.extractToolCalls(response)
		if len(toolCalls) == 0 {
			// No tool calls - return the text response
			return ca.extractTextResponse(response), nil
		}

		// Execute tools and add results to conversation
		err = ca.executeToolCalls(toolCalls)
		if err != nil {
			return "", fmt.Errorf("executing tools: %w", err)
		}
	}
}

// sendRequest sends the current conversation to GPT
func (ca *ChatAgent) sendRequest() (*openai.Response, error) {
	requestBody := map[string]interface{}{
		"model":    ca.model,
		"messages": ca.messages,
	}

	// Add functions if any are registered
	if len(ca.tools) > 0 {
		functions := make([]openai.Function, 0, len(ca.tools))
		for _, tool := range ca.tools {
			functions = append(functions, openai.Function{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			})
		}
		requestBody["functions"] = functions
		requestBody["function_call"] = "auto"
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "marshaling request body",
			Err:       err,
		}
	}

	req, err := http.NewRequest("POST", openai.EndpointCompletions, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+ca.apiKey)
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
			Provider:   "OpenAI",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   openai.EndpointCompletions,
		}
	}

	var openaiResp openai.Response
	if err := json.Unmarshal(bodyText, &openaiResp); err != nil {
		return nil, &llmkit.RequestError{
			Operation: "parsing response",
			Err:       err,
		}
	}

	return &openaiResp, nil
}

// extractToolCalls extracts tool calls from GPT's response
func (ca *ChatAgent) extractToolCalls(response *openai.Response) []openai.FunctionCall {
	var toolCalls []openai.FunctionCall
	if len(response.Choices) > 0 && response.Choices[0].Message.FunctionCall != nil {
		toolCalls = append(toolCalls, *response.Choices[0].Message.FunctionCall)
	}
	return toolCalls
}

// extractTextResponse extracts text content from GPT's response
func (ca *ChatAgent) extractTextResponse(response *openai.Response) string {
	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content
	}
	return ""
}

// executeToolCalls executes tool calls and adds results to conversation
func (ca *ChatAgent) executeToolCalls(toolCalls []openai.FunctionCall) error {
	for _, toolCall := range toolCalls {
		tool, exists := ca.tools[toolCall.Name]
		if !exists {
			return &llmkit.ValidationError{
				Field:   "tool",
				Message: fmt.Sprintf("tool '%s' not found", toolCall.Name),
			}
		}

		// Parse arguments from JSON string
		var arguments map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Arguments), &arguments); err != nil {
			return &llmkit.RequestError{
				Operation: "parsing tool arguments",
				Err:       err,
			}
		}

		// Execute the tool
		result, err := tool.Handler(arguments)
		if err != nil {
			return fmt.Errorf("executing tool '%s': %w", toolCall.Name, err)
		}

		// Add tool result as a function message
		functionMessage := openai.Message{
			Role:    "function",
			Name:    toolCall.Name,
			Content: result,
		}
		ca.messages = append(ca.messages, functionMessage)
	}

	return nil
}

// Reset clears conversation history
func (ca *ChatAgent) Reset() {
	ca.messages = make([]openai.Message, 0)
}
