package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/openai/agents"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}

	// Create agent
	agent, err := agents.New(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== OpenAI Temperature and MaxTokens Test ===\n")

	// Test 1: High temperature for creativity
	fmt.Println("1. Creative response (high temperature):")
	response, err := agent.Chat("Write one sentence about robots", &agents.ChatOptions{
		Temperature: 0.9,
		MaxTokens:   30,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Tokens: %d input, %d output\n\n", response.Raw.Usage.PromptTokens, response.Raw.Usage.CompletionTokens)

	// Test 2: Low temperature for precision
	fmt.Println("2. Precise response (low temperature):")
	response, err = agent.Chat("What is 2+2?", &agents.ChatOptions{
		Temperature: 0.1,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Tokens: %d input, %d output\n\n", response.Raw.Usage.PromptTokens, response.Raw.Usage.CompletionTokens)

	// Test 3: Limited tokens
	fmt.Println("3. Short response (limited tokens):")
	response, err = agent.Chat("Explain artificial intelligence", &agents.ChatOptions{
		MaxTokens: 20,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Tokens: %d input, %d output\n\n", response.Raw.Usage.PromptTokens, response.Raw.Usage.CompletionTokens)

	// Test 4: All options combined
	fmt.Println("4. All options combined:")
	response, err = agent.Chat("Analyze the weather", &agents.ChatOptions{
		SystemPrompt: "You are a meteorologist. Be concise but informative.",
		Temperature:  0.3,
		MaxTokens:    50,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Tokens: %d input, %d output\n", response.Raw.Usage.PromptTokens, response.Raw.Usage.CompletionTokens)

	fmt.Println("\n✅ All OpenAI tests completed successfully!")
}
