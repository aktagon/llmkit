package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/grok"
	"github.com/aktagon/llmkit/grok/types"
)

func main() {
	apiKey := os.Getenv("XAI_API_KEY")
	if apiKey == "" {
		log.Fatal("XAI_API_KEY environment variable is required")
	}

	fmt.Println("=== Basic Prompt Example ===\n")

	// Example 1: Simple prompt
	fmt.Println("1. Simple prompt:")
	systemPrompt := "You are a helpful assistant."
	userPrompt := "What is 2 + 2?"

	response, err := grok.Prompt(systemPrompt, userPrompt, "", apiKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Grok: %s\n", response.Choices[0].Message.Content)
		fmt.Printf("Tokens used: %d total (%d prompt + %d completion)\n\n",
			response.Usage.TotalTokens,
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens)
	}

	// Example 2: Structured output with JSON schema
	fmt.Println("2. Structured output:")
	schema := `{
		"name": "math_result",
		"description": "A mathematical calculation result",
		"strict": true,
		"schema": {
			"type": "object",
			"properties": {
				"equation": {"type": "string"},
				"result": {"type": "number"},
				"explanation": {"type": "string"}
			},
			"required": ["equation", "result", "explanation"],
			"additionalProperties": false
		}
	}`

	response, err = grok.Prompt(
		"You are a math tutor. Provide structured answers.",
		"Calculate 15 * 8 and explain.",
		schema,
		apiKey,
	)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Structured JSON: %s\n\n", response.Choices[0].Message.Content)
	}

	// Example 3: Custom settings
	fmt.Println("3. Custom settings (temperature and max tokens):")
	settings := types.RequestSettings{
		MaxTokens:   50,
		Temperature: 0.2,
	}

	response, err = grok.PromptWithSettings(
		"You are a concise assistant.",
		"Write a one-sentence description of machine learning.",
		"",
		apiKey,
		settings,
	)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Grok (temp=%.1f, max=%d): %s\n",
			settings.Temperature,
			settings.MaxTokens,
			response.Choices[0].Message.Content)
	}
}