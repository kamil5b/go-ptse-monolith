#!/bin/bash
# Generate protobuf code for Product module

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$REPO_ROOT/proto"
INTERNAL_PROTO_DIR="$REPO_ROOT/internal/proto"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install it first:"
    echo "  macOS: brew install protobuf"
    echo "  Linux: apt-get install protobuf-compiler"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

echo "Generating protobuf code..."

# Generate Product v1
protoc \
    --go_out="$REPO_ROOT" \
    --go-grpc_out="$REPO_ROOT" \
    --go_opt=module=github.com/kamil5b/go-ptse-monolith \
    --go-grpc_opt=module=github.com/kamil5b/go-ptse-monolith \
    -I"$PROTO_DIR" \
    "$PROTO_DIR/product/v1/product.proto"

echo "✓ Protobuf code generated successfully!"
echo ""
echo "Generated files:"
echo "  - internal/modules/product/proto/product.pb.go"
echo "  - internal/modules/product/proto/product_grpc.pb.go"
