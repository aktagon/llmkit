package llmkit

import (
	"context"
	"encoding/json"
)

const (
	openaiChatPath  = "/v1/chat/completions"
	openaiFilesPath = "/v1/files"
)

type openaiRequest struct {
	Model            string          `json:"model"`
	Messages         []openaiMessage `json:"messages"`
	Tools            []openaiTool    `json:"tools,omitempty"`
	ResponseFormat   *responseFormat `json:"response_format,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	Stop             []string        `json:"stop,omitempty"`
	Seed             *int64          `json:"seed,omitempty"`
	FrequencyPenalty *float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64        `json:"presence_penalty,omitempty"`
	ReasoningEffort  string          `json:"reasoning_effort,omitempty"`
}

type openaiTool struct {
	Type     string         `json:"type"`
	Function openaiFunction `json:"function"`
}

type openaiFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string `json:"name"`
	Schema any    `json:"schema"`
	Strict bool   `json:"strict"`
}

type openaiMessage struct {
	Role       string              `json:"role"`
	Content    any                 `json:"content,omitempty"`     // []openaiContent or string
	ToolCalls  []openaiToolCall    `json:"tool_calls,omitempty"`  // for assistant
	ToolCallID string              `json:"tool_call_id,omitempty"` // for tool role
}

type openaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

type openaiContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openaiImageURL `json:"image_url,omitempty"`
	File     *openaiFile     `json:"file,omitempty"`
}

type openaiFile struct {
	FileID string `json:"file_id"`
}

type openaiImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content   string           `json:"content"`
			ToolCalls []openaiToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func promptOpenAI(ctx context.Context, p Provider, req Request, o *options) (Response, error) {
	var msgs []openaiMessage
	if req.System != "" {
		msgs = append(msgs, openaiMessage{
			Role:    "system",
			Content: []openaiContent{{Type: "text", Text: req.System}},
		})
	}
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			msgs = append(msgs, openaiMessage{
				Role:    m.Role,
				Content: []openaiContent{{Type: "text", Text: m.Content}},
			})
		}
	} else {
		msgs = append(msgs, openaiMessage{Role: "user", Content: buildOpenAIContent(req)})
	}

	payload := openaiRequest{
		Model:            p.model(),
		Messages:         msgs,
		Temperature:      o.temperature,
		TopP:             o.topP,
		MaxTokens:        o.maxTokens,
		Stop:             o.stopSequences,
		Seed:             o.seed,
		FrequencyPenalty: o.frequencyPenalty,
		PresencePenalty:  o.presencePenalty,
		ReasoningEffort:  o.reasoningEffort,
	}

	if req.Schema != "" {
		var schema any
		if err := json.Unmarshal([]byte(req.Schema), &schema); err != nil {
			return Response{}, err
		}
		payload.ResponseFormat = &responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchema{
				Name:   "response",
				Schema: schema,
				Strict: true,
			},
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + p.APIKey,
	}

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, p.buildURL(openaiChatPath), body, headers)
	if err != nil {
		return Response{}, err
	}

	if statusCode >= 400 {
		return Response{}, parseError(OpenAI, statusCode, respBody, nil)
	}

	var resp openaiResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return Response{}, err
	}

	text := ""
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Message.Content
	}

	return Response{
		Text: text,
		Tokens: Usage{
			Input:  resp.Usage.PromptTokens,
			Output: resp.Usage.CompletionTokens,
		},
	}, nil
}

// buildOpenAIContent creates content array from request.
func buildOpenAIContent(req Request) []openaiContent {
	var content []openaiContent

	// Add files first
	for _, f := range req.Files {
		content = append(content, openaiContent{
			Type: "file",
			File: &openaiFile{FileID: f.ID},
		})
	}

	// Add images
	for _, img := range req.Images {
		c := openaiContent{
			Type: "image_url",
			ImageURL: &openaiImageURL{
				URL:    img.URL,
				Detail: img.Detail,
			},
		}
		if c.ImageURL.Detail == "" {
			c.ImageURL.Detail = "auto"
		}
		content = append(content, c)
	}

	// Add text
	if req.User != "" {
		content = append(content, openaiContent{Type: "text", Text: req.User})
	}

	return content
}

// sendOpenAIWithTools sends a request with tools and returns tool calls.
func sendOpenAIWithTools(ctx context.Context, p Provider, msgs []message, system string, tools []Tool, o *options) (string, []toolCall, Usage, error) {
	// Build messages
	var messages []openaiMessage
	if system != "" {
		messages = append(messages, openaiMessage{
			Role:    "system",
			Content: system,
		})
	}

	for _, m := range msgs {
		if m.toolResult != nil {
			// Tool result message
			messages = append(messages, openaiMessage{
				Role:       "tool",
				Content:    m.toolResult.content,
				ToolCallID: m.toolResult.toolUseID,
			})
		} else if len(m.toolCalls) > 0 {
			// Assistant message with tool calls
			var oaiCalls []openaiToolCall
			for _, tc := range m.toolCalls {
				argsJSON, _ := json.Marshal(tc.input)
				oaiCalls = append(oaiCalls, openaiToolCall{
					ID:   tc.id,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tc.name,
						Arguments: string(argsJSON),
					},
				})
			}
			messages = append(messages, openaiMessage{
				Role:      "assistant",
				ToolCalls: oaiCalls,
			})
		} else {
			// Regular text message
			messages = append(messages, openaiMessage{
				Role:    m.role,
				Content: m.content,
			})
		}
	}

	// Build tools
	var oaiTools []openaiTool
	for _, t := range tools {
		oaiTools = append(oaiTools, openaiTool{
			Type: "function",
			Function: openaiFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Schema,
			},
		})
	}

	payload := openaiRequest{
		Model:            p.model(),
		Messages:         messages,
		Tools:            oaiTools,
		Temperature:      o.temperature,
		TopP:             o.topP,
		MaxTokens:        o.maxTokens,
		Stop:             o.stopSequences,
		Seed:             o.seed,
		FrequencyPenalty: o.frequencyPenalty,
		PresencePenalty:  o.presencePenalty,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", nil, Usage{}, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + p.APIKey,
	}

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, p.buildURL(openaiChatPath), body, headers)
	if err != nil {
		return "", nil, Usage{}, err
	}

	if statusCode >= 400 {
		return "", nil, Usage{}, parseError(OpenAI, statusCode, respBody, nil)
	}

	var resp openaiResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", nil, Usage{}, err
	}

	// Extract text and tool calls
	var text string
	var calls []toolCall
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Message.Content
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			var input map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &input)
			calls = append(calls, toolCall{
				id:    tc.ID,
				name:  tc.Function.Name,
				input: input,
			})
		}
	}

	usage := Usage{
		Input:  resp.Usage.PromptTokens,
		Output: resp.Usage.CompletionTokens,
	}

	return text, calls, usage, nil
}

type openaiFileResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

// uploadOpenAI uploads a file to OpenAI's Files API.
func uploadOpenAI(ctx context.Context, p Provider, data []byte, name string, o *options) (File, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + p.APIKey,
	}
	fields := map[string]string{
		"purpose": "assistants",
	}

	respBody, statusCode, err := doMultipartPost(ctx, o.httpClient, p.buildURL(openaiFilesPath),
		"file", name, data, fields, headers)
	if err != nil {
		return File{}, err
	}

	if statusCode >= 400 {
		return File{}, parseError(OpenAI, statusCode, respBody, nil)
	}

	var resp openaiFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return File{}, err
	}

	return File{
		ID:       resp.ID,
		MimeType: detectMimeType(name),
		Name:     resp.Filename,
	}, nil
}
