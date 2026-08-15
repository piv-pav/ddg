# Build ddg CLI
build:
    export VERSION=$(cat VERSION 2>/dev/null || echo "dev") && \
    go build -o bin/ddg -ldflags "-X main.version=v$VERSION-dev" .

# Install ddg CLI to ~/go/bin
install: build
    mkdir -p ~/go/bin
    cp bin/ddg ~/go/bin/ddg
    chmod +x ~/go/bin/ddg
    echo "Installed to ~/go/bin/ddg"

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
