package llmkit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
)

// doPost sends a POST request and returns the response body.
// Returns error only for non-2xx status codes after reading response.
func doPost(ctx context.Context, client *http.Client, url string, body []byte, headers map[string]string) ([]byte, error) {
	data, statusCode, err := doPostRaw(ctx, client, url, body, headers)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, &APIError{
			StatusCode: statusCode,
			Message:    string(data),
			Retryable:  statusCode == 429 || statusCode >= 500,
		}
	}
	return data, nil
}

// doPostRaw sends a POST request and returns status code and body without error handling.
func doPostRaw(ctx context.Context, client *http.Client, url string, body []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return data, resp.StatusCode, nil
}

// doMultipartPost sends a multipart POST request for file uploads.
// Sets Content-Type based on filename extension.
func doMultipartPost(ctx context.Context, client *http.Client, url string,
	fieldName, filename string, data []byte, fields map[string]string, headers map[string]string) ([]byte, int, error) {

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Add extra fields
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, 0, err
		}
	}

	// Add file with proper MIME type from filename
	mimeType := detectMimeType(filename)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	h.Set("Content-Type", mimeType)
	fw, err := w.CreatePart(h)
	if err != nil {
		return nil, 0, err
	}
	if _, err := fw.Write(data); err != nil {
		return nil, 0, err
	}

	if err := w.Close(); err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respData, resp.StatusCode, nil
}

// detectMimeType returns MIME type based on file extension.
func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".csv":
		return "text/csv"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
