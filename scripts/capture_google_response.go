package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/google"
)

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("GOOGLE_API_KEY environment variable required")
		fmt.Println("Set it and run: go run scripts/capture_google_response.go")
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