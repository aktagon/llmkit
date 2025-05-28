# Memory Features Demo

This example demonstrates all the memory features of the ChatAgent:

## Features Demonstrated

### 1. **Memory as Context**
- Memory is automatically included in system prompts using `<memory></memory>` tags
- Claude sees past facts in every conversation
- Perfect for maintaining context across interactions

### 2. **Memory Tools**
- Claude can actively control memory using `remember_fact` and `recall_fact` tools
- Claude decides what to remember and when to recall
- Great for dynamic, intelligent memory management

### 3. **Memory Persistence**
- Memory automatically saves to disk
- Survives application restarts
- Essential for long-term user relationships

### 4. **Full Memory Suite**
- Combines all memory features for maximum capability
- Context + Tools + Persistence working together

## Usage

```bash
export ANTHROPIC_API_KEY="your-key"
go run main.go
```

## Expected Output

1. **Memory Context**: Claude will reference the pre-loaded facts about Alice
2. **Memory Tools**: Claude will use tools to remember facts about Bob
3. **Persistence**: Memory survives across agent restarts
4. **Full Suite**: Claude actively manages memory while maintaining context

## Memory Format

When memory context is enabled, Claude sees:

```
<memory>
user_name: Alice
favorite_color: blue
job: software engineer
</memory>
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

## Key Benefits

- **Automatic**: Memory context is included automatically
- **Intelligent**: Claude controls what to remember with tools
- **Persistent**: Memory survives restarts
- **Flexible**: Mix and match features as needed
- **Backward Compatible**: Existing code works unchanged
