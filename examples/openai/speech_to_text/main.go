package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aktagon/llmkit/openai"
	"github.com/aktagon/llmkit/openai/types"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Check if audio file is provided as command line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <audio_file_path>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  go run main.go sample.mp3")
		fmt.Println("  go run main.go recording.wav")
		fmt.Println("")
		fmt.Println("Supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm")
		os.Exit(1)
	}

	audioFilePath := os.Args[1]

	// Read the audio file
	fmt.Printf("Reading audio file: %s\n", audioFilePath)
	audioData, err := os.ReadFile(audioFilePath)
	if err != nil {
		log.Fatalf("Failed to read audio file: %v", err)
	}

	fmt.Printf("✓ Audio file loaded (%d bytes)\n\n", len(audioData))

	// Example 1: Basic transcription with defaults
	fmt.Println("Example 1: Basic transcription (Whisper-1 model)")
	fmt.Println("=" + strings.Repeat("=", 50))

	basicText, err := openai.Speech2Text(audioData, audioFilePath, apiKey, nil)
	if err != nil {
		log.Fatalf("Basic transcription failed: %v", err)
	}

	fmt.Printf("Transcribed text: %s\n\n", basicText)

	// Example 2: Advanced transcription with options
	fmt.Println("Example 2: Advanced transcription with custom options")
	fmt.Println("=" + strings.Repeat("=", 50))

	options := &types.STTOptions{
		Model:       types.ModelWhisper1,
		Language:    "en", // English
		Prompt:      "This audio contains technical terminology and proper names.",
		Temperature: 0.2, // More focused/deterministic
	}

	advancedText, err := openai.Speech2Text(audioData, audioFilePath, apiKey, options)
	if err != nil {
		log.Fatalf("Advanced transcription failed: %v", err)
	}

	fmt.Printf("Transcribed text: %s\n\n", advancedText)

	// Example 3: Detailed transcription with metadata
	fmt.Println("Example 3: Detailed transcription with timestamps")
	fmt.Println("=" + strings.Repeat("=", 50))

	detailedOptions := &types.STTOptions{
		Model:          types.ModelWhisper1,
		ResponseFormat: types.STTFormatVerboseJSON,
		TimestampGranularities: []types.TimestampGranularity{
			types.GranularitySegment,
		},
	}

	detailedResponse, err := openai.Speech2TextDetailed(audioData, audioFilePath, apiKey, detailedOptions)
	if err != nil {
		log.Fatalf("Detailed transcription failed: %v", err)
	}

	fmt.Printf("Full text: %s\n", detailedResponse.Text)
	if detailedResponse.Language != "" {
		fmt.Printf("Detected language: %s\n", detailedResponse.Language)
	}
	if detailedResponse.Duration > 0 {
		fmt.Printf("Duration: %.2f seconds\n", detailedResponse.Duration)
	}

	if len(detailedResponse.Segments) > 0 {
		fmt.Println("\nSegments with timestamps:")
		for i, segment := range detailedResponse.Segments {
			fmt.Printf("  [%d] %.2fs - %.2fs: %s\n",
				i+1, segment.Start, segment.End, strings.TrimSpace(segment.Text))
		}
	}

	// Example 4: Different response formats
	fmt.Println("\n" + "Example 4: Different response formats")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Plain text format
	textOptions := &types.STTOptions{
		ResponseFormat: types.STTFormatText,
	}

	plainText, err := openai.Speech2Text(audioData, audioFilePath, apiKey, textOptions)
	if err != nil {
		log.Printf("Text format transcription failed: %v", err)
	} else {
		fmt.Printf("Plain text format: %s\n", plainText)
	}

	// SRT subtitle format
	srtOptions := &types.STTOptions{
		ResponseFormat: types.STTFormatSRT,
	}

	srtText, err := openai.Speech2Text(audioData, audioFilePath, apiKey, srtOptions)
	if err != nil {
		log.Printf("SRT format transcription failed: %v", err)
	} else {
		fmt.Println("\nSRT subtitle format:")
		fmt.Println(srtText)

		// Save SRT to file
		srtFilename := "transcription.srt"
		err = os.WriteFile(srtFilename, []byte(srtText), 0644)
		if err != nil {
			log.Printf("Failed to save SRT file: %v", err)
		} else {
			fmt.Printf("✓ SRT subtitles saved as '%s'\n", srtFilename)
		}
	}

	fmt.Println("\nAll examples completed successfully!")
	fmt.Println("\nAPI Features demonstrated:")
	fmt.Println("  ✓ Basic transcription with defaults")
	fmt.Println("  ✓ Custom options (language, prompt, temperature)")
	fmt.Println("  ✓ Detailed response with metadata and timestamps")
	fmt.Println("  ✓ Multiple response formats (JSON, text, SRT)")
	fmt.Println("  ✓ Error handling and validation")
}
