package agents

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

// WithMemoryPersistence enables automatic memory saving/loading
func WithMemoryPersistence(filepath string) AgentOption {
	return func(ca *ChatAgent) error {
		ca.memoryMode |= MemoryPersistence
		ca.memoryFile = filepath
		return ca.loadMemory()
	}
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

	// Send request to Google
	response, err := ca.sendRequest(options)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	// Add Google's response to conversation history if we have candidates
	if len(response.Candidates) > 0 {
		assistantContent := types.Content{
			Role:  "model",
			Parts: response.Candidates[0].Content.Parts,
		}
		ca.messages = append(ca.messages, assistantContent)
	}

	return &ChatResponse{
		Text: ca.extractTextResponse(response),
		Raw:  response,
	}, nil
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

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "marshaling request body",
			Err:       err,
		}
	}

	endpoint := fmt.Sprintf("%s?key=%s", types.Endpoint, ca.apiKey)
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
			Endpoint:   types.Endpoint,
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
