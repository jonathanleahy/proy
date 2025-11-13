# HTTP Testing Proxy - Makefile
# Professional build configuration for testers

.PHONY: help build test run clean docker docker-run docker-stop lint fmt coverage deps

# Variables
BINARY_NAME=proxy
DOCKER_IMAGE=testing-proxy
DOCKER_TAG=latest
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")
COVERAGE_FILE=coverage.out

# Default target - show help
help: ## Show this help message
	@echo "HTTP Testing Proxy - Build Commands"
	@echo "===================================="
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ""
	@echo "Usage: make [target]"
	@echo ""

build: ## Build the proxy binary
	@echo "🔨 Building proxy binary..."
	@go build -o $(BINARY_NAME) cmd/proxy/main.go
	@echo "✅ Build complete: ./$(BINARY_NAME)"

test: ## Run all tests
	@echo "🧪 Running unit tests..."
	@go test -v ./internal/... -cover
	@echo "✅ All tests passed!"

test-verbose: ## Run tests with verbose output
	@echo "🧪 Running tests with verbose output..."
	@go test -v ./... -cover

coverage: ## Generate test coverage report
	@echo "📊 Generating coverage report..."
	@go test -coverprofile=$(COVERAGE_FILE) ./internal/...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@echo "📈 Coverage summary:"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

run: ## Run the proxy server
	@echo "🚀 Starting proxy server..."
	@go run cmd/proxy/main.go

run-record: ## Run proxy in record mode
	@echo "🔴 Starting proxy in RECORD mode..."
	@go run cmd/proxy/main.go -mode=record

run-playback: ## Run proxy in playback mode
	@echo "▶️  Starting proxy in PLAYBACK mode..."
	@go run cmd/proxy/main.go -mode=playback

clean: ## Clean build artifacts and recordings
	@echo "🧹 Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE) coverage.html
	@echo "⚠️  To clean recordings, run: make clean-recordings"
	@echo "✅ Clean complete!"

clean-recordings: ## Clear all recorded interactions
	@echo "⚠️  This will delete all recordings. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	@rm -rf ./recordings/*
	@echo "✅ All recordings cleared!"

docker: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run proxy in Docker container
	@echo "🐳 Running proxy in Docker..."
	@docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 8080:8080 \
		-v $(PWD)/recordings:/app/recordings \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Proxy running in Docker!"
	@echo "📍 Access at: http://0.0.0.0:8080"

docker-stop: ## Stop and remove Docker container
	@echo "🛑 Stopping Docker container..."
	@docker stop $(DOCKER_IMAGE) || true
	@docker rm $(DOCKER_IMAGE) || true
	@echo "✅ Container stopped and removed"

lint: ## Run Go linter
	@echo "🔍 Running linter..."
	@golangci-lint run || (echo "❌ Please install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	@echo "✅ Linting passed!"

fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@gofmt -s -w $(GO_FILES)
	@goimports -w $(GO_FILES) || (echo "⚠️  goimports not found, using gofmt only" && true)
	@echo "✅ Code formatted!"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated!"

install: build ## Install proxy to system
	@echo "📥 Installing proxy..."
	@cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME) || (echo "❌ Need sudo permissions" && sudo cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME))
	@echo "✅ Proxy installed to /usr/local/bin/$(BINARY_NAME)"

uninstall: ## Uninstall proxy from system
	@echo "🗑️  Uninstalling proxy..."
	@rm -f /usr/local/bin/$(BINARY_NAME) || sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Proxy uninstalled"

dev: ## Run proxy with hot reload (requires air)
	@which air > /dev/null || (echo "❌ Please install air: go install github.com/cosmtrek/air@latest" && exit 1)
	@echo "🔄 Starting with hot reload..."
	@air

# CI/CD targets
ci: deps lint test coverage ## Run CI pipeline
	@echo "✅ CI pipeline complete!"

# Quick commands for testers
record: run-record ## Alias for run-record
playback: run-playback ## Alias for run-playback
clear: clean-recordings ## Alias for clean-recordings

# Default config file
config: ## Create default config file
	@echo "📝 Creating default config file..."
	@cat > proxy.yaml <<EOF
	server:
	  port: 8080
	  host: 0.0.0.0
	storage:
	  type: filesystem
	  path: ./recordings
	mode:
	  default: playback
	tls:
	  skip_verify: true
	EOF
	@echo "✅ Config file created: proxy.yaml"

# Show statistics
stats: ## Show recording statistics
	@echo "📊 Recording Statistics"
	@echo "======================"
	@echo "Total recordings: $$(find ./recordings -name "*.json" 2>/dev/null | wc -l)"
	@echo "By service:"
	@for dir in $$(find ./recordings -type d -maxdepth 1 2>/dev/null | tail -n +2); do \
		service=$$(basename $$dir); \
		count=$$(find $$dir -name "*.json" 2>/dev/null | wc -l); \
		echo "  $$service: $$count"; \
	done