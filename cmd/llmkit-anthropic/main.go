package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/anthropic"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: llmkit-anthropic <system_prompt> <user_prompt> [json_schema]")
	}

	systemPrompt := os.Args[1]
	userPrompt := os.Args[2]

	var jsonSchema string
	if len(os.Args) > 3 {
		jsonSchema = os.Args[3]
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	response, err := anthropic.Prompt(systemPrompt, userPrompt, jsonSchema, apiKey)
	if err != nil {
		log.Fatalf("Error calling Anthropic API: %v", err)
	}

	if len(response.Content) > 0 {
		fmt.Print(response.Content[0].Text)
	}
}
