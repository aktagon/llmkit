package llmkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const googleChatPathFmt = "/v1beta/models/%s:generateContent"

type googleRequest struct {
	Contents         []googleContent       `json:"contents"`
	Tools            []googleTool          `json:"tools,omitempty"`
	SystemInstruct   *googleContent        `json:"systemInstruction,omitempty"`
	GenerationConfig *googleGenerationConf `json:"generationConfig,omitempty"`
}

type googleTool struct {
	FunctionDeclarations []googleFunctionDecl `json:"functionDeclarations"`
}

type googleFunctionDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text             string                  `json:"text,omitempty"`
	InlineData       *googleInlineData       `json:"inline_data,omitempty"`
	FileData         *googleFileData         `json:"file_data,omitempty"`
	FunctionCall     *googleFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *googleFunctionResponse `json:"functionResponse,omitempty"`
}

type googleFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type googleFunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type googleFileData struct {
	FileURI  string `json:"file_uri"`
	MimeType string `json:"mime_type"`
}

type googleInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type googleGenerationConf struct {
	ResponseMimeType string               `json:"responseMimeType,omitempty"`
	ResponseSchema   any                  `json:"responseSchema,omitempty"`
	Temperature      *float64             `json:"temperature,omitempty"`
	TopP             *float64             `json:"topP,omitempty"`
	TopK             *int                 `json:"topK,omitempty"`
	MaxOutputTokens  *int                 `json:"maxOutputTokens,omitempty"`
	StopSequences    []string             `json:"stopSequences,omitempty"`
	ThinkingConfig   *googleThinkingConf  `json:"thinkingConfig,omitempty"`
}

type googleThinkingConf struct {
	ThinkingBudget *int   `json:"thinkingBudget,omitempty"` // Gemini 2.5
	ThinkingMode   string `json:"thinkingMode,omitempty"`   // Gemini 3: "low", "high"
}

type googleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text         string              `json:"text,omitempty"`
				FunctionCall *googleFunctionCall `json:"functionCall,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

func promptGoogle(ctx context.Context, p Provider, req Request, o *options) (Response, error) {
	// Build contents array
	var contents []googleContent
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			role := m.Role
			if role == "assistant" {
				role = "model" // Google uses "model" instead of "assistant"
			}
			contents = append(contents, googleContent{
				Role:  role,
				Parts: []googlePart{{Text: m.Content}},
			})
		}
	} else {
		contents = []googleContent{{Role: "user", Parts: buildGoogleParts(req)}}
	}

	payload := googleRequest{
		Contents: contents,
	}

	if req.System != "" {
		payload.SystemInstruct = &googleContent{
			Parts: []googlePart{{Text: req.System}},
		}
	}

	// Build generation config
	genConfig := &googleGenerationConf{
		Temperature:     o.temperature,
		TopP:            o.topP,
		TopK:            o.topK,
		MaxOutputTokens: o.maxTokens,
		StopSequences:   o.stopSequences,
	}

	// Add thinking config if specified
	if o.thinkingBudget != nil || o.reasoningEffort != "" {
		genConfig.ThinkingConfig = &googleThinkingConf{}
		if o.thinkingBudget != nil {
			genConfig.ThinkingConfig.ThinkingBudget = o.thinkingBudget
		}
		if o.reasoningEffort != "" {
			genConfig.ThinkingConfig.ThinkingMode = o.reasoningEffort
		}
	}

	if req.Schema != "" {
		var schema map[string]any
		if err := json.Unmarshal([]byte(req.Schema), &schema); err != nil {
			return Response{}, err
		}
		// Google doesn't support additionalProperties
		delete(schema, "additionalProperties")
		genConfig.ResponseMimeType = "application/json"
		genConfig.ResponseSchema = schema
	}

	payload.GenerationConfig = genConfig

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	path := fmt.Sprintf(googleChatPathFmt, p.model())
	url := p.buildURL(path) + "?key=" + p.APIKey

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, url, body, nil)
	if err != nil {
		return Response{}, err
	}

	if statusCode >= 400 {
		return Response{}, parseError(Google, statusCode, respBody, nil)
	}

	var resp googleResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return Response{}, err
	}

	text := ""
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		text = resp.Candidates[0].Content.Parts[0].Text
	}

	return Response{
		Text: text,
		Tokens: Usage{
			Input:  resp.UsageMetadata.PromptTokenCount,
			Output: resp.UsageMetadata.CandidatesTokenCount,
		},
	}, nil
}

// buildGoogleParts creates parts array from request.
func buildGoogleParts(req Request) []googlePart {
	var parts []googlePart

	// Add files first
	for _, f := range req.Files {
		parts = append(parts, googlePart{
			FileData: &googleFileData{
				FileURI:  f.URI,
				MimeType: f.MimeType,
			},
		})
	}

	// Add images (Google only supports base64, not URLs)
	for _, img := range req.Images {
		data := img.URL
		if strings.HasPrefix(data, "data:") {
			if idx := strings.Index(data, ","); idx != -1 {
				data = data[idx+1:]
			}
		}
		parts = append(parts, googlePart{
			InlineData: &googleInlineData{
				MimeType: img.MimeType,
				Data:     data,
			},
		})
	}

	// Add text
	if req.User != "" {
		parts = append(parts, googlePart{Text: req.User})
	}

	return parts
}

// sendGoogleWithTools sends a request with tools and returns tool calls.
func sendGoogleWithTools(ctx context.Context, p Provider, msgs []message, system string, tools []Tool, o *options) (string, []toolCall, Usage, error) {
	// Build contents
	var contents []googleContent
	for _, m := range msgs {
		role := m.role
		if role == "assistant" {
			role = "model"
		}

		if m.toolResult != nil {
			// Tool result - Google uses functionResponse in parts
			contents = append(contents, googleContent{
				Role: "user",
				Parts: []googlePart{{
					FunctionResponse: &googleFunctionResponse{
						Name:     m.toolResult.toolUseID, // Google uses function name as ID
						Response: map[string]any{"result": m.toolResult.content},
					},
				}},
			})
		} else if len(m.toolCalls) > 0 {
			// Model message with function calls
			var parts []googlePart
			for _, tc := range m.toolCalls {
				parts = append(parts, googlePart{
					FunctionCall: &googleFunctionCall{
						Name: tc.name,
						Args: tc.input,
					},
				})
			}
			contents = append(contents, googleContent{Role: "model", Parts: parts})
		} else {
			// Regular text message
			contents = append(contents, googleContent{
				Role:  role,
				Parts: []googlePart{{Text: m.content}},
			})
		}
	}

	// Build tools
	var decls []googleFunctionDecl
	for _, t := range tools {
		decls = append(decls, googleFunctionDecl{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Schema,
		})
	}

	payload := googleRequest{
		Contents: contents,
		Tools:    []googleTool{{FunctionDeclarations: decls}},
	}

	if system != "" {
		payload.SystemInstruct = &googleContent{
			Parts: []googlePart{{Text: system}},
		}
	}

	// Build generation config
	genConfig := &googleGenerationConf{
		Temperature:     o.temperature,
		TopP:            o.topP,
		TopK:            o.topK,
		MaxOutputTokens: o.maxTokens,
		StopSequences:   o.stopSequences,
	}
	payload.GenerationConfig = genConfig

	body, err := json.Marshal(payload)
	if err != nil {
		return "", nil, Usage{}, err
	}

	path := fmt.Sprintf(googleChatPathFmt, p.model())
	url := p.buildURL(path) + "?key=" + p.APIKey

	respBody, statusCode, err := doPostRaw(ctx, o.httpClient, url, body, nil)
	if err != nil {
		return "", nil, Usage{}, err
	}

	if statusCode >= 400 {
		return "", nil, Usage{}, parseError(Google, statusCode, respBody, nil)
	}

	var resp googleResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", nil, Usage{}, err
	}

	// Extract text and function calls
	var text string
	var calls []toolCall
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				text = part.Text
			}
			if part.FunctionCall != nil {
				calls = append(calls, toolCall{
					id:    part.FunctionCall.Name, // Google uses name as ID
					name:  part.FunctionCall.Name,
					input: part.FunctionCall.Args,
				})
			}
		}
	}

	usage := Usage{
		Input:  resp.UsageMetadata.PromptTokenCount,
		Output: resp.UsageMetadata.CandidatesTokenCount,
	}

	return text, calls, usage, nil
}

const googleUploadPath = "/upload/v1beta/files"

type googleFileResponse struct {
	File struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		MimeType    string `json:"mimeType"`
		URI         string `json:"uri"`
	} `json:"file"`
}

// uploadGoogle uploads a file to Google's Files API.
func uploadGoogle(ctx context.Context, p Provider, data []byte, name, mimeType string, o *options) (File, error) {
	url := p.buildURL(googleUploadPath) + "?key=" + p.APIKey
	headers := map[string]string{
		"X-Goog-Upload-Protocol": "multipart",
	}

	// Google requires JSON metadata as a separate field
	metadata := fmt.Sprintf(`{"file":{"display_name":"%s"}}`, name)
	fields := map[string]string{
		"metadata": metadata,
	}

	respBody, statusCode, err := doMultipartPost(ctx, o.httpClient, url,
		"file", name, data, fields, headers)
	if err != nil {
		return File{}, err
	}

	if statusCode >= 400 {
		return File{}, parseError(Google, statusCode, respBody, nil)
	}

	var resp googleFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return File{}, err
	}

	return File{
		ID:       resp.File.Name,
		URI:      resp.File.URI,
		MimeType: resp.File.MimeType,
		Name:     resp.File.DisplayName,
	}, nil
}
