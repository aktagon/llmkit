package types

// API Configuration Constants
const (
	// Endpoint is the Google Generative AI API endpoint
	Endpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

	// FilesEndpoint is the Google Files API endpoint for uploads
	FilesEndpoint = "https://generativelanguage.googleapis.com/upload/v1beta/files"

	// Model is the default Gemini model to use
	Model = "gemini-2.5-flash"
)

// RequestSettings contains configuration for API requests
type RequestSettings struct {
	MaxTokens        int         `json:"maxOutputTokens,omitempty"`
	Temperature      float64     `json:"temperature,omitempty"`
	ResponseMimeType string      `json:"responseMimeType,omitempty"`
	ResponseSchema   interface{} `json:"responseSchema,omitempty"`
}

// Content represents a content part in the request
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

// Part represents a text or file part of the content
type Part struct {
	Text     string    `json:"text,omitempty"`
	FileData *FileData `json:"file_data,omitempty"`
}

// GoogleRequest represents the request body for Google's API
type GoogleRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig *RequestSettings `json:"generationConfig,omitempty"`
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

// File represents an uploaded file
type File struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	MimeType    string `json:"mime_type"`
	SizeBytes   int64  `json:"size_bytes"`
	State       string `json:"state"`
}

// FileData represents a file reference in content
type FileData struct {
	MimeType string `json:"mime_type"`
	FileURI  string `json:"file_uri"`
}

// FileUploadRequest represents the initial upload request
type FileUploadRequest struct {
	File FileUploadInfo `json:"file"`
}

// FileUploadInfo represents file metadata for upload
type FileUploadInfo struct {
	DisplayName string `json:"display_name"`
}

// FileUploadResponse wraps the uploaded file
type FileUploadResponse struct {
	File File `json:"file"`
}