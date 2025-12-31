package llmkit_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit"
)

func ExamplePrompt_grok() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Grok, APIKey: os.Getenv("XAI_API_KEY")}
	req := llmkit.Request{User: "What is 2+2? Answer with just the number."}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExamplePrompt_grokStructured() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Grok, APIKey: os.Getenv("XAI_API_KEY")}

	schema := `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name","age"],"additionalProperties":false}`
	req := llmkit.Request{
		User:   "Extract: Frank is 50 years old.",
		Schema: schema,
	}

	resp, err := llmkit.Prompt(ctx, p, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, resp.Text)
	// Output:
}

func ExamplePrompt_grokWithImage() {
	ctx := context.Background()
	p := llmkit.Provider{
		Name:   llmkit.Grok,
		APIKey: os.Getenv("XAI_API_KEY"),
		Model:  "grok-2-vision-latest",
	}

	req := llmkit.Request{
		User: "What color is this square? Answer with one word.",
		Images: []llmkit.Image{{
			URL:      "https://upload.wikimedia.org/wikipedia/commons/thumb/2/29/Solid_green.svg/120px-Solid_green.svg.png",
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

func ExampleNewAgent_grok() {
	ctx := context.Background()
	p := llmkit.Provider{Name: llmkit.Grok, APIKey: os.Getenv("XAI_API_KEY")}

	agent := llmkit.NewAgent(p)
	agent.SetSystem("You are a helpful assistant. Be very concise.")

	resp, err := agent.Chat(ctx, "My name is Grace.")
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
