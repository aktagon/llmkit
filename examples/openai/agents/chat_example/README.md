# OpenAI Chat Agent API

The OpenAI ChatAgent provides an intuitive API for stateful conversations with GPT, including persistent memory and structured output support.

## Features

- **Stateful Conversations**: Maintains conversation history automatically
- **Persistent Memory**: Key-value store for facts that persist across conversations
- **Structured Output**: JSON schema support for reliable data extraction
- **Tool Support**: Register and execute functions during conversations
- **Flexible Options**: Optional system prompts and schemas per message
- **API Consistency**: Identical API to Anthropic ChatAgent

## API Overview

### Core Types

```go
type ChatOptions struct {
    Schema       string  // JSON schema for structured output
    SystemPrompt string  // System prompt for this specific message
    Temperature  float64 // Temperature for response randomness (0.0-1.0, -1 = use default)
    MaxTokens    int     // Maximum tokens in response (0 = use default)
}

type ChatResponse struct {
    Text string          // Extracted text response
    Raw  *openai.Response // Full API response with usage stats
}
```

### Main Methods

```go
// Chat with optional parameters
func (ca *ChatAgent) Chat(message string, opts ...*ChatOptions) (*ChatResponse, error)

// Memory management
func (ca *ChatAgent) Remember(key, value string) error
func (ca *ChatAgent) Recall(key string) (string, bool)
func (ca *ChatAgent) Forget(key string) error
func (ca *ChatAgent) ClearMemory() error
func (ca *ChatAgent) GetMemory() map[string]string

// Conversation management
func (ca *ChatAgent) Reset(clearMemory bool) error
func (ca *ChatAgent) RegisterTool(tool openai.Tool) error
```

## Usage Examples

### 1. Simple Chat

```go
agent, err := agents.New(apiKey)

response, err := agent.Chat("Hello! How are you?")
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Text)
fmt.Printf("Used %d tokens\n", response.Raw.Usage.PromptTokens)
```

### 2. Chat with System Prompt

```go
response, err := agent.Chat("What's the weather like?", &agents.ChatOptions{
    SystemPrompt: "You are a helpful weather assistant.",
})
```

### 3. Control Response Generation

```go
// Control creativity and length
response, err := agent.Chat("Write a poem", &agents.ChatOptions{
    Temperature: 0.8,  // Higher = more creative
    MaxTokens:   50,
})

// Conservative, precise responses
response, err = agent.Chat("What is 2+2?", &agents.ChatOptions{
    Temperature: 0.1,
})
```

### 4. Memory Management

```go
// Store facts
agent.Remember("user_name", "Alice")
agent.Remember("favorite_food", "pizza")

// Retrieve facts
name, exists := agent.Recall("user_name")
if exists {
    fmt.Printf("User name: %s\n", name)
}

// View all memory
memory := agent.GetMemory()
fmt.Printf("All facts: %+v\n", memory)
```

### 5. Memory Features

```go
// Memory as context (automatic inclusion)
agent, err := agents.New(apiKey, agents.WithMemoryContext())

// Memory tools (GPT controls memory)
agent, err := agents.New(apiKey, agents.WithMemoryTools())

// Memory persistence (auto-save/load)
agent, err := agents.New(apiKey, agents.WithMemoryPersistence("memory.json"))

// All memory features
agent, err := agents.New(apiKey,
    agents.WithMemoryContext(),
    agents.WithMemoryTools(),
    agents.WithMemoryPersistence("memory.json"),
)
```

## Key Differences from Anthropic

### System Message Format
- **OpenAI**: System prompts are added as system role messages in the messages array
- **Anthropic**: System prompts use a separate `system` field

### Tool Format
- **OpenAI**: Uses `functions` with `Parameters` field
- **Anthropic**: Uses `tools` with `InputSchema` field

### Response Format
- **OpenAI**: `response.Raw.Usage.PromptTokens` and `CompletionTokens`
- **Anthropic**: `response.Raw.Usage.InputTokens` and `OutputTokens`

## API Consistency

Despite the differences above, the **user-facing API is identical** between providers:

```go
// This code works for both OpenAI and Anthropic
agent, err := agents.New(apiKey,
    agents.WithMemoryContext(),
    agents.WithMemoryTools(),
    agents.WithMemoryPersistence("memory.json"),
)

response, err := agent.Chat("Hello", &agents.ChatOptions{
    SystemPrompt: "You are helpful",
    Temperature:  0.8,
    MaxTokens:    100,
})

fmt.Println(response.Text)
```

## Memory Context Example

When memory context is enabled, GPT sees:

```json
{
  "role": "system",
  "content": "<memory>\nuser_name: Alice\nfavorite_color: blue\njob: software engineer\n</memory>\n\nYou are a helpful assistant."
}
```

## Error Handling

All methods return structured errors for proper handling:

```go
response, err := agent.Chat("Hello")
if err != nil {
    switch e := err.(type) {
    case *llmkit.APIError:
        fmt.Printf("API error: %s\n", e.Message)
    case *llmkit.ValidationError:
        fmt.Printf("Validation error: %s\n", e.Message)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

## Key Benefits

1. **Provider Consistency**: Same API as Anthropic ChatAgent
2. **Simple for Basic Use**: `agent.Chat("Hello")` just works
3. **Powerful for Advanced Use**: Options pattern allows complex scenarios
4. **Memory Persistence**: Facts survive conversation resets
5. **Full Access**: Both processed text and raw response always available
6. **Type Safe**: Compile-time checking for all parameters
7. **Backward Compatible**: Existing code works unchanged

---
Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com

