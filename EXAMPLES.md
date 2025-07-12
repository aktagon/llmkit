# Examples

## Speech-to-Text

```bash
# OpenAI Speech-to-Text
cd examples/openai/speech_to_text
go run main.go sample.mp3
```

## Text-to-Speech

```bash
# OpenAI Text-to-Speech
cd examples/openai/text_to_speech
go run main.go
```

## Chat Agents

```bash
# Anthropic Chat Agent
cd examples/anthropic/agents/chat_example
go run main.go

# OpenAI Chat Agent
cd examples/openai/agents/chat_example
go run main.go
```

## Weather Agents (Interactive)

```bash
# Anthropic Weather Agent
cd examples/anthropic/agents/weather_agent
go run main.go

# OpenAI Weather Agent
cd examples/openai/agents/weather_agent
go run main.go
```

## Temperature & Token Tests

```bash
# Anthropic Temperature Test
cd examples/anthropic/agents/temperature_test
go run main.go

# OpenAI Temperature Test
cd examples/openai/agents/temperature_test
go run main.go
```

## Memory Demos

```bash
# Anthropic Memory Demo
cd examples/anthropic/agents/memory_demo
go run main.go

# OpenAI Memory Demo
cd examples/openai/agents/memory_demo
go run main.go
```

Set `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` environment variables as needed.