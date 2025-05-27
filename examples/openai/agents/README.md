# Chat Agent Examples

This directory contains examples demonstrating how to use the OpenAI Chat Agent with tools.

## Prerequisites

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Examples

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
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param1"},
        "additionalProperties": false,
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

The chat agent handles the entire OpenAI function calling workflow automatically:

1. Sends your message to GPT
2. Detects when GPT wants to call functions
3. Executes the tool handlers
4. Sends results back to GPT
5. Returns GPT's final response

## Key Features

- **Conversation State**: Maintains full conversation history like Anthropic agent
- **Tool Execution**: Automatic tool calling and result handling
- **Error Handling**: Comprehensive error handling throughout
- **Identical API**: Same interface as Anthropic agent for consistency

## Comparison with Anthropic

Both agents now have identical structure and behavior:

| Feature                 | Anthropic        | OpenAI           |
| ----------------------- | ---------------- | ---------------- |
| **Registration**        | `RegisterTool()` | `RegisterTool()` |
| **Conversation State**  | Full history     | Full history     |
| **Auto Tool Execution** | ✅               | ✅               |
| **Multi-turn Chat**     | ✅               | ✅               |
| **Error Handling**      | ✅               | ✅               |
| **API Structure**       | Identical        | Identical        |

The only differences are in the underlying API calls and data formats, but the user experience is identical.
