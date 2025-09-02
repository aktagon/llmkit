# Chat Agent API

The new `ChatAgent` provides an intuitive API for stateful conversations with Claude, including persistent memory and structured output support.

## Features

- **Stateful Conversations**: Maintains conversation history automatically
- **Persistent Memory**: Key-value store for facts that persist across conversations
- **Structured Output**: JSON schema support for reliable data extraction
- **Tool Support**: Register and execute tools during conversations
- **Flexible Options**: Optional system prompts and schemas per message

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
    Text string                        // Extracted text response
    Raw  *anthropic.AnthropicResponse  // Full API response with usage stats
}
```

### Main Methods

```go
// Chat with optional parameters
func (ca *ChatAgent) Chat(message string, opts ...*ChatOptions) (*ChatResponse, error)

// Memory management
func (ca *ChatAgent) Remember(key, value string)
func (ca *ChatAgent) Recall(key string) (string, bool)
func (ca *ChatAgent) Forget(key string)
func (ca *ChatAgent) ClearMemory()
func (ca *ChatAgent) GetMemory() map[string]string

// Conversation management
func (ca *ChatAgent) Reset(clearMemory bool)
func (ca *ChatAgent) RegisterTool(tool anthropic.Tool) error
```

## Usage Examples

### 1. Simple Chat

```go
agent := agents.New(apiKey)

response, err := agent.Chat("Hello! How are you?")
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Text)
fmt.Printf("Used %d tokens\n", response.Raw.Usage.InputTokens)
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

### 4. Structured Output

```go
schema := `{
    "name": "person_info",
    "description": "Extract person information",
    "strict": true,
    "schema": {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "number"}
        },
        "required": ["name", "age"],
        "additionalProperties": false
    }
}`

response, err := agent.Chat("I'm Alice, 25 years old", &agents.ChatOptions{
    Schema: schema,
})
// response.Text will contain: {"name": "Alice", "age": 25}
```

### 5. Memory Management

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

// Use in conversation
response, err := agent.Chat(fmt.Sprintf("Hi %s! Do you still like %s?", 
    name, agent.Recall("favorite_food")))
```

### 6. Conversation Management

```go
// Reset conversation history, keep memory
agent.Reset(false)

// Reset everything
agent.Reset(true)

// Clear just memory
agent.ClearMemory()
```

### 7. Tool Integration

```go
weatherTool := anthropic.Tool{
    Name:        "get_weather",
    Description: "Get current weather for a location",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type": "string",
                "description": "City name",
            },
        },
        "required": []string{"location"},
    },
    Handler: func(input map[string]interface{}) (string, error) {
        location := input["location"].(string)
        return fmt.Sprintf("Weather in %s: 72°F, sunny", location), nil
    },
}

agent.RegisterTool(weatherTool)
response, err := agent.Chat("What's the weather in Paris?")
// Claude will automatically call the weather tool
```

### 8. Combining Options

```go
// Use multiple options together
response, err := agent.Chat("Analyze this data and format as JSON", &agents.ChatOptions{
    SystemPrompt: "You are a data analyst. Be precise and thorough.",
    Schema:       jsonSchema,
    Temperature:  0.3,
    MaxTokens:    200,
})
```

## Key Benefits

1. **Simple for Basic Use**: `agent.Chat("Hello")` just works
2. **Powerful for Advanced Use**: Options pattern allows complex scenarios
3. **Memory Persistence**: Facts survive conversation resets
4. **Full Access**: Both processed text and raw response always available
5. **Type Safe**: Compile-time checking for all parameters
6. **Extensible**: Easy to add new options without breaking existing code
7. **Fine Control**: Adjust creativity (temperature) and response length (max tokens)

## Parameter Ranges

- **Temperature**: 0.0 to 1.0 (0.0 = deterministic, 1.0 = very creative, -1 = use default)
- **MaxTokens**: 1 to model limit (0 = use default 4096 for Claude)
- Simple values, no pointers needed!

## Memory vs Conversation History

- **Conversation History** (`messages`): Full context of the current conversation
- **Persistent Memory** (`memory`): Key-value facts that survive across conversations
- Use `Reset(false)` to start fresh conversations while keeping learned facts
- Use `Reset(true)` to completely start over

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

---
Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com

