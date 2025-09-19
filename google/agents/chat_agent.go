package agents

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/google/types"
	"github.com/aktagon/llmkit/httpclient"
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
	Schema       string       // JSON schema for structured output
	SystemPrompt string       // System prompt for this specific message
	Temperature  float64      // Temperature for response randomness (0.0-1.0, -1 = use default)
	MaxTokens    int          // Maximum tokens in response (0 = omit from request)
	Files        []types.File // File attachments to include in the message
}

// ChatResponse contains both extracted text and full raw response
type ChatResponse struct {
	Text string                // Extracted text response
	Raw  *types.GoogleResponse // Full API response
}

// ChatAgent maintains conversation state and handles tool execution
type ChatAgent struct {
	client     *http.Client
	apiKey     string
	model      string
	messages   []types.Content   // Conversation memory
	memory     map[string]string // Persistent key-value memory
	tools      map[string]types.Tool
	memoryMode MemoryMode        // Controls memory behavior
	memoryFile string            // Path for memory persistence
}

// New creates a new ChatAgent with optional configuration
func New(apiKey string, opts ...AgentOption) (*ChatAgent, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	agent := &ChatAgent{
		client:     httpclient.NewClient(),
		apiKey:     apiKey,
		model:      types.Model,
		messages:   make([]types.Content, 0),
		memory:     make(map[string]string),
		tools:      make(map[string]types.Tool),
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

// RegisterTool adds a tool that Gemini can use
func (ca *ChatAgent) RegisterTool(tool types.Tool) error {
	if tool.Name == "" {
		return &errors.ValidationError{
			Field:   "name",
			Message: "tool name is required",
		}
	}
	if tool.Description == "" {
		return &errors.ValidationError{
			Field:   "description",
			Message: "tool description is required",
		}
	}
	if tool.Parameters == nil {
		return &errors.ValidationError{
			Field:   "parameters",
			Message: "tool parameters schema is required",
		}
	}
	if tool.Handler == nil {
		return &errors.ValidationError{
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
	rememberTool := types.Tool{
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
	recallTool := types.Tool{
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

// Chat sends a message and handles Google's conversational API
func (ca *ChatAgent) Chat(message string, opts ...*ChatOptions) (*ChatResponse, error) {
	// Parse options
	var options *ChatOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}
	if ca.apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}
	if message == "" {
		return nil, &errors.ValidationError{
			Field:   "message",
			Message: "message cannot be empty",
		}
	}

	// Build user message parts with text and optional files
	userParts := []types.Part{{Text: message}}
	if options != nil && len(options.Files) > 0 {
		for _, file := range options.Files {
			userParts = append(userParts, types.Part{
				FileData: &types.FileData{
					FileURI:  file.URI,
					MimeType: file.MimeType,
				},
			})
		}
	}

	// Handle system prompt - Google doesn't have a system role, so we prepend it to the first user message
	if options != nil && options.SystemPrompt != "" && len(ca.messages) == 0 {
		systemPrompt := ca.buildSystemPromptWithMemory(options.SystemPrompt)
		userParts[0].Text = systemPrompt + "\n\n" + message
	} else if ca.memoryMode&MemoryContext > 0 && len(ca.memory) > 0 && len(ca.messages) == 0 {
		// Include memory even without explicit system prompt
		systemPrompt := ca.buildSystemPromptWithMemory("")
		userParts[0].Text = systemPrompt + "\n\n" + message
	}

	// Add user message to conversation history
	userContent := types.Content{
		Role:  "user",
		Parts: userParts,
	}
	ca.messages = append(ca.messages, userContent)

	// Continue conversation until we get a final response
	for {
		// Send request to Gemini
		response, err := ca.sendRequest(options)
		if err != nil {
			return nil, fmt.Errorf("sending request: %w", err)
		}

		// Add Gemini's response to conversation history
		if len(response.Candidates) > 0 {
			assistantContent := types.Content{
				Role:  "model",
				Parts: response.Candidates[0].Content.Parts,
			}
			ca.messages = append(ca.messages, assistantContent)
		}

		// Check if Gemini wants to use tools
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

// sendRequest sends the current conversation to Google
func (ca *ChatAgent) sendRequest(options *ChatOptions) (*types.GoogleResponse, error) {
	// Prepare generation config
	var genConfig *types.RequestSettings
	if options != nil && (options.MaxTokens > 0 || options.Temperature >= 0.0 || options.Schema != "") {
		genConfig = &types.RequestSettings{}
		if options.MaxTokens > 0 {
			genConfig.MaxTokens = options.MaxTokens
		}
		if options.Temperature >= 0.0 {
			genConfig.Temperature = options.Temperature
		}

		// Handle structured output
		if options.Schema != "" && options.Schema != "null" {
			var schema interface{}
			if err := json.Unmarshal([]byte(options.Schema), &schema); err != nil {
				return nil, &errors.SchemaError{
					Field:   "json",
					Message: "invalid JSON: " + err.Error(),
				}
			}
			genConfig.ResponseMimeType = "application/json"
			genConfig.ResponseSchema = schema
		}
	}

	requestBody := types.GoogleRequest{
		Contents:         ca.messages,
		GenerationConfig: genConfig,
	}

	// Add tools if any are registered
	if len(ca.tools) > 0 {
		var toolDeclarations []types.FunctionDeclaration
		for _, tool := range ca.tools {
			toolDeclarations = append(toolDeclarations, types.FunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			})
		}
		requestBody.Tools = []types.ToolConfig{{
			FunctionDeclarations: toolDeclarations,
		}}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "marshaling request body",
			Err:       err,
		}
	}

	endpoint := fmt.Sprintf("%s/%s:generateContent?key=%s", types.BaseEndpoint, ca.model, ca.apiKey)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ca.client.Do(req)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "sending request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "reading response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &errors.APIError{
			Provider:   "Google",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   endpoint,
		}
	}

	var googleResp types.GoogleResponse
	if err := json.Unmarshal(bodyText, &googleResp); err != nil {
		return nil, &errors.RequestError{
			Operation: "parsing response",
			Err:       err,
		}
	}

	return &googleResp, nil
}

// extractToolCalls extracts tool calls from Gemini's response
func (ca *ChatAgent) extractToolCalls(response *types.GoogleResponse) []types.FunctionCall {
	var toolCalls []types.FunctionCall
	if len(response.Candidates) > 0 {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, *part.FunctionCall)
			}
		}
	}
	return toolCalls
}

// extractTextResponse extracts text content from Google's response
func (ca *ChatAgent) extractTextResponse(response *types.GoogleResponse) string {
	var textParts []string
	if len(response.Candidates) > 0 {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
		}
	}
	return strings.Join(textParts, " ")
}

// executeToolCalls executes tool calls and adds results to conversation
func (ca *ChatAgent) executeToolCalls(toolCalls []types.FunctionCall) error {
	var toolResults []types.Part

	for _, toolCall := range toolCalls {
		tool, exists := ca.tools[toolCall.Name]
		if !exists {
			return &errors.ValidationError{
				Field:   "tool",
				Message: fmt.Sprintf("tool '%s' not found", toolCall.Name),
			}
		}

		// Execute the tool
		start := time.Now()
		result, err := tool.Handler(toolCall.Args)
		duration := time.Since(start)

		slog.Debug("Tool execution",
			slog.String("tool", toolCall.Name),
			slog.Any("input", toolCall.Args),
			slog.Duration("duration", duration),
			slog.String("result", result))

		if err != nil {
			return fmt.Errorf("executing tool '%s': %w", toolCall.Name, err)
		}

		// Add tool result to results using Google's function response format
		toolResults = append(toolResults, types.Part{
			FunctionResponse: &types.FunctionResponsePart{
				Name: toolCall.Name,
				Response: struct {
					Result interface{} `json:"result"`
				}{
					Result: result,
				},
			},
		})
	}

	// Add tool results as a user message
	toolMessage := types.Content{
		Role:  "user",
		Parts: toolResults,
	}
	ca.messages = append(ca.messages, toolMessage)

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
	ca.messages = make([]types.Content, 0)
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
