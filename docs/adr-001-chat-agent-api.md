# ADR-001: Chat Agent API with Tool Support

## Status

Proposed

## Context

The current `anthropic/chat.go` implementation provides a basic one-shot request/response pattern without support for:
- Stateful conversations
- Tool calling and execution
- Multi-turn interactions

Based on Anthropic's recommendation for client tool integration, we need a chat agent API that supports:
1. Tool definitions with names, descriptions, and input schemas
2. Tool use detection and execution
3. Multi-turn conversations with state management
4. Tool result handling

## Decision

Implement a minimal `ChatAgent` API in `anthropic/agents/chat_agent.go` that:
- Maintains conversation state
- Supports tool registration and execution
- Handles the complete Anthropic tool workflow
- Provides a simple, stateful interface for multi-turn conversations

## Design

### Core Types

```go
// Tool represents a tool that can be called by Claude
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema interface{} `json:"input_schema"`
    Handler     ToolHandler `json:"-"`
}

// ToolHandler executes tool logic and returns results
type ToolHandler func(input map[string]interface{}) (string, error)

// ToolCall represents a tool use request from Claude
type ToolCall struct {
    ID    string                 `json:"id"`
    Name  string                 `json:"name"`
    Input map[string]interface{} `json:"input"`
}

// ChatAgent maintains conversation state and handles tool execution
type ChatAgent struct {
    client   *http.Client
    apiKey   string
    model    string
    messages []Message
    tools    map[string]Tool
}

// Message represents a conversation message
type Message struct {
    Role    string    `json:"role"`
    Content []Content `json:"content"`
}

// Content represents message content (text or tool use/result)
type Content struct {
    Type       string                 `json:"type"`
    Text       string                 `json:"text,omitempty"`
    ID         string                 `json:"id,omitempty"`
    Name       string                 `json:"name,omitempty"`
    Input      map[string]interface{} `json:"input,omitempty"`
    ToolUseID  string                 `json:"tool_use_id,omitempty"`
}
```

### API Interface

```go
// New creates a new ChatAgent
func New(apiKey string, options ...Option) *ChatAgent

// RegisterTool adds a tool that Claude can use
func (ca *ChatAgent) RegisterTool(tool Tool) error

// Chat sends a message and handles tool execution automatically
func (ca *ChatAgent) Chat(message string) (string, error)

// ChatWithTools sends a message and returns tool calls for manual handling
func (ca *ChatAgent) ChatWithTools(message string) (response string, toolCalls []ToolCall, err error)

// ExecuteToolCall manually executes a tool call and adds result to conversation
func (ca *ChatAgent) ExecuteToolCall(toolCall ToolCall) error

// Reset clears conversation history
func (ca *ChatAgent) Reset()
```

### Request/Response Examples

**Request with tools:**
```json
{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 1024,
    "tools": [{
        "name": "get_weather",
        "description": "Get the current weather in a given location",
        "input_schema": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city and state, e.g. San Francisco, CA"
                },
                "unit": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"],
                    "description": "The unit of temperature"
                }
            },
            "required": ["location"]
        }
    }],
    "messages": [
        {"role": "user", "content": [{"type": "text", "text": "What is the weather like in San Francisco?"}]}
    ]
}
```

**Response with tool use:**
```json
{
    "id": "msg_01Aq9w938a90dw8q",
    "model": "claude-sonnet-4-20250514",
    "stop_reason": "tool_use",
    "role": "assistant",
    "content": [
        {
            "type": "text", 
            "text": "I'll check the weather in San Francisco for you."
        },
        {
            "type": "tool_use",
            "id": "toolu_01A09q90qw90lq917835lq9",
            "name": "get_weather",
            "input": {"location": "San Francisco, CA", "unit": "celsius"}
        }
    ]
}
```

### Usage Example

```go
// Create agent
agent := agents.New(apiKey)

// Register weather tool
weatherTool := agents.Tool{
    Name:        "get_weather",
    Description: "Get the current weather in a given location",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "The city and state, e.g. San Francisco, CA",
            },
        },
        "required": []string{"location"},
    },
    Handler: func(input map[string]interface{}) (string, error) {
        location := input["location"].(string)
        // Call weather API...
        return "72°F, sunny", nil
    },
}

err := agent.RegisterTool(weatherTool)
if err != nil {
    log.Fatal(err)
}

// Chat with automatic tool execution
response, err := agent.Chat("What's the weather in San Francisco?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response) // "The weather in San Francisco is 72°F and sunny."
```

## Implementation Strategy

1. **Phase 1**: Core types and basic message handling
2. **Phase 2**: Tool registration and API request building  
3. **Phase 3**: Tool execution workflow and response parsing
4. **Phase 4**: Conversation state management
5. **Phase 5**: Error handling and validation

## Rationale

- **Minimal API**: Focus on essential tool workflow with simple interface
- **Stateful**: Maintains conversation context unlike current implementation
- **Flexible**: Supports both automatic and manual tool execution modes
- **Type-safe**: Leverages Go's type system for tool definitions
- **Extensible**: Tool handler pattern allows custom tool implementations
- **Familiar**: Similar patterns to existing Python implementation

## Consequences

**Positive:**
- Enables complex multi-turn conversations with tool support
- Maintains conversation state automatically
- Provides both high-level and low-level APIs
- Follows Anthropic's recommended tool integration pattern

**Negative:**
- More complex than current simple chat function
- Requires careful state management
- Additional dependencies for JSON handling

## Alternatives Considered

1. **Stateless tool functions**: Rejected - doesn't support conversation context
2. **Callback-based API**: Rejected - more complex than handler pattern
3. **Separate tool execution service**: Rejected - adds unnecessary complexity for minimal use case
