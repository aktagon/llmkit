package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit"
)

func main() {
	var provider string
	var model string
	var systemPrompt string
	var userPrompt string
	var jsonSchema string

	flag.StringVar(&provider, "provider", "", "LLM provider (anthropic, openai, google, grok)")
	flag.StringVar(&model, "model", "", "Model name (optional, uses provider default)")
	flag.StringVar(&systemPrompt, "system", "", "System prompt")
	flag.StringVar(&userPrompt, "user", "", "User prompt")
	flag.StringVar(&jsonSchema, "schema", "", "JSON schema for structured output (optional)")
	flag.Parse()

	// Handle positional arguments for backwards compatibility
	args := flag.Args()
	if len(args) >= 2 && systemPrompt == "" && userPrompt == "" {
		systemPrompt = args[0]
		userPrompt = args[1]
		if len(args) > 2 {
			jsonSchema = args[2]
		}
	}

	if provider == "" {
		fmt.Fprintln(os.Stderr, "Usage: llmkit -provider <anthropic|openai|google|grok> -system <system_prompt> -user <user_prompt> [-schema <json_schema>]")
		fmt.Fprintln(os.Stderr, "   or: llmkit -provider <provider> <system_prompt> <user_prompt> [json_schema]")
		os.Exit(1)
	}

	if systemPrompt == "" || userPrompt == "" {
		log.Fatal("Both system prompt and user prompt are required")
	}

	apiKey := getAPIKey(provider)

	p := llmkit.Provider{
		Name:   provider,
		APIKey: apiKey,
		Model:  model,
	}

	req := llmkit.Request{
		System: systemPrompt,
		User:   userPrompt,
		Schema: jsonSchema,
	}

	resp, err := llmkit.Prompt(context.Background(), p, req)
	if err != nil {
		log.Fatalf("Error calling %s API: %v", provider, err)
	}

	fmt.Print(resp.Text)
}

func getAPIKey(provider string) string {
	var envVar string
	switch provider {
	case llmkit.Anthropic:
		envVar = "ANTHROPIC_API_KEY"
	case llmkit.OpenAI:
		envVar = "OPENAI_API_KEY"
	case llmkit.Google:
		envVar = "GOOGLE_API_KEY"
	case llmkit.Grok:
		envVar = "GROK_API_KEY"
	default:
		log.Fatalf("Unsupported provider: %s", provider)
	}

	key := os.Getenv(envVar)
	if key == "" {
		log.Fatalf("%s environment variable is required", envVar)
	}
	return key
}
