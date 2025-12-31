package llmkit

import (
	"context"
	"encoding/json"
	"strings"
)

const anthropicChatPath = "/v1/messages"

type anthropicRequest struct {
	Model         string                 `json:"model"`
	MaxTokens     int                    `json:"max_tokens"`
	System        string                 `json:"system,omitempty"`
	Messages      []anthropicMessage     `json:"messages"`
	Tools         []anthropicTool        `json:"tools,omitempty"`
	OutputFormat  *anthropicOutputFormat `json:"output_format,omitempty"`
	Temperature   *float64               `json:"temperature,omitempty"`
	TopP          *float64               `json:"top_p,omitempty"`
	TopK          *int                   `json:"top_k,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`
	Thinking      *anthropicThinking     `json:"thinking,omitempty"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

type anthropicOutputFormat struct {
	Type   string `json:"type"`
	Schema any    `json:"schema"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type      string           `json:"type"`
	Text      string           `json:"text,omitempty"`
	Source    *anthropicSource `json:"source,omitempty"`
	ID        string           `json:"id,omitempty"`          // for tool_use
	Name      string           `json:"name,omitempty"`        // for tool_use
	Input     map[string]any   `json:"input,omitempty"`       // for tool_use
	ToolUseID string           `json:"tool_use_id,omitempty"` // for tool_result
	Content   string           `json:"content,omitempty"`     // for tool_result
}

type anthropicSource struct {
	Type      string `json:"type"`                 // "base64", "url", or "file"
	MediaType string `json:"media_type,omitempty"` // for base64
	Data      string `json:"data,omitempty"`       // for base64
	URL       string `json:"url,omitempty"`        // for url
	FileID    string `json:"file_id,omitempty"`    // for file
}


type anthropicResponse struct {
	Content []struct {
		Type  string         `json:"type"`
		Text  string         `json:"text,omitempty"`
		ID    string         `json:"id,omitempty"`
		Name  string         `json:"name,omitempty"`
		Input map[string]any `json:"input,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func promptAnthropic(ctx context.Context, p Provider, req Request, o *options) (Response, error) {
	maxTokens := 4096
	if o.maxTokens != nil {
		maxTokens = *o.maxTokens
	}

	// Build content array
	content := buildAnthropicContent(req)

	// Build messages array
	var messages []anthropicMessage
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			messages = append(messages, anthropicMessage{
				Role:    m.Role,
				Content: []anthropicContent{{Type: "text", Text: m.Content}},
			})
		}
	} else {
		messages = []anthropicMessage{{Role: "user", Content: content}}
	}

	payload := anthropicRequest{
		Model:         p.model(),
		MaxTokens:     maxTokens,
		System:        req.System,
		Temperature:   o.temperature,
		TopP:          o.topP,
		TopK:          o.topK,
		StopSequences: o.stopSequences,
		Messages:      messages,
	}

	if o.thinkingBudget != nil {
		payload.Thinking = &anthropicThinking{
			Type:         "enabled",
			BudgetTokens: *o.thinkingBudget,
		}
	}

	headers := map[string]string{
		"x-api-key":         p.APIKey,
		"anthropic-version": "2023-06-01",
	}

	if req.Schema != "" {
		var schema any
		if err := json.Unmarshal([]byte(req.Schema), &schema); err != nil {
			return Response{}, err
		}
		payload.OutputFormat = &anthropicOutputFormat{
			Type:   "json_schema",
			Schema: schema,
		}
		headers["anthropic-beta"] = "structured-outputs-2025-11-13"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, p.buildURL(anthropicChatPath), body, headers)
	if err != nil {
		return Response{}, err
	}

	if statusCode >= 400 {
		return Response{}, parseError(Anthropic, statusCode, respBody, nil)
	}

	var resp anthropicResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return Response{}, err
	}

	text := ""
	if len(resp.Content) > 0 {
		text = resp.Content[0].Text
	}

	return Response{
		Text: text,
		Tokens: Usage{
			Input:  resp.Usage.InputTokens,
			Output: resp.Usage.OutputTokens,
		},
	}, nil
}

// buildAnthropicContent creates content array from request.
func buildAnthropicContent(req Request) []anthropicContent {
	var content []anthropicContent

	// Add files first
	for _, f := range req.Files {
		content = append(content, anthropicContent{
			Type: "document",
			Source: &anthropicSource{
				Type:   "file",
				FileID: f.ID,
			},
		})
	}

	// Add images
	for _, img := range req.Images {
		c := anthropicContent{Type: "image"}
		if strings.HasPrefix(img.URL, "data:") {
			c.Source = &anthropicSource{
				Type:      "base64",
				MediaType: img.MimeType,
				Data:      extractBase64Data(img.URL),
			}
		} else {
			c.Source = &anthropicSource{
				Type: "url",
				URL:  img.URL,
			}
		}
		content = append(content, c)
	}

	// Add text
	if req.User != "" {
		content = append(content, anthropicContent{Type: "text", Text: req.User})
	}

	return content
}

// extractBase64Data extracts base64 data from data URI.
func extractBase64Data(dataURI string) string {
	if idx := strings.Index(dataURI, ","); idx != -1 {
		return dataURI[idx+1:]
	}
	return dataURI
}

// sendAnthropicWithTools sends a request with tools and returns tool calls.
func sendAnthropicWithTools(ctx context.Context, p Provider, msgs []message, system string, tools []Tool, o *options) (string, []toolCall, Usage, error) {
	maxTokens := 4096
	if o.maxTokens != nil {
		maxTokens = *o.maxTokens
	}

	// Build messages
	var messages []anthropicMessage
	for _, m := range msgs {
		msg := anthropicMessage{Role: m.role}
		if m.toolResult != nil {
			// Tool result message
			msg.Role = "user"
			msg.Content = []anthropicContent{{
				Type:      "tool_result",
				ToolUseID: m.toolResult.toolUseID,
				Content:   m.toolResult.content,
			}}
		} else if len(m.toolCalls) > 0 {
			// Assistant message with tool calls
			for _, tc := range m.toolCalls {
				msg.Content = append(msg.Content, anthropicContent{
					Type:  "tool_use",
					ID:    tc.id,
					Name:  tc.name,
					Input: tc.input,
				})
			}
		} else {
			// Regular text message
			msg.Content = []anthropicContent{{Type: "text", Text: m.content}}
		}
		messages = append(messages, msg)
	}

	// Build tools
	var anthropicTools []anthropicTool
	for _, t := range tools {
		anthropicTools = append(anthropicTools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Schema,
		})
	}

	payload := anthropicRequest{
		Model:         p.model(),
		MaxTokens:     maxTokens,
		System:        system,
		Messages:      messages,
		Tools:         anthropicTools,
		Temperature:   o.temperature,
		TopP:          o.topP,
		TopK:          o.topK,
		StopSequences: o.stopSequences,
	}

	headers := map[string]string{
		"x-api-key":         p.APIKey,
		"anthropic-version": "2023-06-01",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", nil, Usage{}, err
	}

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, p.buildURL(anthropicChatPath), body, headers)
	if err != nil {
		return "", nil, Usage{}, err
	}

	if statusCode >= 400 {
		return "", nil, Usage{}, parseError(Anthropic, statusCode, respBody, nil)
	}

	var resp anthropicResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", nil, Usage{}, err
	}

	// Extract text and tool calls from response
	var text string
	var calls []toolCall
	for _, c := range resp.Content {
		if c.Type == "text" {
			text = c.Text
		} else if c.Type == "tool_use" {
			calls = append(calls, toolCall{
				id:    c.ID,
				name:  c.Name,
				input: c.Input,
			})
		}
	}

	usage := Usage{
		Input:  resp.Usage.InputTokens,
		Output: resp.Usage.OutputTokens,
	}

	return text, calls, usage, nil
}

const anthropicFilesPath = "/v1/files"

type anthropicFileResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
}

// uploadAnthropic uploads a file to Anthropic's Files API (beta).
func uploadAnthropic(ctx context.Context, p Provider, data []byte, name, mimeType string, o *options) (File, error) {
	headers := map[string]string{
		"x-api-key":         p.APIKey,
		"anthropic-version": "2023-06-01",
		"anthropic-beta":    "files-api-2025-04-14",
	}

	respBody, statusCode, err := doMultipartPost(ctx, o.httpClient, p.buildURL(anthropicFilesPath),
		"file", name, data, nil, headers)
	if err != nil {
		return File{}, err
	}

	if statusCode >= 400 {
		return File{}, parseError(Anthropic, statusCode, respBody, nil)
	}

	var resp anthropicFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return File{}, err
	}

	return File{
		ID:       resp.ID,
		MimeType: resp.MimeType,
		Name:     resp.Filename,
	}, nil
}
