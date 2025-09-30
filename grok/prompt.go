package grok

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/grok/types"
	"github.com/aktagon/llmkit/httpclient"
	openaitypes "github.com/aktagon/llmkit/openai/types"
)

// validateSchema validates that the JSON schema has required top-level attributes
func validateSchema(schemaJSON string) (types.SchemaValidation, error) {
	var schema types.SchemaValidation

	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return schema, &errors.SchemaError{
			Field:   "json",
			Message: "invalid JSON: " + err.Error(),
		}
	}

	if schema.Name == "" {
		return schema, &errors.SchemaError{
			Field:   "name",
			Message: "required field missing",
		}
	}

	if !schema.Strict {
		return schema, &errors.SchemaError{
			Field:   "strict",
			Message: "must be true",
		}
	}

	if schema.Schema == nil {
		return schema, &errors.SchemaError{
			Field:   "schema",
			Message: "required field missing",
		}
	}

	return schema, nil
}

// buildMessageContent creates message content with optional files
func buildMessageContent(text string, files ...openaitypes.FileUploadResponse) interface{} {
	if len(files) == 0 {
		return text
	}

	parts := []types.ContentPart{{Type: "text", Text: text}}
	for _, file := range files {
		parts = append(parts, types.ContentPart{
			Type: "file",
			File: &types.FileReference{FileID: file.ID},
		})
	}
	return parts
}

// buildRequest creates a chat completion request with optional file attachments and schema
func buildRequest(systemPrompt, userPrompt string, settings types.RequestSettings, schema *types.SchemaValidation, files ...openaitypes.FileUploadResponse) ([]byte, error) {
	request := types.ChatRequest{
		Model: types.Model,
		Messages: []types.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildMessageContent(userPrompt, files...)},
		},
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
	}

	if schema != nil {
		request.ResponseFormat = &types.ResponseFormat{
			Type: "json_schema",
			JsonSchema: types.JsonSchema{
				Type:   "json_schema",
				Name:   schema.Name,
				Schema: schema.Schema,
				Strict: schema.Strict,
			},
		}
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "marshaling chat request",
			Err:       err,
		}
	}

	return data, nil
}

// call sends the HTTP request to Grok API
func call(apiKey string, requestBody []byte) (*types.Response, error) {
	client := httpclient.NewClient()

	req, err := http.NewRequest("POST", types.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "sending request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "reading response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &errors.APIError{
			Provider:   "Grok",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   types.Endpoint,
		}
	}

	var response types.Response
	if err := json.Unmarshal(bodyText, &response); err != nil {
		return nil, &errors.RequestError{
			Operation: "parsing response",
			Err:       err,
		}
	}

	return &response, nil
}

// Prompt sends a prompt request to Grok API with optional file attachments
func Prompt(systemPrompt, userPrompt, jsonSchema, apiKey string, files ...openaitypes.FileUploadResponse) (*types.Response, error) {
	return PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey, types.RequestSettings{}, files...)
}

// PromptWithSettings sends a prompt request with custom settings
func PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey string, settings types.RequestSettings, files ...openaitypes.FileUploadResponse) (*types.Response, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	var schema *types.SchemaValidation
	if jsonSchema != "" && jsonSchema != "null" {
		validatedSchema, err := validateSchema(jsonSchema)
		if err != nil {
			return nil, err
		}
		schema = &validatedSchema
	}

	requestBody, err := buildRequest(systemPrompt, userPrompt, settings, schema, files...)
	if err != nil {
		return nil, err
	}

	response, err := call(apiKey, requestBody)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ChatWithMessages sends a chat request with custom message history
func ChatWithMessages(messages []types.Message, apiKey string, settings types.RequestSettings) (*types.Response, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	request := types.ChatRequest{
		Model:       types.Model,
		Messages:    messages,
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "marshaling chat request",
			Err:       err,
		}
	}

	response, err := call(apiKey, data)
	if err != nil {
		return nil, err
	}

	return response, nil
}