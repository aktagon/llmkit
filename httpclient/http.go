package httpclient

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// config holds logging configuration
type config struct {
	HTTP         bool   `yaml:"http"`
	Level        string `yaml:"level"`
	OTELEndpoint string `yaml:"otel_endpoint"`
}

var globalConfig *config

// init loads configuration once at startup
func init() {
	globalConfig = loadConfig()
}

// loggingTransport wraps http.RoundTripper with structured logging
type loggingTransport struct {
	base http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !globalConfig.HTTP {
		return t.base.RoundTrip(req)
	}

	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	duration := time.Since(start)

	attrs := []slog.Attr{
		slog.String("provider", providerFromURL(req.URL.String())),
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(req.Context(), slog.LevelError, "HTTP request failed", attrs...)
		return resp, err
	}

	attrs = append(attrs, slog.Int("status", resp.StatusCode))
	
	level := slog.LevelInfo
	if resp.StatusCode >= 400 {
		level = slog.LevelWarn
	}

	slog.LogAttrs(req.Context(), level, "HTTP request", attrs...)
	return resp, nil
}

// NewClient returns an HTTP client with optional logging
func NewClient() *http.Client {
	if !globalConfig.HTTP {
		return &http.Client{}
	}

	return &http.Client{
		Transport: &loggingTransport{
			base: http.DefaultTransport,
		},
	}
}

// loadConfig loads from file then env vars
func loadConfig() *config {
	c := &config{
		HTTP:  false,
		Level: "info",
	}

	// Try settings file
	if data, err := os.ReadFile("llmkit.yaml"); err == nil {
		var wrapper struct {
			Logging config `yaml:"logging"`
		}
		if yaml.Unmarshal(data, &wrapper) == nil {
			c = &wrapper.Logging
		}
	}

	// Environment overrides
	if v := os.Getenv("LLMKIT_LOG_HTTP"); v != "" {
		c.HTTP, _ = strconv.ParseBool(v)
	}
	if v := os.Getenv("LLMKIT_LOG_LEVEL"); v != "" {
		c.Level = v
	}
	if v := os.Getenv("LLMKIT_OTEL_ENDPOINT"); v != "" {
		c.OTELEndpoint = v
	}

	return c
}

// providerFromURL extracts provider name from API URL
func providerFromURL(url string) string {
	switch {
	case strings.Contains(url, "anthropic.com"):
		return "anthropic"
	case strings.Contains(url, "openai.com"):
		return "openai"
	case strings.Contains(url, "googleapis.com"):
		return "google"
	default:
		return "unknown"
	}
}