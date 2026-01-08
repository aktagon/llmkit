package llmkit

import (
	"context"
	"encoding/json"
)

const (
	// grokResponsesPath is xAI's preferred endpoint, supports files and text.
	// See: https://docs.x.ai/docs/api-reference#chat-responses
	grokResponsesPath = "/v1/responses"
	// grokFilesPath is for uploading files before referencing in prompts.
	grokFilesPath = "/v1/files"
)

// Responses API request types
type grokResponsesRequest struct {
	Model          string               `json:"model"`
	Input          []grokResponsesInput `json:"input"`
	ResponseFormat *grokResponseFormat  `json:"response_format,omitempty"`
	Temperature    *float64             `json:"temperature,omitempty"`
	MaxTokens      *int                 `json:"max_output_tokens,omitempty"`
}

type grokResponseFormat struct {
	Type       string         `json:"type"`
	JSONSchema grokJSONSchema `json:"json_schema"`
}

type grokJSONSchema struct {
	Name   string `json:"name"`
	Schema any    `json:"schema"`
	Strict bool   `json:"strict"`
}

type grokResponsesInput struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []grokContentPart
}

type grokContentPart struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	FileID string `json:"file_id,omitempty"`
}

type grokResponsesResponse struct {
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func promptGrok(ctx context.Context, p Provider, req Request, o *options) (Response, error) {
	var input []grokResponsesInput

	// Add system message if present
	if req.System != "" {
		input = append(input, grokResponsesInput{
			Role:    "system",
			Content: req.System,
		})
	}

	// Build user content (text and/or files)
	if req.User != "" || len(req.Files) > 0 {
		if len(req.Files) == 0 {
			// Simple text-only message
			input = append(input, grokResponsesInput{
				Role:    "user",
				Content: req.User,
			})
		} else {
			// Mixed content with files
			var parts []grokContentPart
			for _, f := range req.Files {
				parts = append(parts, grokContentPart{
					Type:   "file",
					FileID: f.ID,
				})
			}
			if req.User != "" {
				parts = append(parts, grokContentPart{
					Type: "text",
					Text: req.User,
				})
			}
			input = append(input, grokResponsesInput{
				Role:    "user",
				Content: parts,
			})
		}
	}

	payload := grokResponsesRequest{
		Model:       p.model(),
		Input:       input,
		Temperature: o.temperature,
		MaxTokens:   o.maxTokens,
	}

	if req.Schema != "" {
		var schema any
		if err := json.Unmarshal([]byte(req.Schema), &schema); err != nil {
			return Response{}, err
		}
		payload.ResponseFormat = &grokResponseFormat{
			Type: "json_schema",
			JSONSchema: grokJSONSchema{
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

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, p.buildURL(grokResponsesPath), body, headers)
	if err != nil {
		return Response{}, err
	}

	if statusCode >= 400 {
		return Response{}, parseError(Grok, statusCode, respBody, nil)
	}

	var resp grokResponsesResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return Response{}, err
	}

	// Extract text from Responses API format
	text := ""
	if len(resp.Output) > 0 && len(resp.Output[0].Content) > 0 {
		text = resp.Output[0].Content[0].Text
	}

	return Response{
		Text: text,
		Tokens: Usage{
			Input:  resp.Usage.InputTokens,
			Output: resp.Usage.OutputTokens,
		},
	}, nil
}

type grokFileResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

// uploadGrok uploads a file to Grok's Files API.
func uploadGrok(ctx context.Context, p Provider, data []byte, name string, o *options) (File, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + p.APIKey,
	}
	fields := map[string]string{
		"purpose": "assistants",
	}

	respBody, statusCode, err := doMultipartPost(ctx, o.httpClient, p.buildURL(grokFilesPath),
		"file", name, data, fields, headers)
	if err != nil {
		return File{}, err
	}

	if statusCode >= 400 {
		return File{}, parseError(Grok, statusCode, respBody, nil)
	}

	var resp grokFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return File{}, err
	}

	return File{
		ID:       resp.ID,
		MimeType: detectMimeType(name),
		Name:     resp.Filename,
	}, nil
}
