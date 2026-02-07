.PHONY: help build run test clean docker-build docker-run

# Variables
APP_NAME=core
MAIN_PATH=./cmd/app
BUILD_DIR=./build/bin
DOCKER_IMAGE=$(APP_NAME):latest

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/app/main.go -o docs --parseInternal --parseDependency
	@echo "Swagger docs generated in ./docs"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) -f build/package/Dockerfile .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(DOCKER_IMAGE)
