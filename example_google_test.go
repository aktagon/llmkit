package llmkit_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit"
)

func getGoogleAPIKey() string {
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	return getGoogleAPIKey()
}

func ExamplePrompt_google() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Google, APIKey: getGoogleAPIKey()}
	req := llmkit.Request{User: "What is 2+2? Answer with just the number."}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExamplePrompt_googleStructured() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Google, APIKey: getGoogleAPIKey()}

	schema := `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"]}`
	req := llmkit.Request{
		User:   "Extract: Diana is 35 years old.",
		Schema: schema,
	}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExamplePrompt_googleWithImage() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Google, APIKey: getGoogleAPIKey()}

	req := llmkit.Request{
		User: "What color is this image? Answer with one word.",
		Images: []llmkit.Image{{
			URL:      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			MimeType: "image/png",
		}},
	}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExampleNewAgent_google() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Google, APIKey: getGoogleAPIKey()}

	agent := llmkit.NewAgent(p)
	agent.SetSystem("You are a helpful assistant. Be very concise.")

	resp, err := agent.Chat(ctx, "My name is Eve.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, "Turn 1:", resp.Text)

	resp, err = agent.Chat(ctx, "What is my name?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, "Turn 2:", resp.Text)
	// Output:
}
