package llmkit_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aktagon/llmkit"
)

// Basic usage of the llmkit package.
func Example() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}
	req := llmkit.Request{User: "What is 2+2?"}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Using a system prompt to set context or persona.
func ExamplePrompt_withSystem() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Anthropic, APIKey: os.Getenv("ANTHROPIC_API_KEY")}
	req := llmkit.Request{
		System: "You are a helpful assistant. Be concise.",
		User:   "Explain Go interfaces in one sentence.",
	}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Using a JSON schema to get structured output.
func ExamplePrompt_structuredOutput() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}

	schema := `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"],"additionalProperties":false}`
	req := llmkit.Request{
		User:   "Extract: John is 30 years old.",
		Schema: schema,
	}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Handling different error types from the API.
func ExamplePrompt_errorHandling() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: "invalid-key"}
	req := llmkit.Request{User: "Hello"}

	_, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		var apiErr *llmkit.APIError
		var valErr *llmkit.ValidationError
		switch {
		case errors.As(err, &apiErr):
			fmt.Fprintf(os.Stderr, "API error: %s (retryable: %v)\n", apiErr.Message, apiErr.Retryable)
		case errors.As(err, &valErr):
			fmt.Fprintf(os.Stderr, "Validation error: %s - %s\n", valErr.Field, valErr.Message)
		default:
			fmt.Fprintln(os.Stderr, "Unknown error:", err)
		}
	}
	// Output:
}

// Using a hook to log requests before they are sent.
func ExampleWithBeforeRequest() {
	hook := llmkit.WithBeforeRequest(func(ctx context.Context, req *llmkit.Request) error {
		fmt.Fprintf(os.Stderr, "Sending request: %s\n", req.User)
		return nil
	})

	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}
	req := llmkit.Request{User: "Hello"}

	resp, err := llmkit.Prompt(ctx, p, req, hook)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Using a custom HTTP client with timeout.
func ExampleWithHTTPClient() {
	client := &http.Client{Timeout: 30 * time.Second}

	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}
	req := llmkit.Request{User: "Hello"}

	resp, err := llmkit.Prompt(ctx, p, req, llmkit.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Using temperature for creative output.
func ExampleWithTemperature() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}
	req := llmkit.Request{User: "Write a haiku about Go programming"}

	resp, err := llmkit.Prompt(ctx, p, req,
		llmkit.WithTemperature(0.9), // higher = more creative
		llmkit.WithMaxTokens(100),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Using low temperature for deterministic output.
func ExampleWithTemperature_deterministic() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}
	req := llmkit.Request{User: "What is the capital of France?"}

	resp, err := llmkit.Prompt(ctx, p, req,
		llmkit.WithTemperature(0.0), // deterministic
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

// Using a custom base URL for OpenAI-compatible APIs (LiteLLM, vLLM, Ollama, etc.)
func ExampleProvider_customBaseURL() {
	ctx := context.Background()
	p := llmkit.Provider{
		Name:    llmkit.OpenAI,
		APIKey:  "your-api-key",
		BaseURL: "http://localhost:4000", // LiteLLM proxy
	}
	req := llmkit.Request{User: "Hello"}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		// Expected to fail without a running proxy
		fmt.Println("Custom base URL configured")
		return
	}
	fmt.Println(resp.Text)
	// Output: Custom base URL configured
}
