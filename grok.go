package llmkit

import (
	"context"
	"encoding/json"
)

const (
	grokChatPath  = "/v1/chat/completions"
	grokFilesPath = "/v1/files"
)

// Grok uses OpenAI-compatible API format
type grokRequest struct {
	Model            string              `json:"model"`
	Messages         []grokMessage       `json:"messages"`
	ResponseFormat   *grokResponseFormat `json:"response_format,omitempty"`
	Temperature      *float64            `json:"temperature,omitempty"`
	TopP             *float64            `json:"top_p,omitempty"`
	MaxTokens        *int                `json:"max_tokens,omitempty"`
	Stop             []string            `json:"stop,omitempty"`
	Seed             *int64              `json:"seed,omitempty"`
	FrequencyPenalty *float64            `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64            `json:"presence_penalty,omitempty"`
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

type grokMessage struct {
	Role    string         `json:"role"`
	Content []grokContent  `json:"content"`
}

type grokContent struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *grokImageURL `json:"image_url,omitempty"`
	File     *grokFile     `json:"file,omitempty"`
}

type grokFile struct {
	FileID string `json:"file_id"`
}

type grokImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type grokResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func promptGrok(ctx context.Context, p Provider, req Request, o *options) (Response, error) {
	var msgs []grokMessage
	if req.System != "" {
		msgs = append(msgs, grokMessage{
			Role:    "system",
			Content: []grokContent{{Type: "text", Text: req.System}},
		})
	}
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			msgs = append(msgs, grokMessage{
				Role:    m.Role,
				Content: []grokContent{{Type: "text", Text: m.Content}},
			})
		}
	} else {
		msgs = append(msgs, grokMessage{Role: "user", Content: buildGrokContent(req)})
	}

	payload := grokRequest{
		Model:            p.model(),
		Messages:         msgs,
		Temperature:      o.temperature,
		TopP:             o.topP,
		MaxTokens:        o.maxTokens,
		Stop:             o.stopSequences,
		Seed:             o.seed,
		FrequencyPenalty: o.frequencyPenalty,
		PresencePenalty:  o.presencePenalty,
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

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, p.buildURL(grokChatPath), body, headers)
	if err != nil {
		return Response{}, err
	}

	if statusCode >= 400 {
		return Response{}, parseError(Grok, statusCode, respBody, nil)
	}

	var resp grokResponse
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

// buildGrokContent creates content array from request.
func buildGrokContent(req Request) []grokContent {
	var content []grokContent

	// Add files first
	for _, f := range req.Files {
		content = append(content, grokContent{
			Type: "file",
			File: &grokFile{FileID: f.ID},
		})
	}

	// Add images
	for _, img := range req.Images {
		c := grokContent{
			Type: "image_url",
			ImageURL: &grokImageURL{
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
		content = append(content, grokContent{Type: "text", Text: req.User})
	}

	return content
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
