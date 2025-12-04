# Protocol Buffer Definitions

This directory contains all protobuf protocol definitions for gRPC services in the go-modular-monolith project.

## Directory Structure

```
proto/
└── product/
    └── v1/
        └── product.proto          # Product service protobuf definition
```

## Generating Code

### Prerequisites

Before generating protobuf code, ensure you have installed:

1. **protoc compiler**:
   ```bash
   # macOS
   brew install protobuf
   
   # Linux (Debian/Ubuntu)
   apt-get install protobuf-compiler
   
   # Verify installation
   protoc --version
   ```

2. **Go protobuf plugins**:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

### Running the Generation Script

```bash
./scripts/generate-proto.sh
```

This script will:
- Verify protoc is installed
- Install missing Go protoc plugins if needed
- Generate Go code for all .proto files
- Output files to `internal/proto/`

### Manual Generation

If you need to manually regenerate protobuf files:

```bash
protoc \
    --go_out=internal/proto \
    --go-grpc_out=internal/proto \
    --go_opt=module=github.com/kamil5b/go-ptse-monolith \
    --go-grpc_opt=module=github.com/kamil5b/go-ptse-monolith \
    -I. \
    proto/product/v1/product.proto
```

## Product Service (v1)

### File
`product/v1/product.proto`

### Service Definition

The ProductService provides five RPC methods for CRUD operations:

- **Create** - Create a new product
- **Get** - Retrieve a product by ID
- **List** - List all products
- **Update** - Update an existing product
- **Delete** - Soft delete a product

### Message Types

#### Core Message
- **Product** - Represents a product entity with metadata

#### Request Messages
- **CreateProductRequest** - Request to create a new product
- **GetProductRequest** - Request to get a product by ID
- **UpdateProductRequest** - Request to update a product
- **DeleteProductRequest** - Request to delete a product

#### Response Messages
- **CreateProductResponse** - Response containing created product
- **GetProductResponse** - Response containing product
- **ListProductResponse** - Response containing list of products
- **UpdateProductResponse** - Response containing updated product

### Generated Files

```
internal/proto/product/v1/
├── product.pb.go          # Message definitions and serialization
└── product_grpc.pb.go     # Service interfaces and implementations
```

## Adding New Services

To add a new gRPC service to the project:

1. **Create the proto file**:
   ```bash
   mkdir -p proto/<module>/v1
   touch proto/<module>/v1/<module>.proto
   ```

2. **Define your service**:
   ```protobuf
   syntax = "proto3";
   
   package <module>.v1;
   
   option go_package = "github.com/kamil5b/go-ptse-monolith/internal/proto/<module>/v1;<modulev1>";
   
   service <Module>Service {
       rpc Create(<CreateRequest>) returns (<CreateResponse>);
       // ... other methods
   }
   ```

3. **Generate code**:
   ```bash
   ./scripts/generate-proto.sh
   ```

4. **Implement the service**:
   - Create handler in `internal/transports/grpc/<module>.handler.go`
   - Implement the generated `<Module>ServiceServer` interface
   - Register the service in the gRPC server bootstrap

## Best Practices

### Message Design

1. **Use Semantic Versioning**: Keep services versioned (v1, v2, etc.)
2. **Backward Compatibility**: Don't remove or reorder fields without versioning
3. **Optional Fields**: Use `optional` for fields that may not always be present
4. **Well-Known Types**: Use `google.protobuf.Timestamp` for dates, `google.protobuf.Empty` for no-op responses
5. **Field Numbering**: Reserve numbers for future additions

### Service Design

1. **Idempotency**: Design methods to be idempotent where possible
2. **Context Propagation**: Always include context for cancellation and timeouts
3. **Error Handling**: Use gRPC error codes appropriately
4. **Streaming**: Consider streaming for large collections

### Development Workflow

1. **Edit .proto files** with your service definitions
2. **Run generation script** to update Go code
3. **Implement handlers** in the transport layer
4. **Write tests** using mocked service implementations
5. **Update documentation** when adding new services

## Common Patterns

### Optional Fields in Requests

```protobuf
message UpdateProductRequest {
  string id = 1;
  optional string name = 2;
  optional string description = 3;
}
```

### Repeated Fields for Collections

```protobuf
message ListProductResponse {
  repeated Product products = 1;
}
```

### Well-Known Types

```protobuf
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

message Product {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
}

service ProductService {
  rpc List(google.protobuf.Empty) returns (ListResponse);
}
```

## Troubleshooting

### protoc: command not found
- Install protobuf compiler: `brew install protobuf` (macOS) or `apt-get install protobuf-compiler` (Linux)

### protoc-gen-go: program not found
- Install plugin: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

### protoc-gen-go-grpc: program not found
- Install plugin: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

### Generated files have import errors
- Ensure the `--go_opt=module=` flag matches your go.mod module name
- Verify proto files are in the correct directory structure

## References

- [Protocol Buffers Documentation](https://developers.google.com/protocol-buffers)
- [gRPC Go Getting Started](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers v3 Language Guide](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance-best-practices/)
