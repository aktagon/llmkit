# OpenAI Text-to-Speech Example

This example demonstrates how to use the OpenAI Text-to-Speech API with the llmkit library.

## Prerequisites

1. OpenAI API key with access to the TTS API
2. Go 1.24+ installed

## Setup

1. Set your OpenAI API key as an environment variable:
```bash
export OPENAI_API_KEY="your-api-key-here"
```

2. Navigate to this directory:
```bash
cd examples/openai/text_to_speech
```

3. Run the example:
```bash
go run main.go
```

## What it does

The example demonstrates three different use cases:

### 1. Basic Usage
- Uses default settings (tts-1 model, alloy voice, mp3 format)
- Generates `basic_example.mp3`

### 2. Advanced Usage
- Uses tts-1-hd model for higher quality
- Uses Nova voice
- Outputs WAV format
- Increased playback speed (1.2x)
- Generates `advanced_example.wav`

### 3. Voice Comparison
- Generates samples using all available voices
- Creates separate files for each voice: `voice_alloy.mp3`, `voice_echo.mp3`, etc.

## Available Options

### Models
- `openai.ModelTTS1` - Optimized for real-time use cases
- `openai.ModelTTS1HD` - Optimized for quality

### Voices
- `openai.VoiceAlloy` - Neutral, balanced
- `openai.VoiceEcho` - Clear, direct
- `openai.VoiceFable` - Expressive, storytelling
- `openai.VoiceNova` - Bright, optimistic
- `openai.VoiceOnyx` - Deep, authoritative
- `openai.VoiceShimmer` - Soft, whispery

### Audio Formats
- `openai.FormatMP3` - Default, widely compatible
- `openai.FormatOpus` - High compression efficiency
- `openai.FormatAAC` - Good balance of quality and size
- `openai.FormatFLAC` - Lossless compression
- `openai.FormatWAV` - Uncompressed, high quality
- `openai.FormatPCM` - Raw audio data

### Speed
- Range: 0.25 to 4.0
- Default: 1.0 (normal speed)
- 0.5 = half speed, 2.0 = double speed

## API Usage

```go
// Basic usage
audioData, err := openai.Text2Speech("Hello world", apiKey, nil)

// Advanced usage
options := &openai.TTSOptions{
    Model:          openai.ModelTTS1HD,
    Voice:          openai.VoiceNova,
    ResponseFormat: openai.FormatWAV,
    Speed:          1.2,
}
audioData, err := openai.Text2Speech("Hello world", apiKey, options)
```

## Limitations

- Maximum input text: 4096 characters
- Speed range: 0.25 to 4.0
- Returns binary audio data (save to file for playback)
- OpenAI usage policies require disclosure that voice is AI-generated

## Error Handling

The function returns standard llmkit errors:
- `ValidationError` for invalid input parameters
- `RequestError` for HTTP request issues
- `APIError` for OpenAI API errors

## Output Files

After running the example, you'll have several audio files:
- `basic_example.mp3` - Basic TTS example
- `advanced_example.wav` - High-quality TTS with custom options
- `voice_alloy.mp3`, `voice_echo.mp3`, etc. - Voice comparison samples

You can play these files with any audio player to hear the differences between voices and settings.

---
Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com

