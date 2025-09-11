package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable required")
		fmt.Println("Set it and run: go run scripts/capture_openai_response.go")
		os.Exit(1)
	}

	fmt.Println("Calling OpenAI API with prompt 'What is 2+2?'...")
	response, err := openai.Prompt("", "What is 2+2?", "", apiKey)
	if err != nil {
		log.Fatal("API call failed:", err)
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("JSON marshal failed:", err)
	}

	err = os.WriteFile("openai/testdata/simple_response.json", jsonData, 0644)
	if err != nil {
		log.Fatal("Write file failed:", err)
	}

	fmt.Println("✓ OpenAI response saved to openai/testdata/simple_response.json")
	fmt.Printf("Response text: %s\n", response.Choices[0].Message.Content)
}