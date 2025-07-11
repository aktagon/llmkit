package google

// API Configuration Constants
const (
	// Endpoint is the Google Generative AI API endpoint
	Endpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

	// Model is the default Gemini model to use
	Model = "gemini-2.5-flash"
)

// Content represents a content part in the request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a text part of the content
type Part struct {
	Text string `json:"text"`
}

// GenerationConfig configures the generation parameters
type GenerationConfig struct {
	ResponseMimeType string      `json:"responseMimeType,omitempty"`
	ResponseSchema   interface{} `json:"responseSchema,omitempty"`
}

// GoogleRequest represents the request body for Google's API
type GoogleRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content struct {
		Parts []Part `json:"parts"`
		Role  string `json:"role"`
	} `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

// SafetyRating represents safety assessment
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// UsageMetadata represents token usage information
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GoogleResponse represents the API response structure
type GoogleResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

// JsonSchema represents Google's JSON schema structure for structured output
type JsonSchema struct {
	Type        string                 `json:"type"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Items       interface{}            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}