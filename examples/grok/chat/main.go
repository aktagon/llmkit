package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aktagon/llmkit/grok"
	"github.com/aktagon/llmkit/grok/types"
)

// Conversation holds the chat history
type Conversation struct {
	messages []types.Message
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(role, content string) {
	c.messages = append(c.messages, types.Message{
		Role:    role,
		Content: content,
	})
}

// GetMessages returns the conversation messages
func (c *Conversation) GetMessages() []types.Message {
	return c.messages
}

// sendChatRequest sends a chat request with conversation history
func sendChatRequest(conv *Conversation, apiKey string) (string, error) {
	settings := types.RequestSettings{}
	response, err := grok.ChatWithMessages(conv.GetMessages(), apiKey, settings)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from Grok")
	}

	return response.Choices[0].Message.Content, nil
}

func main() {
	apiKey := os.Getenv("XAI_API_KEY")
	if apiKey == "" {
		log.Fatal("XAI_API_KEY environment variable is required")
	}

	conv := &Conversation{}

	// Add system message
	systemPrompt := "You are Grok, a helpful and witty AI assistant. Keep responses concise but informative."
	conv.AddMessage("system", systemPrompt)

	fmt.Println("=== Grok Chat Example ===")
	fmt.Println("Have a conversation with Grok. Type 'exit' to quit.\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if userInput == "" {
			continue
		}

		// Add user message
		conv.AddMessage("user", userInput)

		// Get response
		response, err := sendChatRequest(conv, apiKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			// Remove the failed user message
			conv.messages = conv.messages[:len(conv.messages)-1]
			continue
		}

		// Add assistant response to conversation
		conv.AddMessage("assistant", response)

		fmt.Printf("\nGrok: %s\n\n", response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}