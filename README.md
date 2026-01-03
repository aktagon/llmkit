# LLMKit

Minimal Go library for calling LLM APIs. Unified interface for OpenAI GPT, Anthropic Claude, Google Gemini, and xAI Grok. One package, three functions, zero magic.

> **Feature Freeze:** The API is stable. No new features will be added. Bug fixes only.

## Install

```bash
go get github.com/aktagon/llmkit
```

## Usage

```go
provider := llmkit.Provider{
    Name:   "anthropic",
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
}

resp, err := llmkit.Prompt(ctx, provider, llmkit.Request{
    User: "Hello",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Text)
fmt.Printf("Tokens: %d in, %d out\n", resp.Tokens.Input, resp.Tokens.Output)
```

### System Prompt

```go
resp, err := llmkit.Prompt(ctx, provider, llmkit.Request{
    System: "You are a helpful assistant.",
    User:   "Hello",
})
```

### Custom Model

```go
provider := llmkit.Provider{
    Name:   "anthropic",
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    Model:  "claude-3-opus-20240229",
}
```

### Custom Base URL

Use any OpenAI-compatible API (LiteLLM, vLLM, Ollama, etc.):

```go
provider := llmkit.Provider{
    Name:    "openai",
    APIKey:  os.Getenv("OPENAI_API_KEY"),
    BaseURL: "http://localhost:4000",
}
```

## Providers

| Provider  | Name        | Default Model       | Env Var             |
| --------- | ----------- | ------------------- | ------------------- |
| Anthropic | `anthropic` | claude-sonnet-4-5   | `ANTHROPIC_API_KEY` |
| OpenAI    | `openai`    | gpt-4o-2024-08-06   | `OPENAI_API_KEY`    |
| Google    | `google`    | gemini-2.5-flash    | `GEMINI_API_KEY`    |
| Grok      | `grok`      | grok-3-fast         | `XAI_API_KEY`       |

## Feature Matrix

| Feature           | Anthropic | OpenAI | Google | Grok |
| ----------------- | --------- | ------ | ------ | ---- |
| Prompt            | Y         | Y      | Y      | Y    |
| Agent             | Y         | Y      | Y      | Y    |
| Tools             | Y         | Y      | Y      | Y    |
| Structured Output | Y         | Y      | Y      | Y    |
| File Upload       | Y         | Y      | Y      | Y    |
| Image Input       | Y         | Y      | Y      | Y    |

## Option Support Matrix

| Option                  | Anthropic | OpenAI      | Google       | Grok       |
| ----------------------- | --------- | ----------- | ------------ | ---------- |
| `WithTemperature`       | Y         | Y           | Y            | Y          |
| `WithTopP`              | Y         | Y           | Y            | Y          |
| `WithTopK`              | Y         | -           | Y            | Y          |
| `WithMaxTokens`         | Y (req)   | Y           | Y            | Y          |
| `WithStopSequences`     | Y         | Y           | Y (max 5)    | Grok-3     |
| `WithSeed`              | -         | Y           | Y            | Y          |
| `WithFrequencyPenalty`  | -         | Y           | -            | Grok-3     |
| `WithPresencePenalty`   | -         | Y           | -            | Grok-3     |
| `WithThinkingBudget`    | Y (≥1024) | -           | Gemini 2.5   | -          |
| `WithReasoningEffort`   | -         | Y (o-series)| Gemini 3     | -          |

## API

```go
func Prompt(ctx context.Context, p Provider, req Request) (Response, error)
func NewAgent(p Provider) *Agent
func UploadFile(ctx context.Context, p Provider, path string) (File, error)
```

## License

FSL-1.1-Apache-2.0 - Free for internal use, education, and research. Converts to Apache 2.0 after 2 years. See [LICENSE](LICENSE).
