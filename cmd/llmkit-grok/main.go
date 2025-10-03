package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/grok"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: llmkit-grok <system_prompt> <user_prompt> [json_schema]")
	}

	systemPrompt := os.Args[1]
	userPrompt := os.Args[2]

	var jsonSchema string
	if len(os.Args) > 3 {
		jsonSchema = os.Args[3]
	}

	apiKey := os.Getenv("XAI_API_KEY")
	if apiKey == "" {
		log.Fatal("XAI_API_KEY environment variable is required")
	}

	response, err := grok.Prompt(systemPrompt, userPrompt, jsonSchema, apiKey)
	if err != nil {
		log.Fatalf("Error calling Grok API: %v", err)
	}

	if len(response.Choices) > 0 {
		fmt.Print(response.Choices[0].Message.Content)
	}
}