package types

// API Configuration Constants
const (
	// Model is the default Grok model to use
	Model = "grok-4-fast-non-reasoning"

	// Endpoint is the Grok API endpoint for chat completions
	Endpoint = "https://api.x.ai/v1/chat/completions"
)

// RequestSettings contains configuration for API requests
type RequestSettings struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// Message represents a conversation message in Grok format
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content,omitempty"`
}

// FileReference represents a file reference in message content
type FileReference struct {
	FileID string `json:"file_id"`
}

// ContentPart represents a part of message content (text or file)
type ContentPart struct {
	Type string         `json:"type"`
	Text string         `json:"text,omitempty"`
	File *FileReference `json:"file,omitempty"`
}

// Response represents the API response structure
type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// JsonSchema represents the JSON schema format for structured output
type JsonSchema struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Schema interface{} `json:"schema"`
	Strict bool        `json:"strict"`
}

// SchemaValidation represents the expected schema structure
type SchemaValidation struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Strict      bool        `json:"strict"`
	Schema      interface{} `json:"schema"`
}

// ResponseFormat represents the response format configuration
type ResponseFormat struct {
	Type       string     `json:"type"`
	JsonSchema JsonSchema `json:"json_schema"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}