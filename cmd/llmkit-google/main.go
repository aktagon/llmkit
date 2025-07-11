package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/google"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: llmkit-google <system_prompt> <user_prompt> [json_schema]")
	}

	systemPrompt := os.Args[1]
	userPrompt := os.Args[2]

	var jsonSchema string
	if len(os.Args) > 3 {
		jsonSchema = os.Args[3]
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	response, err := google.Prompt(systemPrompt, userPrompt, jsonSchema, apiKey)
	if err != nil {
		log.Fatalf("Error calling Google API: %v", err)
	}

	fmt.Print(response)
}