package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/openai/types"
)

// UploadFile uploads a file to OpenAI and returns file metadata
func UploadFile(filePath, purpose, apiKey string) (*types.FileUploadResponse, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	if purpose == "" {
		purpose = "fine-tune"
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: fmt.Sprintf("opening file %s", filePath),
			Err:       err,
		}
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add purpose field
	purposeField, err := writer.CreateFormField("purpose")
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating purpose field",
			Err:       err,
		}
	}
	purposeField.Write([]byte(purpose))

	// Add file field
	filename := filepath.Base(filePath)
	fileField, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating file field",
			Err:       err,
		}
	}

	_, err = io.Copy(fileField, file)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "copying file content",
			Err:       err,
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "closing multipart writer",
			Err:       err,
		}
	}

	req, err := http.NewRequest("POST", types.EndpointFiles, &body)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating upload request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "sending upload request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "reading upload response",
			Err:       err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &errors.APIError{
			Provider:   "OpenAI",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   types.EndpointFiles,
		}
	}

	var uploadResponse types.FileUploadResponse
	err = json.Unmarshal(bodyText, &uploadResponse)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "parsing upload response",
			Err:       err,
		}
	}

	return &uploadResponse, nil
}