# ChatAgent with Memory

## Implementation Summary

The ChatAgent now has comprehensive memory functionality with three configurable modes:

### **Memory Features**

1. **Memory Context** - Automatic inclusion in system prompts with `<memory></memory>` tags
2. **Memory Tools** - Claude can actively manage memory using `remember_fact` and `recall_fact` tools
3. **Memory Persistence** - Automatic saving/loading from disk

### **New API**

```go
// Create agent with memory features
agent, err := agents.New(apiKey,
    agents.WithMemoryContext(),     // Auto-include memory in prompts
    agents.WithMemoryTools(),       // Enable memory tools
    agents.WithMemoryPersistence("memory.json"), // Auto-save to disk
)
```

### **Configuration Options**

#### **Functional Options Pattern**

```go
type AgentOption func(*ChatAgent) error

// Available options:
func WithMemoryContext() AgentOption
func WithMemoryTools() AgentOption
func WithMemoryPersistence(filepath string) AgentOption
```

#### **Memory Modes (Bitwise Flags)**

```go
type MemoryMode int

const (
    MemoryDisabled    MemoryMode = 0
    MemoryContext     MemoryMode = 1 << 0 // Include in system prompts
    MemoryTools       MemoryMode = 1 << 1 // Expose as tools
    MemoryPersistence MemoryMode = 1 << 2 // Save to disk
)
```

### **Usage Examples**

#### **1. Simple Memory Context**

```go
agent, err := agents.New(apiKey, agents.WithMemoryContext())

agent.Remember("user_name", "Alice")
agent.Remember("favorite_color", "blue")

response, err := agent.Chat("What do you know about me?")
// Claude sees: <memory>\nuser_name: Alice\nfavorite_color: blue\n</memory>
```

#### **2. Memory Tools (Claude-Controlled)**

```go
agent, err := agents.New(apiKey, agents.WithMemoryTools())

response, err := agent.Chat("Hi! I'm Bob and I love pizza. Remember this!")
// Claude automatically uses remember_fact tool to store information
```

#### **3. Memory with Persistence**

```go
agent, err := agents.New(apiKey,
    agents.WithMemoryContext(),
    agents.WithMemoryPersistence("user_memory.json"),
)

agent.Remember("session", "important_data")
// Automatically saves to disk

// Later - create new agent, memory is automatically loaded
agent2, err := agents.New(apiKey, agents.WithMemoryPersistence("user_memory.json"))
// Memory is restored from disk
```

#### **4. Full Memory Suite**

```go
agent, err := agents.New(apiKey,
    agents.WithMemoryContext(),     // Auto-context
    agents.WithMemoryTools(),       // Tool control
    agents.WithMemoryPersistence("memory.json"), // Persistence
)
```

### **Backward Compatibility**

** All existing code works unchanged:**

```go
// Old code still works
agent, err := agents.New(apiKey)
response, err := agent.Chat("Hello")
```

### **Updated Methods**

**Memory methods now return errors for persistence:**

```go
// Before
func (ca *ChatAgent) Remember(key, value string)
func (ca *ChatAgent) Forget(key string)
func (ca *ChatAgent) ClearMemory()
func (ca *ChatAgent) Reset(clearMemory bool)

// After
func (ca *ChatAgent) Remember(key, value string) error
func (ca *ChatAgent) Forget(key string) error
func (ca *ChatAgent) ClearMemory() error
func (ca *ChatAgent) Reset(clearMemory bool) error
```

### **How Memory Context Works**

When `WithMemoryContext()` is enabled, the system prompt automatically includes:

```
<memory>
user_name: Alice
favorite_color: blue
job: software engineer
</memory>

Your actual system prompt here...
```

Claude sees this context in every message, enabling seamless memory recall.

### **Memory Tools**

When `WithMemoryTools()` is enabled, Claude gets access to:

**`remember_fact`** - Store information

```json
{
  "name": "remember_fact",
  "description": "Store important information about the user for future conversations",
  "input_schema": {
    "type": "object",
    "properties": {
      "key": { "type": "string", "description": "What to remember" },
      "value": { "type": "string", "description": "The information" }
    },
    "required": ["key", "value"]
  }
}
```

**`recall_fact`** - Retrieve information

```json
{
  "name": "recall_fact",
  "description": "Retrieve previously stored information about the user",
  "input_schema": {
    "type": "object",
    "properties": {
      "key": { "type": "string", "description": "What to recall" }
    },
    "required": ["key"]
  }
}
```

### **Persistence Implementation**

Memory automatically saves to JSON:

```json
{
  "user_name": "Alice",
  "favorite_color": "blue",
  "last_seen": "2024-01-15",
  "preferences": "likes technical discussions"
}
```

**Features:**

- Auto-save on every memory change
- Auto-load on agent creation
- Graceful handling of missing files
- Error handling for disk operations

### **Key Benefits**

1. ** Flexible Configuration**: Mix and match memory features
2. ** Backward Compatible**: Existing code unchanged
3. ** Zero Overhead**: No memory features = no performance cost
4. ** Intelligent**: Claude actively manages memory when enabled
5. ** Persistent**: Memory survives application restarts
6. ** Type Safe**: Compile-time validation of options
7. ** Production Ready**: Proper error handling and persistence

### **Performance**

- **Memory Disabled**: Zero allocation overhead
- **Memory Context**: Minimal string concatenation cost
- **Memory Tools**: Only active when Claude chooses to use them
- **Memory Persistence**: Async I/O, doesn't block chat operations

### **Use Cases**

**Memory Context**: Perfect for customer service, personal assistants, tutoring
**Memory Tools**: Great for dynamic conversations where Claude learns about users
**Memory Persistence**: Essential for long-term user relationships, multi-session apps
