package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/anthropic/agents"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set ANTHROPIC_API_KEY environment variable")
	}

	fmt.Println("=== Memory Features Demo ===")

	// Example 1: Memory as Context
	fmt.Println("1. Memory as Context:")
	agent1, err := agents.New(apiKey, agents.WithMemoryContext())
	if err != nil {
		log.Fatal(err)
	}

	// Add some memory manually
	agent1.Remember("user_name", "Alice")
	agent1.Remember("favorite_color", "blue")
	agent1.Remember("job", "software engineer")

	response, err := agent1.Chat("What do you know about me?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Claude with memory context: %s\n\n", response.Text)

	// Example 2: Memory Tools (Claude controls memory)
	fmt.Println("2. Memory Tools:")
	agent2, err := agents.New(apiKey, agents.WithMemoryTools())
	if err != nil {
		log.Fatal(err)
	}

	response, err = agent2.Chat("Hi! I'm Bob, I love pizza and I work as a teacher. Please remember these facts about me.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Claude with memory tools: %s\n", response.Text)

	// Check what Claude remembered
	memory := agent2.GetMemory()
	fmt.Printf("Memory after Claude's response: %+v\n\n", memory)

	// Example 3: Memory with Persistence
	fmt.Println("3. Memory with Persistence:")
	memoryFile := "./test_memory.json"
	defer os.Remove(memoryFile) // Clean up

	agent3, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence(memoryFile),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add memory and it will auto-save
	agent3.Remember("persistent_fact", "This will survive restart")
	agent3.Remember("session_id", "12345")

	fmt.Printf("Memory saved to %s: %+v\n", memoryFile, agent3.GetMemory())

	// Simulate restart by creating new agent
	agent3Restart, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence(memoryFile),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Memory after restart: %+v\n", agent3Restart.GetMemory())

	response, err = agent3Restart.Chat("What persistent facts do you know?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Claude after restart: %s\n\n", response.Text)

	// Example 4: Full Memory Suite
	fmt.Println("4. Full Memory Suite (Context + Tools + Persistence):")
	fullAgent, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryTools(),
		agents.WithMemoryPersistence("./full_memory.json"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("./full_memory.json")

	response, err = fullAgent.Chat("I'm Charlie, I'm 30 years old, and I love hiking. Also, remember that my birthday is March 15th.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Full memory agent: %s\n", response.Text)

	// Test memory recall
	response, err = fullAgent.Chat("What's my name and when is my birthday?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Memory recall test: %s\n", response.Text)

	fmt.Println("\nMemory features demo completed!")
}
