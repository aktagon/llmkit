package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aktagon/llmkit/errors"
	"github.com/aktagon/llmkit/google/types"
)

// UploadFile uploads a file to Google and returns file metadata
func UploadFile(filePath, apiKey string) (*types.File, error) {
	if apiKey == "" {
		return nil, &errors.ValidationError{
			Field:   "apiKey",
			Message: "API key is required",
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: fmt.Sprintf("opening file %s", filePath),
			Err:       err,
		}
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, &errors.RequestError{
			Operation: fmt.Sprintf("getting file stats for %s", filePath),
			Err:       err,
		}
	}
	fileSize := stat.Size()

	// Detect MIME type from file extension
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	displayName := filepath.Base(filePath)

	// Step 1: Initial resumable request to get upload URL
	uploadURL, err := initiateUpload(apiKey, fileSize, mimeType, displayName)
	if err != nil {
		return nil, err
	}

	// Step 2: Upload the actual file bytes
	uploadedFile, err := uploadFileBytes(uploadURL, file, fileSize)
	if err != nil {
		return nil, err
	}

	return uploadedFile, nil
}

// initiateUpload starts the resumable upload and returns the upload URL
func initiateUpload(apiKey string, fileSize int64, mimeType, displayName string) (string, error) {
	uploadRequest := types.FileUploadRequest{
		File: types.FileUploadInfo{
			DisplayName: displayName,
		},
	}

	requestBody, err := json.Marshal(uploadRequest)
	if err != nil {
		return "", &errors.RequestError{
			Operation: "marshaling upload request",
			Err:       err,
		}
	}

	req, err := http.NewRequest("POST", types.FilesEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", &errors.RequestError{
			Operation: "creating upload request",
			Err:       err,
		}
	}

	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("X-Goog-Upload-Protocol", "resumable")
	req.Header.Set("X-Goog-Upload-Command", "start")
	req.Header.Set("X-Goog-Upload-Header-Content-Length", strconv.FormatInt(fileSize, 10))
	req.Header.Set("X-Goog-Upload-Header-Content-Type", mimeType)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &errors.RequestError{
			Operation: "sending upload initiation request",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return "", &errors.APIError{
			Provider:   "Google",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   types.FilesEndpoint,
		}
	}

	// Extract upload URL from response headers
	uploadURL := resp.Header.Get("X-Goog-Upload-URL")
	if uploadURL == "" {
		return "", &errors.RequestError{
			Operation: "extracting upload URL",
			Err:       fmt.Errorf("upload URL not found in response headers"),
		}
	}

	return uploadURL, nil
}

// uploadFileBytes uploads the actual file content and returns the file metadata
func uploadFileBytes(uploadURL string, file *os.File, fileSize int64) (*types.File, error) {
	req, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "creating file upload request",
			Err:       err,
		}
	}

	req.Header.Set("Content-Length", strconv.FormatInt(fileSize, 10))
	req.Header.Set("X-Goog-Upload-Offset", "0")
	req.Header.Set("X-Goog-Upload-Command", "upload, finalize")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &errors.RequestError{
			Operation: "uploading file bytes",
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
			Provider:   "Google",
			StatusCode: resp.StatusCode,
			Message:    string(bodyText),
			Endpoint:   uploadURL,
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

	return &uploadResponse.File, nil
}
