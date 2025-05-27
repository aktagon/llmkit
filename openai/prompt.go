package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aktagon/llmkit"
)

// validateSchema validates that the JSON schema has required top-level attributes
func validateSchema(schemaJSON string) (SchemaValidation, error) {
	var schema SchemaValidation

	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return schema, &llmkit.SchemaError{
			Field:   "json",
			Message: "invalid JSON: " + err.Error(),
		}
	}

	if schema.Name == "" {
		return schema, &llmkit.SchemaError{
			Field:   "name",
			Message: "required field missing",
		}
	}

	if schema.Description == "" {
		return schema, &llmkit.SchemaError{
			Field:   "description",
			Message: "required field missing",
		}
	}

	if !schema.Strict {
		return schema, &llmkit.SchemaError{
			Field:   "strict",
			Message: "must be true",
		}
	}

	if schema.Schema == nil {
		return schema, &llmkit.SchemaError{
			Field:   "schema",
			Message: "required field missing",
		}
	}

	return schema, nil
}

// buildStructuredRequest creates a structured output request
func buildStructuredRequest(systemPrompt, userPrompt string, schema SchemaValidation) ([]byte, error) {
	request := StructuredRequest{
		Model: Model,
		Input: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Text: TextFormat{
			Format: JsonSchema{
				Type:   "json_schema",
				Name:   schema.Name,
				Schema: schema.Schema,
				Strict: schema.Strict,
			},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "marshaling structured request",
			Err:       err,
		}
	}

	return data, nil
}

// buildRequest creates a standard chat completion request
func buildRequest(systemPrompt, userPrompt string) ([]byte, error) {
	request := ChatRequest{
		Model: Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "marshaling chat request",
			Err:       err,
		}
	}

	return data, nil
}

// call sends the HTTP request to OpenAI API
func call(endpoint, apiKey string, requestBody []byte) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", &llmkit.RequestError{
			Operation: "creating request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", &llmkit.RequestError{
			Operation: "sending request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &llmkit.RequestError{
			Operation: "reading response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return "", &llmkit.APIError{
			Provider:   "OpenAI",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   endpoint,
		}
	}

	return string(bodyText), nil
}

// Prompt sends a prompt request to OpenAI API
func Prompt(systemPrompt, userPrompt, jsonSchema, apiKey string) (string, error) {
	if apiKey == "" {
		return "", &llmkit.ValidationError{
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
			return "", err
		}
		requestBody, err = buildStructuredRequest(systemPrompt, userPrompt, schema)
		endpoint = EndpointResponses
	} else {
		// Use standard chat completion
		requestBody, err = buildRequest(systemPrompt, userPrompt)
		endpoint = EndpointCompletions
	}

	if err != nil {
		return "", err
	}

	response, err := call(endpoint, apiKey, requestBody)
	if err != nil {
		return "", err
	}

	return response, nil
}
