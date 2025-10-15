# Makefile for Chargeback API

.PHONY: test test-coverage test-internal test-unit test-integration clean build help

# Build configuration
APP_NAME=chargeback-api
BUILD_DIR=bin
COVERAGE_DIR=coverage

# Go test configuration
INTERNAL_PACKAGES=./internal/...
UNIT_PACKAGES=./internal/domain/... ./internal/usecase/...
INTEGRATION_PACKAGES=./internal/infra/... ./internal/api/... ./internal/server/...

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the application
	@echo "🔨 Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

clean: ## Clean build artifacts and coverage reports
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR) *.out *.html
	@echo "✅ Clean complete"

test: ## Run all tests
	@echo "🧪 Running all tests..."
	@go test -v ./...

test-internal: ## Run tests only for internal packages (excluding examples)
	@echo "🧪 Running internal tests..."
	@go test -v $(INTERNAL_PACKAGES)

test-unit: ## Run unit tests (domain + usecase)
	@echo "🧪 Running unit tests..."
	@go test -v $(UNIT_PACKAGES)

test-integration: ## Run integration tests (infra + api + server)
	@echo "🧪 Running integration tests..."
	@go test -v $(INTEGRATION_PACKAGES)

test-coverage: ## Generate coverage report for internal packages only
	@echo "📊 Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out $(INTERNAL_PACKAGES)
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "📈 Coverage report: $(COVERAGE_DIR)/coverage.html"

test-coverage-summary: ## Show coverage summary for internal packages
	@echo "📊 Coverage Summary (Internal Packages Only):"
	@echo "=============================================="
	@go test -cover $(INTERNAL_PACKAGES) 2>/dev/null | grep -E "(coverage:|ok)" | sort

# Exclude directories that don't need tests
test-focus: ## Run tests excluding examples, docs, and build artifacts
	@echo "🎯 Running focused tests (excluding examples, docs, build artifacts)..."
	@go test -v $(INTERNAL_PACKAGES) ./cmd/api

run: build ## Build and run the application
	@echo "🚀 Starting $(APP_NAME)..."
	@./$(BUILD_DIR)/$(APP_NAME)

# Development helpers
fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@echo "🔍 Running linter..."
	@golangci-lint run

mod-tidy: ## Tidy go modules
	@echo "📦 Tidying modules..."
	@go mod tidy

dev-setup: mod-tidy fmt ## Setup development environment
	@echo "🛠️  Development setup complete"

# CI/CD helpers
ci-test: test-coverage ## Run tests for CI/CD (with coverage)
	@echo "🏗️  CI tests complete"

# Docker helpers (if needed later)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(APP_NAME) .

# Coverage thresholds
coverage-check: test-coverage ## Check if coverage meets minimum thresholds
	@echo "🎯 Checking coverage thresholds..."
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | grep "total:" | awk '{if ($$3+0 < 70) {print "❌ Coverage " $$3 " below 70% threshold"; exit 1} else {print "✅ Coverage " $$3 " meets threshold"}}'