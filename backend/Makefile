.PHONY: build run test clean dev deps lint

# Binary name
BINARY_NAME=feats-api

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api

# Run the application
run: build
	./$(BINARY_NAME)

# Run in development mode (without building)
dev:
	GIN_MODE=debug $(GORUN) ./cmd/api

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Lint the code (requires golangci-lint)
lint:
	golangci-lint run

# Check for security vulnerabilities
security:
	govulncheck ./...

# Create initial admin user (interactive)
create-admin:
	@echo "Creating admin user..."
	$(GORUN) ./cmd/admin/main.go

# Database operations
db-reset:
	rm -f feats.db
	@echo "Database reset. Run 'make dev' to recreate."

# Generate a secure JWT secret
generate-secret:
	@openssl rand -base64 32

# All-in-one setup
setup: deps
	@echo "Setup complete. Copy .env.example to .env and configure your settings."
	@echo "Then run 'make dev' to start the server."
