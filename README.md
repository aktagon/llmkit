# LLM API Examples in Go

Minimal Go library for calling LLM APIs using only the standard library - no
external dependencies required.

## Providers

- **Anthropic Claude** - Chat completions and structured output
- **OpenAI GPT** - Chat completions and structured output

## Structure

```
├── cmd/                # Command-line interfaces
│   ├── anthropic/      # Anthropic CLI
│   └── openai/         # OpenAI CLI
├── anthropic/          # Anthropic (Claude) API package
│   ├── chat.go         # API implementation
│   └── README.md       # Usage examples
├── openai/             # OpenAI API package
│   ├── chat.go         # API implementation
│   └── README.md       # Usage examples
├── docs/               # API documentation
├── examples/           # Example JSON schemas
└── errors.go           # Structured error types
```

## Installation

### Install CLI Tools

Install the command-line tools globally:

```bash
# Install Anthropic CLI
go install github.com/aktagon/llmkit/cmd/anthropic@latest

# Install OpenAI CLI
go install github.com/aktagon/llmkit/cmd/openai@latest
```

Make sure your `$GOPATH/bin` is in your `$PATH` to use the installed binaries:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
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
anthropic "You are helpful" "Hello Claude"
```

**Using go run:**

```bash
export ANTHROPIC_API_KEY="your-key"
go run cmd/anthropic/main.go "You are helpful" "Hello Claude"
```

**Structured output:**

```bash
anthropic \
  "You are an expert at structured data extraction." \
  "What's the weather like in San Francisco? I prefer Celsius." \
  "$(cat examples/openai/schemas/weather-schema.json)"
```

### OpenAI

**Using installed CLI:**

```bash
export OPENAI_API_KEY="your-key"
openai "You are helpful" "Hello GPT"
```

**Using go run:**

```bash
export OPENAI_API_KEY="your-key"
go run cmd/openai/main.go "You are helpful" "Hello GPT"
```

**Structured output:**

```bash
openai \
  "You are an expert at structured data extraction." \
  "What's the weather like in San Francisco? I prefer Celsius." \
  "$(cat examples/openai/schemas/weather-schema.json)"
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
)

func main() {
    // Anthropic example
    anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
    response, err := anthropic.Chat(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple chat
        anthropicKey,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Anthropic:", response)

    // OpenAI example
    openaiKey := os.Getenv("OPENAI_API_KEY")
    response, err = openai.Chat(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple chat
        openaiKey,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("OpenAI:", response)
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

    response, err := openai.Chat(
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
- Structured output with **minimal** JSON schema validation/support
- Pure Go standard library implementation
- Command-line interface
- API for usage in Go projects

## Error Handling

The library provides structured error types:

- `APIError` - Errors from LLM APIs
- `ValidationError` - Input validation errors
- `RequestError` - Request building/sending errors
- `SchemaError` - JSON schema validation errors

```go
response, err := openai.Chat(systemPrompt, userPrompt, schema, apiKey)
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
