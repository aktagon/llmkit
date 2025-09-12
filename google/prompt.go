package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/google/types"
	"github.com/aktagon/llmkit/httpclient"
)

// buildPrompt constructs the combined prompt from system and user prompts
func buildPrompt(systemPrompt, userPrompt string) string {
	if systemPrompt == "" {
		return userPrompt
	}
	return fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)
}

// buildRequest creates the JSON request body for Google's API with optional files
func buildRequest(systemPrompt, userPrompt, jsonSchema string, settings types.RequestSettings, files ...types.File) ([]byte, error) {
	combinedPrompt := buildPrompt(systemPrompt, userPrompt)

	parts := []types.Part{
		{Text: combinedPrompt},
	}

	// Add file parts if provided
	for _, file := range files {
		parts = append(parts, types.Part{
			FileData: &types.FileData{
				MimeType: file.MimeType,
				FileURI:  file.URI,
			},
		})
	}

	request := types.GoogleRequest{
		Contents: []types.Content{
			{
				Parts: parts,
			},
		},
	}

	// Add generation configuration
	requestSettings := &types.RequestSettings{
		MaxTokens:   settings.MaxTokens,
		Temperature: settings.Temperature,
	}

	if jsonSchema != "" && jsonSchema != "null" {
		var schema interface{}
		if err := json.Unmarshal([]byte(jsonSchema), &schema); err != nil {
			return nil, &errors.SchemaError{
				Field:   "json",
				Message: "invalid JSON: " + err.Error(),
			}
		}
		requestSettings.ResponseMimeType = "application/json"
		requestSettings.ResponseSchema = schema
	}

	request.GenerationConfig = requestSettings

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "marshaling request body",
			Err:       err,
		}
	}

	return data, nil
}

// call makes the HTTP request to Google's API
func call(apiKey string, requestBody []byte) (*types.GoogleResponse, error) {
	client := httpclient.NewClient()

	url := fmt.Sprintf("%s?key=%s", types.Endpoint, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

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
			Provider:   "Google",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   types.Endpoint,
		}
	}

	var response types.GoogleResponse
	if err := json.Unmarshal(bodyText, &response); err != nil {
		return nil, &errors.RequestError{
			Operation: "parsing response",
			Err:       err,
		}
	}

	return &response, nil
}

// Prompt sends a prompt request to Google's Generative AI API
func Prompt(systemPrompt, userPrompt, jsonSchema, apiKey string, files ...types.File) (*types.GoogleResponse, error) {
	return PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey, types.RequestSettings{}, files...)
}

// PromptWithSettings sends a prompt request with custom settings
func PromptWithSettings(systemPrompt, userPrompt, jsonSchema, apiKey string, settings types.RequestSettings, files ...types.File) (*types.GoogleResponse, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	requestBody, err := buildRequest(systemPrompt, userPrompt, jsonSchema, settings, files...)
	if err != nil {
		return nil, fmt.Errorf("building request body: %w", err)
	}

	response, err := call(apiKey, requestBody)
	if err != nil {
		return nil, fmt.Errorf("calling Google API: %w", err)
	}

	return response, nil
}
