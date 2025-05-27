package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/aktagon/llmkit"
)

// validateTTSInput validates the input text and options
func validateTTSInput(input string, options *TTSOptions) error {
	if strings.TrimSpace(input) == "" {
		return &llmkit.ValidationError{
			Field:   "input",
			Message: "text input is required",
		}
	}

	if len(input) > 4096 {
		return &llmkit.ValidationError{
			Field:   "input",
			Message: "input text exceeds 4096 character limit",
		}
	}

	if options != nil {
		if options.Speed != 0 && (options.Speed < 0.25 || options.Speed > 4.0) {
			return &llmkit.ValidationError{
				Field:   "speed",
				Message: "speed must be between 0.25 and 4.0",
			}
		}
	}

	return nil
}

// buildTTSRequest creates a text-to-speech request with defaults
func buildTTSRequest(input string, options *TTSOptions) ([]byte, error) {
	request := TTSRequest{
		Model: ModelTTS1,
		Input: input,
		Voice: VoiceAlloy,
	}

	// Apply options if provided
	if options != nil {
		if options.Model != "" {
			request.Model = options.Model
		}
		if options.Voice != "" {
			request.Voice = options.Voice
		}
		if options.ResponseFormat != "" {
			request.ResponseFormat = options.ResponseFormat
		}
		if options.Speed != 0 {
			request.Speed = options.Speed
		}
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "marshaling TTS request",
			Err:       err,
		}
	}

	return data, nil
}

// callTTS sends the HTTP request to OpenAI TTS API and returns binary audio data
func callTTS(apiKey string, requestBody []byte) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", EndpointSpeech, bytes.NewReader(requestBody))
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "creating TTS request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "sending TTS request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "reading TTS response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &llmkit.APIError{
			Provider:   "OpenAI",
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
			Endpoint:   EndpointSpeech,
		}
	}

	return bodyBytes, nil
}

// Text2Speech converts text to speech using OpenAI's TTS API
// Returns the binary audio data and error
func Text2Speech(input string, apiKey string, options *TTSOptions) ([]byte, error) {
	if apiKey == "" {
		return nil, &llmkit.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	if err := validateTTSInput(input, options); err != nil {
		return nil, err
	}

	requestBody, err := buildTTSRequest(input, options)
	if err != nil {
		return nil, err
	}

	audioData, err := callTTS(apiKey, requestBody)
	if err != nil {
		return nil, err
	}

	return audioData, nil
}
