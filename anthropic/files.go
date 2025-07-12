package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/aktagon/llmkit/errors"
)

// UploadFile uploads a file to Anthropic and returns file metadata
func UploadFile(filePath, apiKey string) (*File, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "opening file",
			Err:       err,
		}
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Detect MIME type from file extension
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Create form file with proper MIME type
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath)))
	header.Set("Content-Type", mimeType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating form file",
			Err:       err,
		}
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "copying file data",
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

	req, err := http.NewRequest("POST", FilesEndpoint, &body)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating upload request",
			Err:       err,
		}
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", AnthropicVersion)
	req.Header.Set("anthropic-beta", FilesBetaHeader)
	req.Header.Set("content-type", writer.FormDataContentType())

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
			Provider:   "Anthropic",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   FilesEndpoint,
		}
	}

	var uploadedFile File
	err = json.Unmarshal(bodyText, &uploadedFile)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "parsing upload response",
			Err:       err,
		}
	}

	return &uploadedFile, nil
}
