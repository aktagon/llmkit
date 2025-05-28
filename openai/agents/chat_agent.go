package agents

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aktagon/llmkit"
	"github.com/aktagon/llmkit/openai"
)

// MemoryMode controls memory behavior using bitwise flags
type MemoryMode int

const (
	MemoryDisabled    MemoryMode = 0
	MemoryContext     MemoryMode = 1 << 0 // Auto-include in system prompts
	MemoryTools       MemoryMode = 1 << 1 // Expose remember/recall as tools
	MemoryPersistence MemoryMode = 1 << 2 // Auto-save to disk
)

// AgentOption configures the ChatAgent
type AgentOption func(*ChatAgent) error

// ChatOptions provides optional parameters for chat requests
type ChatOptions struct {
	Schema       string  // JSON schema for structured output
	SystemPrompt string  // System prompt for this specific message
	Temperature  float64 // Temperature for response randomness (0.0-1.0, -1 = use default)
	MaxTokens    int     // Maximum tokens in response (0 = use default)
}

// ChatResponse contains both extracted text and full raw response
type ChatResponse struct {
	Text string         // Extracted text response
	Raw  *openai.Response // Full API response
}

// ChatAgent maintains conversation state and handles tool execution
type ChatAgent struct {
	client     *http.Client
	apiKey     string
	model      string
	messages   []openai.Message          // Conversation memory
	memory     map[string]string         // Persistent key-value memory
	tools      map[string]openai.Tool
	memoryMode MemoryMode                // Controls memory behavior
	memoryFile string                    // Path for memory persistence
}

// New creates a new ChatAgent with optional configuration
func New(apiKey string, opts ...AgentOption) (*ChatAgent, error) {
	if apiKey == "" {
		return nil, &llmkit.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	agent := &ChatAgent{
		client:     &http.Client{},
		apiKey:     apiKey,
		model:      openai.Model,
		messages:   make([]openai.Message, 0),
		memory:     make(map[string]string),
		tools:      make(map[string]openai.Tool),
		memoryMode: MemoryDisabled,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(agent); err != nil {
			return nil, err
		}
	}

	return agent, nil
}

// WithMemoryContext enables automatic memory inclusion in system prompts
func WithMemoryContext() AgentOption {
	return func(ca *ChatAgent) error {
		ca.memoryMode |= MemoryContext
		return nil
	}
}

// WithMemoryTools enables memory tools (remember_fact, recall_fact)
func WithMemoryTools() AgentOption {
	return func(ca *ChatAgent) error {
		ca.memoryMode |= MemoryTools
		return ca.registerMemoryTools()
	}
}

// WithMemoryPersistence enables automatic memory saving/loading
func WithMemoryPersistence(filepath string) AgentOption {
	return func(ca *ChatAgent) error {
		ca.memoryMode |= MemoryPersistence
		ca.memoryFile = filepath
		return ca.loadMemory()
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

// registerMemoryTools adds remember_fact and recall_fact tools
func (ca *ChatAgent) registerMemoryTools() error {
	// Remember tool
	rememberTool := openai.Tool{
		Name:        "remember_fact",
		Description: "Store important information about the user for future conversations",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "What to remember (e.g., 'favorite_color', 'job_title', 'preference')",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "The information to remember",
				},
			},
			"required":             []string{"key", "value"},
			"additionalProperties": false,
		},
		Handler: func(input map[string]interface{}) (string, error) {
			key := input["key"].(string)
			value := input["value"].(string)
			err := ca.Remember(key, value)
			if err != nil {
				return "", fmt.Errorf("failed to remember %s: %w", key, err)
			}
			return fmt.Sprintf("I'll remember that %s: %s", key, value), nil
		},
	}

	// Recall tool
	recallTool := openai.Tool{
		Name:        "recall_fact",
		Description: "Retrieve previously stored information about the user",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "What to recall (e.g., 'favorite_color', 'job_title', 'preference')",
				},
			},
			"required":             []string{"key"},
			"additionalProperties": false,
		},
		Handler: func(input map[string]interface{}) (string, error) {
			key := input["key"].(string)
			if value, exists := ca.memory[key]; exists {
				return value, nil
			}
			return fmt.Sprintf("I don't have any information stored about %s", key), nil
		},
	}

	ca.RegisterTool(rememberTool)
	ca.RegisterTool(recallTool)
	return nil
}

// buildSystemPromptWithMemory combines memory context with user system prompt
func (ca *ChatAgent) buildSystemPromptWithMemory(userSystemPrompt string) string {
	var parts []string

	// Add memory context if enabled and memory exists
	if ca.memoryMode&MemoryContext > 0 && len(ca.memory) > 0 {
		memoryContext := "<memory>"
		for key, value := range ca.memory {
			memoryContext += fmt.Sprintf("\n%s: %s", key, value)
		}
		memoryContext += "\n</memory>"
		parts = append(parts, memoryContext)
	}

	// Add user's system prompt
	if userSystemPrompt != "" {
		parts = append(parts, userSystemPrompt)
	}

	return strings.Join(parts, "\n\n")
}

// Chat sends a message and handles tool execution automatically
func (ca *ChatAgent) Chat(message string, opts ...*ChatOptions) (*ChatResponse, error) {
	// Parse options
	var options *ChatOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}
	if ca.apiKey == "" {
		return nil, &llmkit.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}
	if message == "" {
		return nil, &llmkit.ValidationError{
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
		response, err := ca.sendRequest(options)
		if err != nil {
			return nil, fmt.Errorf("sending request: %w", err)
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
			// No tool calls - return the response
			return &ChatResponse{
				Text: ca.extractTextResponse(response),
				Raw:  response,
			}, nil
		}

		// Execute tools and add results to conversation
		err = ca.executeToolCalls(toolCalls)
		if err != nil {
			return nil, fmt.Errorf("executing tools: %w", err)
		}
	}
}

// sendRequest sends the current conversation to GPT
func (ca *ChatAgent) sendRequest(options *ChatOptions) (*openai.Response, error) {
	// Prepare messages with system prompt if needed
	messages := make([]openai.Message, 0, len(ca.messages)+1)

	// Add system message if we have memory context or system prompt
	systemPrompt := ""
	if options != nil && options.SystemPrompt != "" {
		systemPrompt = ca.buildSystemPromptWithMemory(options.SystemPrompt)
	} else if ca.memoryMode&MemoryContext > 0 && len(ca.memory) > 0 {
		systemPrompt = ca.buildSystemPromptWithMemory("")
	}

	if systemPrompt != "" {
		systemMessage := openai.Message{
			Role:    "system",
			Content: systemPrompt,
		}
		messages = append(messages, systemMessage)
	}

	// Add conversation messages
	messages = append(messages, ca.messages...)

	// Handle schema by modifying the last user message if schema is provided
	if options != nil && options.Schema != "" {
		// Find the last user message and append schema instructions
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				schemaInstructions := fmt.Sprintf("\n\nYou must output only the raw JSON without further explanation or formatting. Use the following JSON schema for the output format:\n\n%s", options.Schema)
				messages[i].Content += schemaInstructions
				break
			}
		}
	}

	requestBody := map[string]interface{}{
		"model":    ca.model,
		"messages": messages,
	}

	// Override max_tokens if provided in options
	if options != nil && options.MaxTokens > 0 {
		requestBody["max_tokens"] = options.MaxTokens
	}

	// Add temperature if provided (use -1 as sentinel for "not set")
	if options != nil && options.Temperature >= 0.0 {
		requestBody["temperature"] = options.Temperature
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

// Remember stores a key-value pair in persistent memory
func (ca *ChatAgent) Remember(key, value string) error {
	ca.memory[key] = value
	return ca.saveMemory()
}

// Recall retrieves a value from persistent memory
func (ca *ChatAgent) Recall(key string) (string, bool) {
	value, exists := ca.memory[key]
	return value, exists
}

// Forget removes a key from persistent memory
func (ca *ChatAgent) Forget(key string) error {
	delete(ca.memory, key)
	return ca.saveMemory()
}

// ClearMemory removes all keys from persistent memory
func (ca *ChatAgent) ClearMemory() error {
	ca.memory = make(map[string]string)
	return ca.saveMemory()
}

// GetMemory returns a copy of all memory key-value pairs
func (ca *ChatAgent) GetMemory() map[string]string {
	memory := make(map[string]string)
	for k, v := range ca.memory {
		memory[k] = v
	}
	return memory
}

// Reset clears conversation history and optionally persistent memory
func (ca *ChatAgent) Reset(clearMemory bool) error {
	ca.messages = make([]openai.Message, 0)
	if clearMemory {
		return ca.ClearMemory()
	}
	return nil
}

// loadMemory loads memory from disk if persistence is enabled
func (ca *ChatAgent) loadMemory() error {
	if ca.memoryMode&MemoryPersistence == 0 || ca.memoryFile == "" {
		return nil
	}

	data, err := os.ReadFile(ca.memoryFile)
	if os.IsNotExist(err) {
		return nil // No memory file yet
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &ca.memory)
}

// saveMemory saves memory to disk if persistence is enabled
func (ca *ChatAgent) saveMemory() error {
	if ca.memoryMode&MemoryPersistence == 0 || ca.memoryFile == "" {
		return nil
	}

	data, err := json.Marshal(ca.memory)
	if err != nil {
		return err
	}

	return os.WriteFile(ca.memoryFile, data, 0644)
}
