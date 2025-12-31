.PHONY: build test clean lint vet fmt check release version help

BINARY_NAME := llmkit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

BUILD_DIR := build
GO := go

help:
	@echo "llmkit - Unified LLM Provider Interface"
	@echo ""
	@echo "Usage:"
	@echo "  make build      Build binaries for all platforms"
	@echo "  make test       Run tests"
	@echo "  make vet        Run go vet"
	@echo "  make fmt        Format code"
	@echo "  make check      Run all checks (fmt, vet, test)"
	@echo "  make clean      Remove build artifacts"
	@echo "  make version    Show current version"
	@echo "  make release    Trigger a release (push to master)"
	@echo ""

build:
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GO) build -ldflags "-X main.Version=$(VERSION)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$$([ "$${platform%/*}" = "windows" ] && echo ".exe") . ; \
		echo "Built $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done

test:
	$(GO) test -v -race ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...
	@echo "Code formatted"

check: fmt vet test

clean:
	rm -rf $(BUILD_DIR)
	$(GO) clean

version:
	@echo "Current version: $(VERSION)"
	@echo "Latest tag: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'no tags')"

release:
	@echo "Release is automated via GitHub Actions."
	@echo ""
	@echo "To trigger a release:"
	@echo "  1. Push to master branch (auto-triggers release)"
	@echo "  2. Or manually trigger via GitHub Actions UI"
	@echo ""
	@echo "Current version: $(VERSION)"
	@git status --short
