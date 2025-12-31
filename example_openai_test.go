package llmkit_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit"
)

func ExamplePrompt_openai() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}
	req := llmkit.Request{User: "What is 2+2? Answer with just the number."}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExamplePrompt_openaiStructured() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}

	schema := `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"],"additionalProperties":false}`
	req := llmkit.Request{
		User:   "Extract: Bob is 42 years old.",
		Schema: schema,
	}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExamplePrompt_openaiWithImage() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}

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

func ExampleNewAgent_openai() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.OpenAI, APIKey: os.Getenv("OPENAI_API_KEY")}

	agent := llmkit.NewAgent(p)
	agent.SetSystem("You are a helpful assistant. Be very concise.")

	resp, err := agent.Chat(ctx, "My name is Charlie.")
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
