.PHONY: run test lint migrate deps-check

# Run the application
run:
    go run .

# Run server only (skip migrations)
server:
    go run . server

# Run all tests
test:
    go test -v -race -cover ./...

# Run unit tests only
test-unit:
    go test -v -short ./...

# Run linter
lint:
    golangci-lint run ./...

# Check module dependencies
deps-check:
    go run cmd/lint-deps/main.go

# Run SQL migrations
migrate-up:
    go run . migration sql up

migrate-down:
    go run . migration sql down

# Run MongoDB migrations
migrate-mongo:
    go run . migration mongo up

# Generate mocks (requires mockery)
mocks:
    mockery --all --output=internal/mocks --outpkg=mocks

# Build the application
build:
    go build -o bin/app .

# Clean build artifacts
clean:
    rm -rf bin/

# All checks before commit
pre-commit: deps-check lint test