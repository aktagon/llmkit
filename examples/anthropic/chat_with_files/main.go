package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/anthropic/agents"
	"github.com/aktagon/llmkit/anthropic/types"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <filename>")
	}

	filename := os.Args[1]

	agent, err := agents.New(apiKey)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	file, err := anthropic.UploadFile(filename, apiKey)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	options := &agents.ChatOptions{
		Files:     []types.File{*file},
		MaxTokens: 1000,
	}

	response, err := agent.Chat("Summarize this document.", options)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	fmt.Println(response.Text)
}
