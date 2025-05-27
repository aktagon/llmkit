// Simple example of OpenAI Speech-to-Text usage
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

	// Check for audio file argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run simple.go <audio_file>")
	}

	audioFile := os.Args[1]

	// Read audio file
	fmt.Printf("Transcribing: %s\n", audioFile)
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		log.Fatalf("Failed to read audio file: %v", err)
	}

	// Simple transcription with defaults
	text, err := openai.Speech2Text(audioData, audioFile, apiKey, nil)
	if err != nil {
		log.Fatalf("Transcription failed: %v", err)
	}

	fmt.Printf("Transcribed text: %s\n", text)
}
