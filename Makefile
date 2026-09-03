.PHONY: all build run dev test test-unit test-integration lint fmt vet docker-up docker-down docker-logs migrate-up migrate-down generate-secrets coverage clean help

APP_NAME := dbplatform
BIN_DIR  := bin
CMD_PATH := ./cmd/api
GO_FILES := $(shell find . -name '*.go' -not -path '*/vendor/*')
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")

all: fmt vet test build

build: ## Build the binary
	@echo ">>> Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.version=$(VERSION) -w -s" -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo ">>> Built: $(BIN_DIR)/$(APP_NAME)"

run: build ## Run the binary
	./$(BIN_DIR)/$(APP_NAME)

dev: ## Run in development mode (with .env)
	@echo ">>> Starting in development mode..."
	APP_ENV=development go run $(CMD_PATH)

test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only
	go test -v -race -short ./internal/...

test-integration: ## Run integration tests (requires Docker)
	go test -v -race -timeout 120s -run TestIntegration ./tests/integration/...

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format Go code
	gofmt -w $(GO_FILES)

vet: ## Run go vet
	go vet ./...

docker-up: ## Start all services with Docker Compose
	docker compose up -d
	@echo ">>> Services started. API: http://localhost:8080"
	@echo ">>> Grafana:     http://localhost:3001 (admin/admin123)"
	@echo ">>> Prometheus:  http://localhost:9090"

docker-down: ## Stop all Docker services
	docker compose down

docker-logs: ## Follow logs for api service
	docker compose logs -f api

docker-build: ## Build Docker images
	docker compose build

migrate-up: ## Run database migrations up
	go run ./cmd/migrate up

migrate-down: ## Roll back last migration
	go run ./cmd/migrate down 1

migrate-version: ## Show current migration version
	go run ./cmd/migrate version

generate-secrets: ## Generate JWT_SECRET and ENCRYPTION_KEY
	@echo "Add these to your .env file:"
	@printf "JWT_SECRET=%s\n" "$$(openssl rand -base64 48 2>/dev/null || head -c 48 /dev/urandom | base64)"
	@printf "ENCRYPTION_KEY=%s\n" "$$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)"

coverage: test ## Generate coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo ">>> Coverage report: coverage.html"

clean: ## Clean build artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
