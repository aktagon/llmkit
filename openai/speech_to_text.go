package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aktagon/llmkit"
)

// validateSTTInput validates the audio data and options
func validateSTTInput(audioData []byte, filename string, options *STTOptions) error {
	if len(audioData) == 0 {
		return &llmkit.ValidationError{
			Field:   "audioData",
			Message: "audio data is required",
		}
	}

	if strings.TrimSpace(filename) == "" {
		return &llmkit.ValidationError{
			Field:   "filename",
			Message: "filename is required",
		}
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".flac", ".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".ogg", ".wav", ".webm"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}
	if !isValid {
		return &llmkit.ValidationError{
			Field:   "filename",
			Message: "unsupported audio format. Supported: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm",
		}
	}

	if options != nil {
		if options.Temperature < 0 || options.Temperature > 1 {
			return &llmkit.ValidationError{
				Field:   "temperature",
				Message: "temperature must be between 0 and 1",
			}
		}
	}

	return nil
}

// buildSTTRequest creates a multipart form request for transcription
func buildSTTRequest(audioData []byte, filename string, options *STTOptions) (*bytes.Buffer, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add the audio file
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, "", &llmkit.RequestError{
			Operation: "creating form file field",
			Err:       err,
		}
	}

	_, err = fileWriter.Write(audioData)
	if err != nil {
		return nil, "", &llmkit.RequestError{
			Operation: "writing audio data",
			Err:       err,
		}
	}

	// Add model (required)
	model := ModelWhisper1
	if options != nil && options.Model != "" {
		model = options.Model
	}
	
	err = writer.WriteField("model", model)
	if err != nil {
		return nil, "", &llmkit.RequestError{
			Operation: "writing model field",
			Err:       err,
		}
	}

	// Add optional fields
	if options != nil {
		if options.Language != "" {
			err = writer.WriteField("language", options.Language)
			if err != nil {
				return nil, "", &llmkit.RequestError{
					Operation: "writing language field",
					Err:       err,
				}
			}
		}

		if options.Prompt != "" {
			err = writer.WriteField("prompt", options.Prompt)
			if err != nil {
				return nil, "", &llmkit.RequestError{
					Operation: "writing prompt field",
					Err:       err,
				}
			}
		}

		if options.ResponseFormat != "" {
			err = writer.WriteField("response_format", string(options.ResponseFormat))
			if err != nil {
				return nil, "", &llmkit.RequestError{
					Operation: "writing response_format field",
					Err:       err,
				}
			}
		}

		if options.Temperature != 0 {
			err = writer.WriteField("temperature", strconv.FormatFloat(options.Temperature, 'f', -1, 64))
			if err != nil {
				return nil, "", &llmkit.RequestError{
					Operation: "writing temperature field",
					Err:       err,
				}
			}
		}

		if len(options.TimestampGranularities) > 0 {
			for _, granularity := range options.TimestampGranularities {
				err = writer.WriteField("timestamp_granularities[]", string(granularity))
				if err != nil {
					return nil, "", &llmkit.RequestError{
						Operation: "writing timestamp_granularities field",
						Err:       err,
					}
				}
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, "", &llmkit.RequestError{
			Operation: "closing multipart writer",
			Err:       err,
		}
	}

	return &body, writer.FormDataContentType(), nil
}

// callSTT sends the HTTP request to OpenAI transcription API
func callSTT(apiKey string, body *bytes.Buffer, contentType string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", EndpointTranscriptions, body)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "creating STT request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "sending STT request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "reading STT response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &llmkit.APIError{
			Provider:   "OpenAI",
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
			Endpoint:   EndpointTranscriptions,
		}
	}

	return bodyBytes, nil
}

// Speech2Text transcribes audio to text using OpenAI's Whisper API
// Returns the transcribed text and error for simple usage
func Speech2Text(audioData []byte, filename string, apiKey string, options *STTOptions) (string, error) {
	if apiKey == "" {
		return "", &llmkit.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	if err := validateSTTInput(audioData, filename, options); err != nil {
		return "", err
	}

	body, contentType, err := buildSTTRequest(audioData, filename, options)
	if err != nil {
		return "", err
	}

	responseBytes, err := callSTT(apiKey, body, contentType)
	if err != nil {
		return "", err
	}

	// Handle different response formats
	if options != nil && options.ResponseFormat == STTFormatText {
		// For text format, the response is plain text
		return string(responseBytes), nil
	}

	// For JSON formats, parse the response
	var response STTResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return "", &llmkit.RequestError{
			Operation: "parsing STT response",
			Err:       err,
		}
	}

	return response.Text, nil
}

// Speech2TextDetailed transcribes audio and returns detailed response with metadata
// Use this when you need timestamps, segments, or other metadata
func Speech2TextDetailed(audioData []byte, filename string, apiKey string, options *STTOptions) (*STTResponse, error) {
	if apiKey == "" {
		return nil, &llmkit.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	if err := validateSTTInput(audioData, filename, options); err != nil {
		return nil, err
	}

	// Force JSON format for detailed response
	if options == nil {
		options = &STTOptions{ResponseFormat: STTFormatJSON}
	} else if options.ResponseFormat == "" || options.ResponseFormat == STTFormatText {
		options.ResponseFormat = STTFormatJSON
	}

	body, contentType, err := buildSTTRequest(audioData, filename, options)
	if err != nil {
		return nil, err
	}

	responseBytes, err := callSTT(apiKey, body, contentType)
	if err != nil {
		return nil, err
	}

	var response STTResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, &llmkit.RequestError{
			Operation: "parsing detailed STT response",
			Err:       err,
		}
	}

	return &response, nil
}
