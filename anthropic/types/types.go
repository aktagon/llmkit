package types

// API Configuration Constants
const (
	// Endpoint is the Anthropic API endpoint for chat completions
	Endpoint = "https://api.anthropic.com/v1/messages"

	// FilesEndpoint is the Anthropic API endpoint for file uploads
	FilesEndpoint = "https://api.anthropic.com/v1/files"

	// Model is the default Claude model to use
	Model = "claude-sonnet-4-20250514"

	// AnthropicVersion is the API version header value
	AnthropicVersion = "2023-06-01"

	// FilesBetaHeader is the beta header value for files API
	FilesBetaHeader = "files-api-2025-04-14"
)

// RequestSettings contains configuration for API requests
type RequestSettings struct {
	Model       string  // Model to use for the request (defaults to types.Model if empty)
	MaxTokens   int
	Temperature float64
	TopK        int     // Only sample from the top K options for each subsequent token
	TopP        float64 // Use nucleus sampling
}

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

// Content represents message content (text, tool use/result, or file)
type Content struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	Source    *FileSource            `json:"source,omitempty"`
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

// Usage represents token usage information
type Usage struct {
	InputTokens              int    `json:"input_tokens"`
	CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
	OutputTokens             int    `json:"output_tokens"`
	ServiceTier              string `json:"service_tier"`
}

// File represents an uploaded file
type File struct {
	ID string `json:"id"`
}

// FileSource represents a file reference in content
type FileSource struct {
	Type   string `json:"type"`
	FileID string `json:"file_id"`
}

// AnthropicResponse represents the API response structure
type AnthropicResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason"`
	StopSequence *string   `json:"stop_sequence"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Usage        Usage     `json:"usage"`
}
