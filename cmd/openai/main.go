package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/openai"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: llmkit-openai <system_prompt> <user_prompt> [json_schema]")
	}

	systemPrompt := os.Args[1]
	userPrompt := os.Args[2]

	var jsonSchema string
	if len(os.Args) > 3 {
		jsonSchema = os.Args[3]
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	response, err := openai.Chat(systemPrompt, userPrompt, jsonSchema, apiKey)
	if err != nil {
		log.Fatalf("Error calling OpenAI API: %v", err)
	}

	fmt.Print(response)
}
