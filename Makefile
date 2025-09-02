.PHONY: all build clean test fmt vet compile-examples

# Default target
all: build

# Build all commands
build: llmkit llmkit-openai llmkit-anthropic llmkit-google

# Build unified llmkit command
llmkit:
	go build -o bin/llmkit ./cmd/llmkit

# Build OpenAI-specific command
llmkit-openai:
	go build -o bin/llmkit-openai ./cmd/llmkit-openai

# Build Anthropic-specific command
llmkit-anthropic:
	go build -o bin/llmkit-anthropic ./cmd/llmkit-anthropic

# Build Google-specific command
llmkit-google:
	go build -o bin/llmkit-google ./cmd/llmkit-google

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
	go install ./cmd/llmkit-google

# Compile all examples
compile-examples:
	@echo "Compiling examples..."
	@for dir in $$(find examples -name Makefile -exec dirname {} \;); do \
		echo "Compiling $$dir..."; \
		cd $$dir && make compile && cd - > /dev/null || exit 1; \
	done
	@echo "All examples compiled successfully"