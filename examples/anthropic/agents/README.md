# Chat Agent Examples

This directory contains examples demonstrating how to use the Anthropic Chat Agent with tools.

## Prerequisites

Set your Anthropic API key:
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

## Examples

### Simple Weather Tool (`weather_simple/`)

A minimal example showing how to create and use a weather tool:

```bash
cd weather_simple
go run main.go
```

Shows the basic pattern:
1. Create a chat agent
2. Define a tool with name, description, schema, and handler
3. Register the tool
4. Chat - tool execution happens automatically

### Interactive Weather Agent (`weather_agent/`)

A complete interactive chat application with a weather tool:

```bash
cd weather_agent
go run main.go
```

Features:
- Interactive command-line interface
- Mock weather data for multiple cities
- Support for Celsius and Fahrenheit
- Flexible city name matching
- Conversation continues until user types "exit"

Try asking:
- "What's the weather in San Francisco?"
- "How's the weather in New York in Celsius?"
- "Tell me about London's weather"

## Tool Definition Pattern

All tools follow this pattern:

```go
tool := agents.Tool{
    Name:        "tool_name",
    Description: "What the tool does",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param1"},
    },
    Handler: func(input map[string]interface{}) (string, error) {
        // Extract parameters
        param1 := input["param1"].(string)
        
        // Do work
        result := doSomething(param1)
        
        // Return result
        return result, nil
    },
}
```

The chat agent handles the entire Anthropic tool workflow automatically:
1. Sends your message to Claude
2. Detects when Claude wants to use tools
3. Executes the tool handlers
4. Sends results back to Claude
5. Returns Claude's final response
