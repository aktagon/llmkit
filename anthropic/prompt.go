package anthropic

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aktagon/llmkit/errors"
)

// validateSchema checks if the provided JSON schema has required top-level attributes
func validateSchema(schemaStr string) (*JsonSchema, error) {
	var schema JsonSchema
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return nil, &errors.SchemaError{
			Field:   "json",
			Message: "invalid JSON: " + err.Error(),
		}
	}

	if schema.Name == "" {
		return nil, &errors.SchemaError{
			Field:   "name",
			Message: "required field missing",
		}
	}
	if schema.Description == "" {
		return nil, &errors.SchemaError{
			Field:   "description",
			Message: "required field missing",
		}
	}
	if !schema.Strict {
		return nil, &errors.SchemaError{
			Field:   "strict",
			Message: "must be true",
		}
	}
	if schema.JsonSchema == nil {
		return nil, &errors.SchemaError{
			Field:   "schema",
			Message: "required field missing",
		}
	}

	return &schema, nil
}

// buildPrompt constructs the user prompt, optionally appending schema instructions
func buildPrompt(userPrompt, jsonSchema string) (string, error) {
	if jsonSchema == "" {
		return userPrompt, nil
	}

	// Validate the schema
	_, err := validateSchema(jsonSchema)
	if err != nil {
		return "", fmt.Errorf("schema validation failed: %w", err)
	}

	// Append schema instructions to user prompt
	return fmt.Sprintf("You must output only the raw JSON without further explanation or formatting. %s\n\nUse the following JSON schema for the output format:\n\n%s", userPrompt, jsonSchema), nil
}

// buildRequest creates the JSON request body for the Anthropic API
func buildRequest(systemPrompt, userPrompt string, files []File) (string, error) {
	var content interface{}

	if len(files) == 0 {
		content = userPrompt
	} else {
		blocks := []Content{{Type: "text", Text: userPrompt}}
		for _, file := range files {
			blocks = append(blocks, Content{
				Type:   "document",
				Source: &FileSource{Type: "file", FileID: file.ID},
			})
		}
		content = blocks
	}

	messages := []map[string]interface{}{
		{"role": "user", "content": content},
	}

	requestBody := map[string]interface{}{
		"model":      Model,
		"max_tokens": MaxTokens,
		"messages":   messages,
	}

	if systemPrompt != "" {
		requestBody["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", &errors.RequestError{
			Operation: "marshaling request body",
			Err:       err,
		}
	}

	return string(jsonData), nil
}

// call makes the HTTP request to the Anthropic API
func call(endpoint, apiKey string, requestBody string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(requestBody))
	if err != nil {
		return "", &errors.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", AnthropicVersion)
	req.Header.Set("anthropic-beta", FilesBetaHeader)
	req.Header.Set("content-type", "application/json")

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
			Provider:   "Anthropic",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   endpoint,
		}
	}

	return string(bodyText), nil
}

// Prompt sends a prompt request to Anthropic API with optional file attachments
func Prompt(systemPrompt, userPrompt, jsonSchema, apiKey string, files ...File) (string, error) {
	if apiKey == "" {
		return "", &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	// Build the complete user prompt with optional schema
	finalUserPrompt, err := buildPrompt(userPrompt, jsonSchema)
	if err != nil {
		return "", fmt.Errorf("building user prompt: %w", err)
	}

	// Build the request body
	requestBody, err := buildRequest(systemPrompt, finalUserPrompt, files)
	if err != nil {
		return "", fmt.Errorf("building request body: %w", err)
	}

	// Make the API call
	response, err := call(Endpoint, apiKey, requestBody)
	if err != nil {
		return "", fmt.Errorf("calling Anthropic API: %w", err)
	}

	return response, nil
}
