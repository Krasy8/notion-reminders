.PHONY: build install clean test run

# Binary name
BINARY=notion-reminder

# Installation directory
INSTALL_DIR=$(HOME)/.local/bin

# Build the binary
build:
	@echo "Building $(BINARY)..."
	@go build -o $(BINARY) main.go
	@echo "✓ Build complete: $(BINARY)"

# Build with optimizations (smaller binary)
build-optimized:
	@echo "Building optimized $(BINARY)..."
	@go build -ldflags="-s -w" -o $(BINARY) main.go
	@echo "✓ Optimized build complete"
	@ls -lh $(BINARY)

# Install to ~/.local/bin
install: build-optimized
	@echo "Installing to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY)
	@echo "✓ Clean complete"

# Run without installing
run: build
	@./$(BINARY)

# Run tests (if you add tests later)
test:
	@go test -v ./...

# Format code
fmt:
	@go fmt ./...

# Check code
vet:
	@go vet ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  build-optimized  - Build with size optimizations"
	@echo "  install          - Build and install to ~/.local/bin"
	@echo "  run              - Build and run immediately"
	@echo "  clean            - Remove build artifacts"
	@echo "  fmt              - Format Go code"
	@echo "  vet              - Run Go vet"
	@echo "  test             - Run tests"
	@echo "  help             - Show this help"
