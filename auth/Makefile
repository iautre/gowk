# Makefile for auth protobuf and sqlc code generation

.PHONY: proto sqlc generate clean build help

# Default target
help:
	@echo "Available targets:"
	@echo "  proto      - Generate protobuf Go files"
	@echo "  sqlc       - Generate SQL code from queries"
	@echo "  generate   - Generate both proto and sqlc files"
	@echo "  clean      - Remove generated files"
	@echo "  build      - Build the combined server"
	@echo "  help       - Show this help message"

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@mkdir -p proto
	PATH=$$PATH:$(shell go env GOPATH)/bin protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/auth.proto
	@echo "✅ Protobuf generation completed!"
	@echo "Generated files: proto/auth.pb.go, proto/auth_grpc.pb.go"

# Generate SQL code from queries
sqlc:
	@echo "Generating SQL code..."
	@if ! command -v sqlc >/dev/null 2>&1; then \
		echo "❌ sqlc not found. Install with: go install github.com/kyleconroy/sqlc/cmd/sqlc@latest"; \
		exit 1; \
	fi
	sqlc generate
	@echo "✅ SQL code generation completed!"
	@echo "Generated files in db/ directory"

# Generate both proto and sqlc files
generate: proto sqlc
	@echo "✅ All code generation completed!"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -f proto/auth.pb.go proto/auth_grpc.pb.go
	@rm -rf db/*.go
	@echo "✅ Clean completed!"

# Install required tools (one-time setup)
install-tools:
	@echo "Installing required tools..."
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Installing sqlc..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "✅ All tools installed! Make sure ~/.protoc/bin is in your PATH."

# Check if tools are installed
check-tools:
	@echo "Checking installed tools..."
	@command -v protoc >/dev/null 2>&1 && echo "✅ protoc found" || echo "❌ protoc not found"
	@command -v protoc-gen-go >/dev/null 2>&1 && echo "✅ protoc-gen-go found" || echo "❌ protoc-gen-go not found"
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 && echo "✅ protoc-gen-go-grpc found" || echo "❌ protoc-gen-go-grpc not found"
	@command -v sqlc >/dev/null 2>&1 && echo "✅ sqlc found" || echo "❌ sqlc not found"

# Build combined server
build:
	@echo "Building combined HTTP/gRPC server..."
	go build -o bin/server ./cmd
	@echo "✅ Build completed! Binary: bin/server"

# Run combined server
run: build
	@echo "Starting combined HTTP/gRPC server..."
	./bin/server -http-port=8087 -grpc-port=50051
