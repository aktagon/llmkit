package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/anthropic"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY environment variable required")
		fmt.Println("Set it and run: go run scripts/capture_anthropic_response.go")
		os.Exit(1)
	}

	fmt.Println("Calling Anthropic API with prompt 'What is 2+2?'...")
	response, err := anthropic.Prompt("", "What is 2+2?", "", apiKey)
	if err != nil {
		log.Fatal("API call failed:", err)
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("JSON marshal failed:", err)
	}

	err = os.WriteFile("anthropic/testdata/simple_response.json", jsonData, 0644)
	if err != nil {
		log.Fatal("Write file failed:", err)
	}

	fmt.Println("✓ Anthropic response saved to anthropic/testdata/simple_response.json")
	fmt.Printf("Response text: %s\n", response.Content[0].Text)
}