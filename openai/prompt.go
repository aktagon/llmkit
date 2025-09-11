package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/httpclient"
	"github.com/aktagon/llmkit/openai/types"
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

	if schema.Description == "" {
		return schema, &errors.SchemaError{
			Field:   "description",
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
func buildMessageContent(text string, files ...types.FileUploadResponse) interface{} {
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

// buildStructuredRequest creates a structured output request with optional file attachments
func buildStructuredRequest(systemPrompt, userPrompt string, schema types.SchemaValidation, settings types.RequestSettings, files ...types.FileUploadResponse) ([]byte, error) {
	request := types.StructuredRequest{
		Model: types.Model,
		Input: []types.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildMessageContent(userPrompt, files...)},
		},
		Text: types.TextFormat{
			Format: types.JsonSchema{
				Type:   "json_schema",
				Name:   schema.Name,
				Schema: schema.Schema,
				Strict: schema.Strict,
			},
		},
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "marshaling structured request",
			Err:       err,
		}
	}

	return data, nil
}

// buildRequest creates a standard chat completion request with optional file attachments
func buildRequest(systemPrompt, userPrompt string, settings types.RequestSettings, files ...types.FileUploadResponse) ([]byte, error) {
	request := types.ChatRequest{
		Model: types.Model,
		Messages: []types.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildMessageContent(userPrompt, files...)},
		},
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

	return data, nil
}

// call sends the HTTP request to OpenAI API
func call(endpoint, apiKey string, requestBody []byte) (*types.Response, error) {
	client := httpclient.NewClient()

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(requestBody))
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
			Provider:   "OpenAI",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   endpoint,
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

// Prompt sends a prompt request to OpenAI API with optional file attachments
func Prompt(systemPrompt, userPrompt, jsonSchema, apiKey string, files ...types.FileUploadResponse) (*types.Response, error) {
	return PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey, types.RequestSettings{}, files...)
}

// PromptWithSettings sends a prompt request with custom settings
func PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey string, settings types.RequestSettings, files ...types.FileUploadResponse) (*types.Response, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	var requestBody []byte
	var err error
	var endpoint string

	if jsonSchema != "" && jsonSchema != "null" {
		// Validate and use structured output
		schema, err := validateSchema(jsonSchema)
		if err != nil {
			return nil, err
		}
		requestBody, err = buildStructuredRequest(systemPrompt, userPrompt, schema, settings, files...)
		endpoint = types.EndpointResponses
	} else {
		// Use standard chat completion
		requestBody, err = buildRequest(systemPrompt, userPrompt, settings, files...)
		endpoint = types.EndpointCompletions
	}

	if err != nil {
		return nil, err
	}

	response, err := call(endpoint, apiKey, requestBody)
	if err != nil {
		return nil, err
	}

	return response, nil
}
