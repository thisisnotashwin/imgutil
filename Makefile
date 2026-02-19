BINARY  := imgutil
MODULE  := github.com/thisisnotashwin/imgutil
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION)"
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build binary for current platform → ./bin/imgutil
	@mkdir -p bin
	@go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/imgutil

.PHONY: clean
clean: ## Remove build artefacts (bin/ and dist/)
	@rm -rf bin dist

.PHONY: test
test: ## Run all tests with race detector
	@go test -race -v ./...

.PHONY: lint
lint: ## Run golangci-lint
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code and run go vet
	@gofmt -w .
	@go vet ./...

.PHONY: install
install: ## Install binary to $GOPATH/bin
	@go install $(LDFLAGS) ./cmd/imgutil

.PHONY: release
release: ## Cross-compile for all platforms → ./dist/
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		echo "Building $(OS)/$(ARCH)..."; \
		GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) \
			-o dist/$(BINARY)-$(OS)-$(ARCH) ./cmd/imgutil; \
	)
