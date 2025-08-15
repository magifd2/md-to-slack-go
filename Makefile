.PHONY: build test clean lint install-lint

# Build the application
build:
	go build -o md-to-slack ./cmd/md-to-slack

# Run tests
test:
	go test ./...

# Clean up build artifacts
clean:
	rm -f md-to-slack

# Lint the code
# You may need to install golangci-lint first: make install-lint
lint:
	golangci-lint run

# Install golangci-lint
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
