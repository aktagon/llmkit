package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aktagon/llmkit/errors"
)

// buildPrompt constructs the combined prompt from system and user prompts
func buildPrompt(systemPrompt, userPrompt string) string {
	if systemPrompt == "" {
		return userPrompt
	}
	return fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)
}

// buildRequest creates the JSON request body for Google's API
func buildRequest(systemPrompt, userPrompt, jsonSchema string) ([]byte, error) {
	combinedPrompt := buildPrompt(systemPrompt, userPrompt)
	
	request := GoogleRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: combinedPrompt},
				},
			},
		},
	}

	// Add structured output configuration if schema provided
	if jsonSchema != "" && jsonSchema != "null" {
		var schema interface{}
		if err := json.Unmarshal([]byte(jsonSchema), &schema); err != nil {
			return nil, &errors.SchemaError{
				Field:   "json",
				Message: "invalid JSON: " + err.Error(),
			}
		}

		request.GenerationConfig = &GenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   schema,
		}
	}

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
func call(apiKey string, requestBody []byte) (string, error) {
	client := &http.Client{}
	
	url := fmt.Sprintf("%s?key=%s", Endpoint, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return "", &errors.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", &errors.RequestError{
			Operation: "sending request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &errors.RequestError{
			Operation: "reading response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return "", &errors.APIError{
			Provider:   "Google",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   Endpoint,
		}
	}

	return string(bodyText), nil
}

// Prompt sends a prompt request to Google's Generative AI API
func Prompt(systemPrompt, userPrompt, jsonSchema, apiKey string) (string, error) {
	if apiKey == "" {
		return "", &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	requestBody, err := buildRequest(systemPrompt, userPrompt, jsonSchema)
	if err != nil {
		return "", fmt.Errorf("building request body: %w", err)
	}

	response, err := call(apiKey, requestBody)
	if err != nil {
		return "", fmt.Errorf("calling Google API: %w", err)
	}

	return response, nil
}