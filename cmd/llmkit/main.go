package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit"
)

func main() {
	var provider string
	var systemPrompt string
	var userPrompt string
	var jsonSchema string

	flag.StringVar(&provider, "provider", "", "LLM provider (openai or anthropic)")
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
		log.Fatal("Usage: llmkit -provider <openai|anthropic> -system <system_prompt> -user <user_prompt> [-schema <json_schema>]\n" +
			"   or: llmkit -provider <openai|anthropic> <system_prompt> <user_prompt> [json_schema]")
	}

	if systemPrompt == "" || userPrompt == "" {
		log.Fatal("Both system prompt and user prompt are required")
	}

	var apiKey string
	var providerType llmkit.Provider

	switch provider {
	case "openai":
		providerType = llmkit.ProviderOpenAI
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable is required")
		}
	case "anthropic":
		providerType = llmkit.ProviderAnthropic
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			log.Fatal("ANTHROPIC_API_KEY environment variable is required")
		}
	default:
		log.Fatalf("Unsupported provider: %s. Use 'openai' or 'anthropic'", provider)
	}

	response, err := llmkit.Prompt(llmkit.PromptOptions{
		Provider:     providerType,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		JSONSchema:   jsonSchema,
		APIKey:       apiKey,
	})
	if err != nil {
		log.Fatalf("Error calling %s API: %v", provider, err)
	}

	fmt.Print(response)
}
