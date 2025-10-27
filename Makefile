.PHONY: build run run-api test clean deps install web-dev web-build web-preview web-install web-clean

# Build the application (single binary)
build:
	go build -o bin/javinizer ./cmd/cli

# Run the CLI (primary target)
run:
	go run ./cmd/cli

# Run the API server using subcommand
run-api:
	go run ./cmd/cli api

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

# Install the binary
install:
	go build -o $(GOPATH)/bin/javinizer ./cmd/cli

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Generate API documentation
docs:
	swag init -g cmd/cli/api.go -o api/docs

# Web frontend targets
web-dev:
	cd web/frontend && npm run dev

web-build:
	cd web/frontend && npm run build

web-preview:
	cd web/frontend && npm run preview

web-install:
	cd web/frontend && npm install

web-clean:
	rm -rf web/frontend/node_modules web/frontend/.svelte-kit web/frontend/build
