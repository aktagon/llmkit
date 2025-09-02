# OpenAI Temperature and MaxTokens Test

This example demonstrates the `Temperature` and `MaxTokens` options in the OpenAI ChatAgent API.

## What it tests:

1. **High Temperature (0.9)**: More creative, varied responses
2. **Low Temperature (0.1)**: More deterministic, precise responses  
3. **MaxTokens Limit**: Controlling response length
4. **Combined Options**: Using multiple options together

## Run the test:

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Expected behavior:

- **High temperature**: More creative and varied sentence about robots
- **Low temperature**: Consistent, precise answer to "What is 2+2?"
- **Token limit**: Short response that cuts off around 20 tokens
- **Combined**: Weather analysis with specific persona, temperature, and length

## API Consistency

The OpenAI ChatAgent uses the same API as Anthropic:

```go
response, err := agent.Chat("Hello", &agents.ChatOptions{
    Temperature: 0.8,
    MaxTokens:   100,
})
```

## Usage Statistics

OpenAI provides different token usage fields:
- `response.Raw.Usage.PromptTokens` - Input tokens
- `response.Raw.Usage.CompletionTokens` - Output tokens

## Key learnings:

- Temperature affects creativity vs determinism (same as Anthropic)
- MaxTokens controls response length (hard cutoff)
- Options work together seamlessly
- API is consistent between OpenAI and Anthropic providers

---
Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com

