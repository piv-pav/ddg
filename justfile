# Build ddg CLI
build:
    export VERSION=$(cat VERSION 2>/dev/null || echo "dev") && \
    go build -o bin/ddg -ldflags "-X main.version=$VERSION" .

# Install ddg CLI to ~/.local/bin
install: build
    mkdir -p ~/.local/bin
    cp bin/ddg ~/.local/bin/ddg
    chmod +x ~/.local/bin/ddg
    echo "Installed to ~/.local/bin/ddg (ensure ~/.local/bin is in PATH)"

# Run tests
test:
    go test -v ./...

# Clean build artifacts
clean:
    rm -rf bin/

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run
