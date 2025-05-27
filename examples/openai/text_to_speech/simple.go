// Example of basic OpenAI Text-to-Speech usage
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/openai"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}

	// Simple usage with defaults
	text := "Hello! This is a simple text-to-speech example."

	fmt.Println("Converting text to speech...")
	audioData, err := openai.Text2Speech(text, apiKey, nil)
	if err != nil {
		log.Fatalf("TTS conversion failed: %v", err)
	}

	// Save the audio file
	filename := "simple_example.mp3"
	err = os.WriteFile(filename, audioData, 0644)
	if err != nil {
		log.Fatalf("Failed to save audio: %v", err)
	}

	fmt.Printf("Audio saved as '%s' (%d bytes)\n", filename, len(audioData))
	fmt.Println("You can now play the audio file!")
}
