package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/google"
	"github.com/aktagon/llmkit/openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/tools/capture_response.go <provider>")
		fmt.Println("Providers: anthropic, google, openai")
		os.Exit(1)
	}

	provider := strings.ToLower(os.Args[1])

	switch provider {
	case "anthropic":
		captureAnthropicResponse()
	case "google":
		captureGoogleResponse()
	case "openai":
		captureOpenAIResponse()
	default:
		fmt.Printf("Unknown provider: %s\n", provider)
		fmt.Println("Supported providers: anthropic, google, openai")
		os.Exit(1)
	}
}

func captureAnthropicResponse() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY environment variable required")
		fmt.Println("Set it and run: go run cmd/tools/capture_response.go anthropic")
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

func captureGoogleResponse() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("GOOGLE_API_KEY environment variable required")
		fmt.Println("Set it and run: go run cmd/tools/capture_response.go google")
		os.Exit(1)
	}

	fmt.Println("Calling Google API with prompt 'What is 2+2?'...")
	response, err := google.Prompt("", "What is 2+2?", "", apiKey)
	if err != nil {
		log.Fatal("API call failed:", err)
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("JSON marshal failed:", err)
	}

	err = os.WriteFile("google/testdata/simple_response.json", jsonData, 0644)
	if err != nil {
		log.Fatal("Write file failed:", err)
	}

	fmt.Println("✓ Google response saved to google/testdata/simple_response.json")
	fmt.Printf("Response text: %s\n", response.Candidates[0].Content.Parts[0].Text)
}

func captureOpenAIResponse() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable required")
		fmt.Println("Set it and run: go run cmd/tools/capture_response.go openai")
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
