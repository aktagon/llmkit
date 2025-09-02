# OpenAI Memory Features Demo

This example demonstrates all the memory features of the OpenAI ChatAgent:

## Features Demonstrated

### 1. **Memory as Context**
- Memory is automatically included in system messages using `<memory></memory>` tags
- GPT sees past facts in every conversation
- Perfect for maintaining context across interactions

### 2. **Memory Tools**
- GPT can actively control memory using `remember_fact` and `recall_fact` functions
- GPT decides what to remember and when to recall
- Great for dynamic, intelligent memory management

### 3. **Memory Persistence**
- Memory automatically saves to disk
- Survives application restarts
- Essential for long-term user relationships

### 4. **Full Memory Suite**
- Combines all memory features for maximum capability
- Context + Tools + Persistence working together

### 5. **Advanced Options**
- Memory works seamlessly with temperature, max tokens, and system prompts
- All ChatOptions work with memory features

## Usage

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Expected Output

1. **Memory Context**: GPT will reference the pre-loaded facts about Alice
2. **Memory Tools**: GPT will use functions to remember facts about Bob
3. **Persistence**: Memory survives across agent restarts
4. **Full Suite**: GPT actively manages memory while maintaining context
5. **Advanced**: Memory combined with creative parameters

## Memory Format

When memory context is enabled, GPT sees:

```
{
  "role": "system",
  "content": "<memory>\nuser_name: Alice\nfavorite_color: blue\njob: software engineer\n</memory>\n\nYour prompt here"
}
```

## Configuration Options

```go
// Memory as context only
agent, _ := agents.New(apiKey, agents.WithMemoryContext())

// Memory tools only
agent, _ := agents.New(apiKey, agents.WithMemoryTools())

// Memory persistence only
agent, _ := agents.New(apiKey, agents.WithMemoryPersistence("memory.json"))

// All memory features
agent, _ := agents.New(apiKey,
    agents.WithMemoryContext(),
    agents.WithMemoryTools(),
    agents.WithMemoryPersistence("memory.json"),
)
```

## API Consistency

The OpenAI ChatAgent has the exact same API as the Anthropic ChatAgent:

```go
// Works identically for both providers
agent, err := agents.New(apiKey,
    agents.WithMemoryContext(),
    agents.WithMemoryTools(),
    agents.WithMemoryPersistence("memory.json"),
)

response, err := agent.Chat("Hello", &agents.ChatOptions{
    Temperature: 0.8,
    MaxTokens:   100,
})
```

## Key Benefits

- **Provider Consistency**: Same API across Anthropic and OpenAI
- **Automatic**: Memory context is included automatically
- **Intelligent**: GPT controls what to remember with functions
- **Persistent**: Memory survives restarts
- **Flexible**: Mix and match features as needed
- **Backward Compatible**: Existing code works unchanged

---
Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com

