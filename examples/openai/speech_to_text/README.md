# OpenAI Speech-to-Text Example

This example demonstrates how to use the OpenAI Speech-to-Text (Whisper) API with the llmkit library.

## Prerequisites

1. OpenAI API key with access to the transcription API
2. Go 1.24+ installed
3. An audio file to transcribe

## Setup

1. Set your OpenAI API key as an environment variable:
```bash
export OPENAI_API_KEY="your-api-key-here"
```

2. Navigate to this directory:
```bash
cd examples/openai/speech_to_text
```

3. Run the simple example:
```bash
go run simple.go path/to/your/audio.mp3
```

4. Run the comprehensive example:
```bash
go run main.go path/to/your/audio.mp3
```

## What it does

### Simple Example (`simple.go`)
- Basic transcription with default settings
- Uses Whisper-1 model
- Returns plain text transcription

### Comprehensive Example (`main.go`)
The comprehensive example demonstrates four different use cases:

#### 1. Basic Transcription
- Uses default settings (Whisper-1 model, JSON format)
- Simple text output

#### 2. Advanced Transcription
- Custom language specification
- Custom prompt for better accuracy
- Temperature control for deterministic output

#### 3. Detailed Transcription
- Verbose JSON response with metadata
- Timestamp granularities (segments)
- Language detection and duration info

#### 4. Different Response Formats
- Plain text format
- SRT subtitle format (saved to file)
- Demonstrates format flexibility

## API Usage

### Basic Usage
```go
// Simple transcription
text, err := openai.Speech2Text(audioData, filename, apiKey, nil)
```

### Advanced Usage
```go
// With custom options
options := &openai.STTOptions{
    Model:       openai.ModelWhisper1,
    Language:    "en",
    Prompt:      "This audio contains technical terms.",
    Temperature: 0.2,
}
text, err := openai.Speech2Text(audioData, filename, apiKey, options)
```

### Detailed Response
```go
// Get detailed response with metadata
options := &openai.STTOptions{
    ResponseFormat: openai.STTFormatVerboseJSON,
    TimestampGranularities: []openai.TimestampGranularity{
        openai.GranularitySegment,
    },
}
response, err := openai.Speech2TextDetailed(audioData, filename, apiKey, options)
```

## Available Options

### Models
- `openai.ModelWhisper1` - Whisper V2 model (default)
- `openai.ModelGPT4OTranscribe` - GPT-4o transcription model
- `openai.ModelGPT4OMiniTranscribe` - GPT-4o mini transcription model

### Response Formats
- `openai.STTFormatJSON` - JSON response (default)
- `openai.STTFormatText` - Plain text
- `openai.STTFormatSRT` - SRT subtitle format
- `openai.STTFormatVerboseJSON` - Detailed JSON with metadata
- `openai.STTFormatVTT` - VTT subtitle format

### Timestamp Granularities
- `openai.GranularitySegment` - Segment-level timestamps
- `openai.GranularityWord` - Word-level timestamps

### Other Options
- **Language**: ISO-639-1 language code (e.g., "en", "es", "fr")
- **Prompt**: Text to guide the model's style or context
- **Temperature**: 0.0 to 1.0 (deterministic to random)

## Supported Audio Formats

- **FLAC** (.flac) - Lossless compression
- **MP3** (.mp3) - Popular compressed format
- **MP4** (.mp4) - Video container with audio
- **MPEG** (.mpeg) - Video/audio format
- **MPGA** (.mpga) - MPEG audio
- **M4A** (.m4a) - Apple audio format
- **OGG** (.ogg) - Open source format
- **WAV** (.wav) - Uncompressed audio
- **WebM** (.webm) - Web-optimized format

## Two API Functions

### `Speech2Text()` - Simple Interface
- Returns transcribed text as string
- Good for basic transcription needs
- Handles different response formats automatically

### `Speech2TextDetailed()` - Full Response
- Returns complete `STTResponse` struct
- Includes metadata like language, duration
- Provides timestamps and segments
- Best for applications needing detailed information

## Error Handling

The function returns standard llmkit errors:
- `ValidationError` for invalid input parameters
- `RequestError` for HTTP/multipart form issues  
- `APIError` for OpenAI API errors

## Example Output

```
Reading audio file: sample.mp3
✓ Audio file loaded (245760 bytes)

Example 1: Basic transcription (Whisper-1 model)
==================================================
Transcribed text: Hello, this is a test recording for speech-to-text transcription.

Example 2: Advanced transcription with custom options
==================================================
Transcribed text: Hello, this is a test recording for speech-to-text transcription.

Example 3: Detailed transcription with timestamps
==================================================
Full text: Hello, this is a test recording for speech-to-text transcription.
Detected language: en
Duration: 3.84 seconds

Segments with timestamps:
  [1] 0.00s - 3.84s: Hello, this is a test recording for speech-to-text transcription.
```

## Tips

1. **File Size**: Larger files may take longer to process
2. **Quality**: Better audio quality = better transcription accuracy
3. **Language**: Specify the language for better accuracy
4. **Prompts**: Use prompts to improve transcription of technical terms or proper names
5. **Formats**: Use SRT/VTT for subtitle applications
6. **Temperature**: Use 0.0-0.2 for consistent results, higher for more variation

## Limitations

- File size limits apply (check OpenAI documentation)
- Some models support only specific response formats
- Word-level timestamps may increase processing time
- Streaming not supported for Whisper-1 model
