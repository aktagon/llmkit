package anthropic

// API Configuration Constants
const (
	// Endpoint is the Anthropic API endpoint for chat completions
	Endpoint = "https://api.anthropic.com/v1/messages"
	
	// Model is the default Claude model to use
	Model = "claude-sonnet-4-20250514"
	
	// AnthropicVersion is the API version header value
	AnthropicVersion = "2023-06-01"
	
	// MaxTokens is the default maximum tokens for responses
	MaxTokens = 4096
)

// JsonSchema represents the JSON schema structure for structured output
type JsonSchema struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Strict      bool        `json:"strict"`
	JsonSchema  interface{} `json:"schema"`
}

// ToolHandler executes tool logic and returns results
type ToolHandler func(input map[string]interface{}) (string, error)

// Tool represents a tool that can be called by Claude
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
	Handler     ToolHandler `json:"-"`
}

// ToolCall represents a tool use request from Claude
type ToolCall struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// Content represents message content (text or tool use/result)
type Content struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
}

// Message represents a conversation message
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// AnthropicResponse represents the API response structure
type AnthropicResponse struct {
	ID         string    `json:"id"`
	Model      string    `json:"model"`
	StopReason string    `json:"stop_reason"`
	Role       string    `json:"role"`
	Content    []Content `json:"content"`
}
