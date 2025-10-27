.PHONY: build run test clean deps install

# Build the application
build:
	go build -o bin/javinizer-api cmd/api/main.go
	go build -o bin/javinizer ./cmd/cli

# Run the API server
run-api:
	go run cmd/api/main.go

# Run the CLI (for backwards compatibility)
cli:
	go run ./cmd/cli

# Run the CLI (primary target)
run:
	go run ./cmd/cli

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install the binaries
install:
	go build -o $(GOPATH)/bin/javinizer-api cmd/api/main.go
	go build -o $(GOPATH)/bin/javinizer ./cmd/cli

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Generate API documentation
docs:
	swag init -g cmd/api/main.go -o api/docs
