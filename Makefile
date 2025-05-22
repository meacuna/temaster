.PHONY: build test lint clean

# Build the application
build:
	go build -o temaster ./cmd

# Run tests
test:
	go test -v ./...

# Run linters
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f temaster
	rm -f *.test
	rm -f *.out

# Install development tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 