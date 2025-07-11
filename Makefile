.PHONY: all build clean test fmt vet

# Default target
all: build

# Build all commands
build: llmkit llmkit-openai llmkit-anthropic

# Build unified llmkit command
llmkit:
	go build -o bin/llmkit ./cmd/llmkit

# Build OpenAI-specific command
llmkit-openai:
	go build -o bin/llmkit-openai ./cmd/llmkit-openai

# Build Anthropic-specific command
llmkit-anthropic:
	go build -o bin/llmkit-anthropic ./cmd/llmkit-anthropic

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run vet
vet:
	go vet ./...

# Install commands to GOPATH/bin
install:
	go install ./cmd/llmkit
	go install ./cmd/llmkit-openai
	go install ./cmd/llmkit-anthropic