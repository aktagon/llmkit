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

	// Create a new chat agent
	agent, err := agents.New(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== OpenAI Chat Agent Example ===\n")

	// Example 1: Simple chat
	fmt.Println("1. Simple chat:")
	response, err := agent.Chat("Hello! What's your name?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GPT: %s\n", response.Text)
	fmt.Printf("Tokens used: %d input, %d output\n\n",
		response.Raw.Usage.PromptTokens, response.Raw.Usage.CompletionTokens)

	// Example 2: Chat with system prompt
	fmt.Println("2. Chat with system prompt:")
	response, err = agent.Chat("What's the weather like?", &agents.ChatOptions{
		SystemPrompt: "You are a helpful weather assistant. Always be cheerful and positive.",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GPT: %s\n\n", response.Text)

	// Example 3: Memory usage
	fmt.Println("3. Memory example:")
	agent.Remember("user_name", "Alice")
	agent.Remember("favorite_color", "blue")

	name, exists := agent.Recall("user_name")
	if exists {
		fmt.Printf("Remembered user name: %s\n", name)
	}

	fmt.Printf("All memory: %+v\n", agent.GetMemory())

	color, _ := agent.Recall("favorite_color")
	response, err = agent.Chat(fmt.Sprintf("Hi! My name is %s and I love the color %s. Please remember this!",
		name, color))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GPT: %s\n\n", response.Text)

	// Example 4: Temperature and MaxTokens control
	fmt.Println("4. Temperature and token control:")
	response, err = agent.Chat("Write a creative story about a robot", &agents.ChatOptions{
		Temperature: 0.2,
		MaxTokens:   100,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Creative story (temp=%.1f, max=%d tokens): %s\n\n", 0.2, 100, response.Text)

	// Example 5: Structured output with schema
	fmt.Println("5. Structured output:")
	schema := `{
		"name": "person_info",
		"description": "Extract person information",
		"strict": true,
		"schema": {
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "number"},
				"city": {"type": "string"}
			},
			"required": ["name", "age", "city"],
			"additionalProperties": false
		}
	}`

	response, err = agent.Chat("I'm John, 25 years old, living in New York", &agents.ChatOptions{
		Schema: schema,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Structured response: %s\n\n", response.Text)

	// Example 6: Reset conversation (keep memory)
	fmt.Println("6. Reset conversation:")
	agent.Reset(false) // Keep memory, clear conversation

	response, err = agent.Chat("Do you remember anything about me?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GPT: %s\n", response.Text)

	name, exists = agent.Recall("user_name")
	if exists {
		fmt.Printf("Memory still intact: %s\n", name)
	}

	fmt.Println("\nOpenAI Chat Agent example completed!")
}
