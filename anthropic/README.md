# Anthropic API Example

A Go package that shows how to make Anthropic API calls with support for both
standard chat completions and structured output using JSON schemas.

## Setup

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

## CLI Usage

```bash
go run cmd/anthropic/main.go <system_prompt> <user_prompt> [json_schema]
```

## Programmatic Usage

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/aktagon/llmkit/anthropic"
)

func main() {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    
    response, err := anthropic.Prompt(
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

## Examples

### 1. Standard Chat Completion

Simple question-answer interaction without structured output.

**CLI Request:**

```bash
go run cmd/anthropic/main.go \
  "You are a helpful assistant." \
  "What is the capital of France?"
```

**Programmatic Request:**

```go
response, err := anthropic.Prompt(
    "You are a helpful assistant.",
    "What is the capital of France?",
    "",
    apiKey,
)
```

**Response:**

```json
{
  "id": "msg_01Hg8ijcD6z1rsPVNxzzSXA5",
  "type": "message",
  "role": "assistant",
  "model": "claude-sonnet-4-20250514",
  "content": [
    {
      "type": "text",
      "text": "The capital of France is Paris."
    }
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 20,
    "cache_creation_input_tokens": 0,
    "cache_read_input_tokens": 0,
    "output_tokens": 10,
    "service_tier": "standard"
  }
}
```

### 2. Structured Output with JSON Schema

Data extraction with enforced JSON structure for reliable parsing.

**CLI Request:**

```bash
go run cmd/anthropic/main.go \
  "You are an expert at structured data extraction." \
  "Extract the author and title from: 'The Great Gatsby by F. Scott Fitzgerald'" \
  '{"name": "book_extraction", "description": "Extracts book information", "strict": true, "schema": {"type": "object", "properties": {"title": {"type": "string"}, "author": {"type": "string"}}, "required": ["title", "author"], "additionalProperties": false}}'
```

**Programmatic Request:**

```go
schema := `{"name": "book_extraction", "description": "Extracts book information", "strict": true, "schema": {"type": "object", "properties": {"title": {"type": "string"}, "author": {"type": "string"}}, "required": ["title", "author"], "additionalProperties": false}}`

response, err := anthropic.Prompt(
    "You are an expert at structured data extraction.",
    "Extract the author and title from: 'The Great Gatsby by F. Scott Fitzgerald'",
    schema,
    apiKey,
)
```

**Response:**

```json
{
  "id": "msg_01EQCswcQiBc7BGqsxeKUdyD",
  "type": "message",
  "role": "assistant",
  "model": "claude-sonnet-4-20250514",
  "content": [
    {
      "type": "text",
      "text": "{\"title\": \"The Great Gatsby\", \"author\": \"F. Scott Fitzgerald\"}"
    }
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 132,
    "cache_creation_input_tokens": 0,
    "cache_read_input_tokens": 0,
    "output_tokens": 22,
    "service_tier": "standard"
  }
}
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
response, err := anthropic.Prompt(systemPrompt, userPrompt, schema, apiKey)
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
