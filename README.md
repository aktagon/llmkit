# LLM Kit

Minimal Go library for calling LLM APIs using only the standard library - no
external dependencies required.

## Providers

- **Anthropic Claude** - Chat completions and structured output
- **OpenAI GPT** - Chat completions and structured output
- **Google Gemini** - Chat completions and structured output

## Structure

```
├── cmd/                    # Command-line interfaces
│   ├── llmkit-anthropic/   # Anthropic CLI
│   ├── llmkit-openai/      # OpenAI CLI
│   └── llmkit-google/      # Google CLI
├── anthropic/              # Anthropic (Claude) API package
│   ├── prompt.go           # API implementation
│   └── README.md           # Usage examples
├── openai/                 # OpenAI API package
│   ├── prompt.go           # API implementation
│   └── README.md           # Usage examples
├── google/                 # Google (Gemini) API package
│   ├── prompt.go           # API implementation
│   └── README.md           # Usage examples
├── docs/                   # API documentation
├── examples/               # Example JSON schemas
└── errors.go               # Structured error types
```

## Installation

### Homebrew (Recommended)

Install using Homebrew:

```bash
brew install aktagon/llmkit/llmkit
```

This installs the `llmkit` binary and all provider-specific CLI tools (`llmkit-anthropic`, `llmkit-openai`, `llmkit-google`).

### Install CLI Tools

Install the command-line tools globally:

```bash
# Install Anthropic CLI
go install github.com/aktagon/llmkit/cmd/llmkit-anthropic@latest

# Install OpenAI CLI
go install github.com/aktagon/llmkit/cmd/llmkit-openai@latest

# Install Google CLI
go install github.com/aktagon/llmkit/cmd/llmkit-google@latest
```

Make sure your `$GOPATH/bin` is in your `$PATH` to use the installed binaries:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

**Check installation location:**

```bash
# See where Go installs binaries
echo $(go env GOPATH)/bin

# List installed llmkit tools
ls -la $(go env GOPATH)/bin/llmkit-*
```

### Use as Library

Add to your Go project:

```bash
go get github.com/aktagon/llmkit
```

## Quick Start

### Anthropic

**Using installed CLI:**

```bash
export ANTHROPIC_API_KEY="your-key"
llmkit-anthropic "You are helpful" "Hello Claude"
```

**Using go run:**

```bash
export ANTHROPIC_API_KEY="your-key"
go run cmd/llmkit-anthropic/main.go "You are helpful" "Hello Claude"
```

**Structured output:**

```bash
llmkit-anthropic \
  "You are an expert at structured data extraction." \
  "What's the weather like in San Francisco? I prefer Celsius." \
  "$(cat examples/openai/schemas/weather-schema.json)"
```

### OpenAI

**Using installed CLI:**

```bash
export OPENAI_API_KEY="your-key"
llmkit-openai "You are helpful" "Hello GPT"
```

**Structured output:**

```bash
llmkit-openai \
  "You are an expert at structured data extraction." \
  "What's the weather like in San Francisco? I prefer Celsius." \
  "$(cat examples/openai/schemas/weather-schema.json)"
```

### Google

**Using installed CLI:**

```bash
export GOOGLE_API_KEY="your-key"
llmkit-google "You are helpful" "Hello Gemini"
```

**Structured output:**

```bash
llmkit-google \
  "You are an expert at structured data extraction." \
  "What's the weather like in San Francisco? I prefer Celsius." \
  "$(cat examples/google/schemas/weather-schema.json)"
```

## Programmatic Usage

### As a Library

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/aktagon/llmkit/anthropic"
    "github.com/aktagon/llmkit/openai"
    "github.com/aktagon/llmkit/google"
)

func main() {
    // Anthropic example
    anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
    response, err := anthropic.Prompt(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple prompt
        anthropicKey,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Anthropic:", response)

    // OpenAI example
    openaiKey := os.Getenv("OPENAI_API_KEY")
    response, err = openai.Prompt(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple prompt
        openaiKey,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("OpenAI:", response)

    // Google example
    googleKey := os.Getenv("GOOGLE_API_KEY")
    response, err = google.Prompt(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple prompt
        googleKey,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Google:", response)
}
```

### Structured Output Example

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/aktagon/llmkit/openai"
)

func main() {
    schema := `{
        "name": "weather_info",
        "description": "Weather information extraction",
        "strict": true,
        "schema": {
            "type": "object",
            "properties": {
                "location": {"type": "string"},
                "temperature": {"type": "number"},
                "unit": {"type": "string", "enum": ["C", "F"]}
            },
            "required": ["location", "temperature", "unit"],
            "additionalProperties": false
        }
    }`

    response, err := openai.Prompt(
        "You are a weather assistant.",
        "What's the weather in Tokyo? Use Celsius.",
        schema,
        os.Getenv("OPENAI_API_KEY"),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

## Features

- Standard chat completions
- Structured output with JSON schema validation/support
- Pure Go standard library implementation
- Command-line interface
- Structured error types for better error handling
- Programmatic API for library usage

## Error Handling

The library provides structured error types:

- `APIError` - Errors from LLM APIs
- `ValidationError` - Input validation errors
- `RequestError` - Request building/sending errors
- `SchemaError` - JSON schema validation errors

```go
response, err := openai.Prompt(systemPrompt, userPrompt, schema, apiKey)
if err != nil {
    switch e := err.(type) {
    case *errors.APIError:
        fmt.Printf("API error: %s (status %d)\n", e.Message, e.StatusCode)
    case *errors.SchemaError:
        fmt.Printf("Schema validation error: %s\n", e.Message)
    case *errors.ValidationError:
        fmt.Printf("Input validation error: %s\n", e.Message)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

Each provider directory contains detailed examples and usage instructions.

## Support

Commercial support is available. Contact christian@aktagon.com.

## License

MIT
