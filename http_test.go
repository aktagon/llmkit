package llmkit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoPost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Authorization = %s, want Bearer test-key", auth)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	client := server.Client()
	headers := map[string]string{"Authorization": "Bearer test-key"}

	body, err := doPost(context.Background(), client, server.URL, []byte(`{}`), headers)
	if err != nil {
		t.Fatalf("doPost() error = %v", err)
	}
	if string(body) != `{"result":"ok"}` {
		t.Errorf("body = %s, want {\"result\":\"ok\"}", string(body))
	}
}

func TestDoPost_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited","type":"rate_limit_error"}}`))
	}))
	defer server.Close()

	client := server.Client()
	body, statusCode, err := doPostRaw(context.Background(), client, server.URL, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("doPostRaw() network error = %v", err)
	}
	if statusCode != http.StatusTooManyRequests {
		t.Errorf("statusCode = %d, want %d", statusCode, http.StatusTooManyRequests)
	}
	if string(body) != `{"error":{"message":"rate limited","type":"rate_limit_error"}}` {
		t.Errorf("body = %s", string(body))
	}
}

func TestDoPost_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := server.Client()
	_, err := doPost(ctx, client, server.URL, []byte(`{}`), nil)
	if err == nil {
		t.Error("doPost() expected error for canceled context")
	}
}
