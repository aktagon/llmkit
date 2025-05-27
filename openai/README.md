# OpenAI API Package

A Go package for OpenAI API integration with support for:

- Chat completions (standard and structured output with JSON schemas)
- Text-to-Speech (TTS) conversion with multiple voices and formats
- Speech-to-Text (STT) transcription with Whisper models

## Setup

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## CLI Usage

```bash
go run cmd/openai/main.go <system_prompt> <user_prompt> [json_schema]
```

## Programmatic Usage

### Chat Completion

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/aktagon/llmkit/openai"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")

    response, err := openai.Prompt(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple prompt
        apiKey,
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

### Text-to-Speech

```go
package main

import (
    "log"
    "os"

    "github.com/aktagon/llmkit/openai"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")

    // Basic usage with defaults
    audioData, err := openai.Text2Speech("Hello world!", apiKey, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Save audio to file
    err = os.WriteFile("output.mp3", audioData, 0644)
    if err != nil {
        log.Fatal(err)
    }

    // Advanced usage with custom options
    options := &openai.TTSOptions{
        Model:          openai.ModelTTS1HD,
        Voice:          openai.VoiceNova,
        ResponseFormat: openai.FormatWAV,
        Speed:          1.2,
    }

    audioData, err = openai.Text2Speech("Custom voice example", apiKey, options)
    if err != nil {
        log.Fatal(err)
    }

    err = os.WriteFile("custom.wav", audioData, 0644)
    if err != nil {
        log.Fatal(err)
    }
### Speech-to-Text

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/aktagon/llmkit/openai"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    
    // Read audio file
    audioData, err := os.ReadFile("recording.mp3")
    if err != nil {
        log.Fatal(err)
    }
    
    // Basic transcription with defaults
    text, err := openai.Speech2Text(audioData, "recording.mp3", apiKey, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Transcribed: %s\n", text)
    
    // Advanced transcription with options
    options := &openai.STTOptions{
        Model:       openai.ModelWhisper1,
        Language:    "en",
        Prompt:      "This audio contains technical terms.",
        Temperature: 0.2,
    }
    
    text, err = openai.Speech2Text(audioData, "recording.mp3", apiKey, options)
    if err != nil {
        log.Fatal(err)
    }
    
    // Detailed transcription with metadata
    detailedOptions := &openai.STTOptions{
        ResponseFormat: openai.STTFormatVerboseJSON,
        TimestampGranularities: []openai.TimestampGranularity{
            openai.GranularitySegment,
        },
    }
    
    response, err := openai.Speech2TextDetailed(audioData, "recording.mp3", apiKey, detailedOptions)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Text: %s\n", response.Text)
    fmt.Printf("Language: %s\n", response.Language)
    fmt.Printf("Duration: %.2fs\n", response.Duration)
}
```

## Examples

### 1. Text-to-Speech

Convert text to high-quality audio with various voices and formats.

**Available Models:**

- `openai.ModelTTS1` - Optimized for real-time use cases
- `openai.ModelTTS1HD` - Optimized for quality

**Available Voices:**

- `openai.VoiceAlloy`, `openai.VoiceEcho`, `openai.VoiceFable`
- `openai.VoiceNova`, `openai.VoiceOnyx`, `openai.VoiceShimmer`

**Available Formats:**

- `openai.FormatMP3`, `openai.FormatOpus`, `openai.FormatAAC`
- `openai.FormatFLAC`, `openai.FormatWAV`, `openai.FormatPCM`

See [examples/openai/text_to_speech/](examples/openai/text_to_speech/) for complete examples.

### 2. Speech-to-Text

Transcribe audio files to text with high accuracy using Whisper models.

**Available Models:**
- `openai.ModelWhisper1` - Whisper V2 model (default)
- `openai.ModelGPT4OTranscribe` - GPT-4o transcription model
- `openai.ModelGPT4OMiniTranscribe` - GPT-4o mini transcription model

**Response Formats:**
- `openai.STTFormatJSON` - JSON response with text (default)
- `openai.STTFormatText` - Plain text only
- `openai.STTFormatSRT` - SRT subtitle format
- `openai.STTFormatVerboseJSON` - Detailed JSON with timestamps
- `openai.STTFormatVTT` - VTT subtitle format

**Supported Audio Formats:**
- FLAC, MP3, MP4, MPEG, MPGA, M4A, OGG, WAV, WebM

See [examples/openai/speech_to_text/](examples/openai/speech_to_text/) for complete examples.

### 3. Standard Chat Completion

Simple question-answer interaction without structured output.

**CLI Request:**

```bash
go run cmd/openai/main.go \
  "You are a helpful assistant." \
  "What is the capital of France?"
```

**Programmatic Request:**

```go
response, err := openai.Prompt(
    "You are a helpful assistant.",
    "What is the capital of France?",
    "",
    apiKey,
)
```

### 4. Structured Output with JSON Schema

Data extraction with enforced JSON structure for reliable parsing.

**CLI Request:**

```bash
go run cmd/openai/main.go \
  "You are an expert at structured data extraction." \
  "Extract the author and title from: 'The Great Gatsby by F. Scott Fitzgerald'" \
  '{"name": "book_extraction", "description": "Extracts book information", "strict": true, "schema": {"type": "object", "properties": {"title": {"type": "string"}, "author": {"type": "string"}}, "required": ["title", "author"], "additionalProperties": false}}'
```

**Programmatic Request:**

```go
schema := `{"name": "book_extraction", "description": "Extracts book information", "strict": true, "schema": {"type": "object", "properties": {"title": {"type": "string"}, "author": {"type": "string"}}, "required": ["title", "author"], "additionalProperties": false}}`

response, err := openai.Prompt(
    "You are an expert at structured data extraction.",
    "Extract the author and title from: 'The Great Gatsby by F. Scott Fitzgerald'",
    schema,
    apiKey,
)
```

## Schema Requirements

When using structured output, the JSON schema must include:

- `name`: string (required)
- `description`: string (required)
- `strict`: true (required)
- `schema`: object (required)

## Error Handling

The package returns structured errors that can be handled appropriately:

```go
response, err := openai.Prompt(systemPrompt, userPrompt, schema, apiKey)
if err != nil {
    switch e := err.(type) {
    case *llmkit.APIError:
        fmt.Printf("API error: %s (status %d)\n", e.Message, e.StatusCode)
    case *llmkit.SchemaError:
        fmt.Printf("Schema validation error: %s\n", e.Message)
    case *llmkit.ValidationError:
        fmt.Printf("Input validation error: %s\n", e.Message)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```
