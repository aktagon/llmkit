package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/openai"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Example 1: Basic usage with defaults (tts-1 model, alloy voice, mp3 format)
	fmt.Println("Example 1: Basic text-to-speech conversion")
	basicText := "Hello, this is a basic text-to-speech example using OpenAI's TTS API."

	audioData, err := openai.Text2Speech(basicText, apiKey, nil)
	if err != nil {
		log.Fatalf("Basic TTS failed: %v", err)
	}

	// Save to file
	err = os.WriteFile("basic_example.mp3", audioData, 0644)
	if err != nil {
		log.Fatalf("Failed to save basic audio file: %v", err)
	}

	fmt.Printf("✓ Basic audio saved as 'basic_example.mp3' (%d bytes)\n\n", len(audioData))

	// Example 2: Advanced usage with custom options
	fmt.Println("Example 2: Advanced text-to-speech with custom options")
	advancedText := "This is an advanced example using a high-definition model with a different voice and format."

	options := &openai.TTSOptions{
		Model:          openai.ModelTTS1HD,
		Voice:          openai.VoiceNova,
		ResponseFormat: openai.FormatWAV,
		Speed:          1.2,
	}

	audioData, err = openai.Text2Speech(advancedText, apiKey, options)
	if err != nil {
		log.Fatalf("Advanced TTS failed: %v", err)
	}

	// Save to file
	err = os.WriteFile("advanced_example.wav", audioData, 0644)
	if err != nil {
		log.Fatalf("Failed to save advanced audio file: %v", err)
	}

	fmt.Printf("✓ Advanced audio saved as 'advanced_example.wav' (%d bytes)\n\n", len(audioData))

	// Example 3: Different voices demonstration
	fmt.Println("Example 3: Demonstrating different voices")
	voices := []struct {
		voice openai.Voice
		name  string
	}{
		{openai.VoiceAlloy, "alloy"},
		{openai.VoiceEcho, "echo"},
		{openai.VoiceFable, "fable"},
		{openai.VoiceOnyx, "onyx"},
		{openai.VoiceShimmer, "shimmer"},
	}

	text := "This is a demonstration of different OpenAI TTS voices."

	for _, v := range voices {
		options := &openai.TTSOptions{
			Voice: v.voice,
		}

		audioData, err := openai.Text2Speech(text, apiKey, options)
		if err != nil {
			log.Printf("Failed to generate audio for voice %s: %v", v.name, err)
			continue
		}

		filename := fmt.Sprintf("voice_%s.mp3", v.name)
		err = os.WriteFile(filename, audioData, 0644)
		if err != nil {
			log.Printf("Failed to save audio file for voice %s: %v", v.name, err)
			continue
		}

		fmt.Printf("✓ Voice '%s' saved as '%s' (%d bytes)\n", v.name, filename, len(audioData))
	}

	fmt.Println("\n🎉 All examples completed successfully!")
	fmt.Println("Generated audio files:")
	fmt.Println("  - basic_example.mp3 (default settings)")
	fmt.Println("  - advanced_example.wav (HD model, Nova voice, WAV format)")
	fmt.Println("  - voice_*.mp3 (different voice samples)")
}
