# OpenAI API Example

A Go package that shows how to make OpenAI API calls with support for both
standard chat completions and structured output using JSON schemas.

## Setup

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## CLI Usage

```bash
go run cmd/openai/main.go <system_prompt> <user_prompt> [json_schema]
```

## Programmatic Usage

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
    
    response, err := openai.Chat(
        "You are a helpful assistant",
        "What is the capital of France?",
        "", // no schema for simple chat
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
go run cmd/openai/main.go \
  "You are a helpful assistant." \
  "What is the capital of France?"
```

**Programmatic Request:**

```go
response, err := openai.Chat(
    "You are a helpful assistant.",
    "What is the capital of France?",
    "",
    apiKey,
)
```

### 2. Structured Output with JSON Schema

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

response, err := openai.Chat(
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
