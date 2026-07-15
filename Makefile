.PHONY: build run run-dev test test-coverage test-all benchmark mock clean deps lint fmt vet config-validate config-show version docker-build docker-run docker-compose-up docker-compose-down docker-compose-logs docker-compose-restart clean-all swagger swagger-clean help

# Variables
APP_NAME := orange-tv
VERSION := 1.0.0
BUILD_DIR := build
MAIN_PATH := ./main.go

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-X github.com/ilaziness/orange-tv/cmd.version=$(VERSION)"

# Fail if $(1) is not in PATH; $(2) = suggested install command
define require_tool
	@command -v $(1) >/dev/null 2>&1 || { \
		printf 'Error: %s not found in PATH.\nInstall with:\n  %s\n' '$(1)' '$(2)'; \
		exit 1; \
	}
endef

# Default target
all: build

## build: Build the application
build:
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

## run: Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run $(MAIN_PATH) serve

## run-dev: Run with development config
run-dev:
	@echo "Running $(APP_NAME) in development mode..."
	$(GOCMD) run $(MAIN_PATH) serve -c configs/config.dev.yaml

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -func=coverage.out
	@echo "Coverage report: coverage.html"

## test-all: Run all tests with coverage and race detection
test-all:
	@echo "Running all tests with coverage and race detection..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -func=coverage.out

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## mock: Generate mock files using mockgen (add sources when services exist)
mock:
	$(call require_tool,mockgen,go install go.uber.org/mock/mockgen@latest)
	@echo "No mock sources configured (user example removed). Add mockgen lines when services exist."

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## lint: Run linter
lint:
	$(call require_tool,golangci-lint,go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2)
	@echo "Running linter..."
	golangci-lint run --timeout=5m ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## config-validate: Validate configuration
config-validate:
	@$(GOCMD) run $(MAIN_PATH) config validate

## config-show: Show current configuration
config-show:
	@$(GOCMD) run $(MAIN_PATH) config show

## version: Show version
version:
	@echo "$(APP_NAME) version $(VERSION)"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

## docker-compose-up: Start services with docker-compose
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

## docker-compose-down: Stop services with docker-compose
docker-compose-down:
	@echo "Stopping services..."
	docker-compose down

## docker-compose-logs: View logs
docker-compose-logs:
	docker-compose logs -f

## docker-compose-restart: Restart services
docker-compose-restart:
	@echo "Restarting services..."
	docker-compose restart

## clean-all: Clean all artifacts including Docker
clean-all: clean
	@echo "Cleaning all artifacts..."
	@docker system prune -f

## swagger: Generate Swagger documentation
swagger:
	$(call require_tool,swag,go install github.com/swaggo/swag/cmd/swag@latest)
	@echo "Generating Swagger documentation..."
	swag init -g main.go -o docs/swagger
	@echo "Swagger documentation generated at docs/swagger/"

## swagger-clean: Clean generated Swagger documentation
swagger-clean:
	@echo "Cleaning Swagger documentation..."
	@rm -rf docs/swagger/docs.go docs/swagger/swagger.json docs/swagger/swagger.yaml
	@echo "Swagger documentation cleaned"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build the application"
	@echo "  run                Run the application"
	@echo "  run-dev            Run with development config"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage"
	@echo "  test-all           Run all tests with coverage and race detection"
	@echo "  benchmark          Run benchmarks"
	@echo "  mock               Generate mock files"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Download dependencies"
	@echo "  lint               Run linter"
	@echo "  fmt                Format code"
	@echo "  vet                Run go vet"
	@echo "  config-validate    Validate configuration"
	@echo "  config-show        Show current configuration"
	@echo "  version            Show version"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-run         Run Docker container"
	@echo "  docker-compose-up  Start services with docker-compose"
	@echo "  docker-compose-down  Stop services"
	@echo "  docker-compose-logs  View logs"
	@echo "  docker-compose-restart  Restart services"
	@echo "  clean-all          Clean all artifacts including Docker"
	@echo "  swagger            Generate Swagger documentation"
	@echo "  swagger-clean      Clean generated Swagger documentation"
	@echo "  help               Show this help message"
