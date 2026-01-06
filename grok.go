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
	Instructions   string               `json:"instructions,omitempty"`
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

	// Add files as input_file
	for _, f := range req.Files {
		input = append(input, grokResponsesInput{
			Type:   "input_file",
			FileID: f.ID,
		})
	}

	// Add user text as input_text
	if req.User != "" {
		input = append(input, grokResponsesInput{
			Type: "input_text",
			Text: req.User,
		})
	}

	payload := grokResponsesRequest{
		Model:        p.model(),
		Instructions: req.System,
		Input:        input,
		Temperature:  o.temperature,
		MaxTokens:    o.maxTokens,
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
