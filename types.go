package llmkit

// Provider constants
const (
	Anthropic = "anthropic"
	OpenAI    = "openai"
	Google    = "google"
	Grok      = "grok"
)

// Default models per provider
var defaultModels = map[string]string{
	Anthropic: "claude-sonnet-4-5",
	OpenAI:    "gpt-4o-2024-08-06",
	Google:    "gemini-2.5-flash",
	Grok:      "grok-3-fast",
}

// Provider configures which LLM to use.
type Provider struct {
	Name    string // "anthropic", "openai", "google", "grok"
	APIKey  string
	Model   string // optional, uses default if empty
	BaseURL string // optional, overrides default API endpoint
}

// model returns the configured model or the default for the provider.
func (p Provider) model() string {
	if p.Model != "" {
		return p.Model
	}
	return defaultModels[p.Name]
}

// Default base URLs per provider
var defaultBaseURLs = map[string]string{
	Anthropic: "https://api.anthropic.com",
	OpenAI:    "https://api.openai.com",
	Google:    "https://generativelanguage.googleapis.com",
	Grok:      "https://api.x.ai",
}

// buildURL constructs the full URL using custom BaseURL or default.
func (p Provider) buildURL(path string) string {
	base := p.BaseURL
	if base == "" {
		base = defaultBaseURLs[p.Name]
	}
	return base + path
}

// Message represents a conversation message.
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Request contains the input for an LLM call.
type Request struct {
	System   string    // system prompt
	User     string    // user message (for single-turn)
	Messages []Message // conversation history (for multi-turn)
	Schema   string    // JSON schema for structured output (optional)
	Files    []File    // file attachments (optional)
	Images   []Image   // image inputs (optional)
}

// Response contains the LLM output.
type Response struct {
	Text   string
	Tokens Usage
}

// Usage tracks token consumption.
type Usage struct {
	Input  int
	Output int
}

// File represents an uploaded file reference.
type File struct {
	ID       string
	URI      string
	MimeType string
	Name     string
}

// Image represents an image input.
type Image struct {
	URL      string // URL or base64 data URI
	MimeType string
	Detail   string // "auto", "low", "high" (provider-specific)
}

// Tool defines a function the LLM can call.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
	Run         func(map[string]any) (string, error)
}
