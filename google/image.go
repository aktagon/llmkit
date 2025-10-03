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

// ImageInput represents an input image for editing
type ImageInput struct {
	Data     []byte // Raw image data (JPEG, PNG, etc.)
	MimeType string // MIME type (e.g., "image/png", "image/jpeg")
}

// GenerateImage generates an image from a text prompt using Gemini's image generation model
// For text-to-image: pass only the prompt
// For image editing: pass prompt and optional input image
// Returns the image data as bytes (PNG format)
func GenerateImage(prompt, apiKey string, inputImage *ImageInput) ([]byte, error) {
	return GenerateImageWithSettings(prompt, apiKey, types.RequestSettings{
		Model: types.ImageModel,
	}, inputImage)
}

// GenerateImageWithSettings generates an image with custom settings
// Returns the image data as bytes (PNG format)
func GenerateImageWithSettings(prompt, apiKey string, settings types.RequestSettings, inputImage *ImageInput) ([]byte, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	if prompt == "" {
		return nil, &errors.ValidationError{
			Field:   "prompt",
			Message: "Prompt is required",
		}
	}

	// Use image model if not specified
	if settings.Model == "" {
		settings.Model = types.ImageModel
	}

	requestBody, err := buildImageRequest(prompt, settings, inputImage)
	if err != nil {
		return nil, fmt.Errorf("building request body: %w", err)
	}

	response, err := callImageAPI(apiKey, requestBody, settings)
	if err != nil {
		return nil, fmt.Errorf("calling Google API: %w", err)
	}

	// Extract image from response
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, &errors.APIError{
			Provider: "Google",
			Message:  "No image generated in response",
		}
	}

	// Find the first part with inline data (image)
	for _, part := range response.Candidates[0].Content.Parts {
		if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			// Go's JSON decoder automatically base64-decodes into []byte
			return part.InlineData.Data, nil
		}
	}

	return nil, &errors.APIError{
		Provider: "Google",
		Message:  "No image data found in response",
	}
}

// buildImageRequest constructs the request body for image generation
func buildImageRequest(prompt string, settings types.RequestSettings, inputImage *ImageInput) ([]byte, error) {
	parts := []types.Part{
		{Text: prompt},
	}

	// Add input image if provided (for image editing)
	if inputImage != nil {
		parts = append(parts, types.Part{
			InlineData: &types.InlineData{
				MimeType: inputImage.MimeType,
				Data:     inputImage.Data,
			},
		})
	}

	request := types.GoogleRequest{
		Contents: []types.Content{
			{
				Parts: parts,
			},
		},
		GenerationConfig: &settings,
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

// callImageAPI makes the HTTP request to Google's image generation API
func callImageAPI(apiKey string, requestBody []byte, settings types.RequestSettings) (*types.GoogleResponse, error) {
	client := httpclient.NewClient()

	model := settings.Model
	if model == "" {
		model = types.ImageModel
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", types.BaseEndpoint, model, apiKey)
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
			Endpoint:   url,
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
