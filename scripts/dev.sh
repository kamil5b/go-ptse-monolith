#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Development Helper Script
# ============================================================================
# Provides quick access to common development commands
# Usage: ./scripts/dev.sh [command]
# ============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Function to print usage
print_usage() {
  cat << EOF
${BOLD}Go Modular Monolith - Development Helper${NC}

${BLUE}Usage:${NC}
  ./scripts/dev.sh [command] [options]

${BLUE}Commands:${NC}

  ${BOLD}run${NC}                 Run the application (with migrations)
  ${BOLD}server${NC}              Start server only (skip migrations)
  ${BOLD}test${NC} [pattern]      Run tests (optional: filter by pattern)
  ${BOLD}test:unit${NC}           Run unit tests only
  ${BOLD}test:cover${NC}          Run tests with coverage report
  ${BOLD}lint${NC}                Run linter (golangci-lint)
  ${BOLD}deps${NC}                Check module dependencies for violations
  ${BOLD}mocks${NC}               Generate mocks from interfaces
  ${BOLD}fmt${NC}                 Format code (gofmt)
  ${BOLD}vet${NC}                 Run go vet
  ${BOLD}mod:tidy${NC}            Tidy Go modules
  ${BOLD}mod:verify${NC}          Verify Go module integrity
  ${BOLD}db:up${NC}               Run SQL migrations up
  ${BOLD}db:down${NC}             Run SQL migrations down
  ${BOLD}db:mongo:up${NC}         Run MongoDB migrations up
  ${BOLD}mongo:shell${NC}         Connect to MongoDB shell
  ${BOLD}postgres:shell${NC}      Connect to PostgreSQL shell
  ${BOLD}clean${NC}               Remove generated files and caches
  ${BOLD}help${NC}                Show this help message

${BLUE}Examples:${NC}
  ./scripts/dev.sh run
  ./scripts/dev.sh test ./internal/modules/user
  ./scripts/dev.sh test:cover
  ./scripts/dev.sh lint
  ./scripts/dev.sh db:up
  ./scripts/dev.sh postgres:shell
EOF
}

# Function to run tests with coverage
run_coverage() {
  echo "${BLUE}📊 Running tests with coverage...${NC}"
  go test -v -race -coverprofile=coverage.out ./...
  echo "${BLUE}📊 Generating coverage report...${NC}"
  go tool cover -html=coverage.out -o coverage.html
  echo "${GREEN}✅ Coverage report generated: coverage.html${NC}"
  
  # Also show coverage summary
  go tool cover -func=coverage.out | tail -1
}

# Function to check if command exists
command_exists() {
  command -v "$1" &> /dev/null
}

# Main command handling
case "${1:-help}" in
  run)
    echo "${BLUE}🚀 Running application with migrations...${NC}"
    go run .
    ;;
  
  server)
    echo "${BLUE}🚀 Starting server (skip migrations)...${NC}"
    go run . server
    ;;
  
  test)
    PATTERN="${2:-.}"
    echo "${BLUE}🧪 Running tests (pattern: $PATTERN)...${NC}"
    go test -v -race -run "$PATTERN" ./...
    ;;
  
  test:unit)
    echo "${BLUE}🧪 Running unit tests...${NC}"
    go test -v -short ./internal/...
    ;;
  
  test:cover)
    run_coverage
    ;;
  
  lint)
    echo "${BLUE}✨ Running linter...${NC}"
    if command_exists golangci-lint; then
      golangci-lint run ./...
      echo "${GREEN}✅ Linter checks passed${NC}"
    else
      echo "${RED}❌ golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
      exit 1
    fi
    ;;
  
  deps)
    echo "${BLUE}🔗 Checking module dependencies...${NC}"
    go run cmd/lint-deps/main.go
    ;;
  
  mocks)
    echo "${BLUE}🎭 Generating mocks...${NC}"
    if [ -f "scripts/generate_mocks_from_source.sh" ]; then
      bash scripts/generate_mocks_from_source.sh
      echo "${GREEN}✅ Mocks generated${NC}"
    else
      echo "${RED}❌ generate_mocks_from_source.sh not found${NC}"
      exit 1
    fi
    ;;
  
  fmt)
    echo "${BLUE}🎨 Formatting code...${NC}"
    go fmt ./...
    echo "${GREEN}✅ Code formatted${NC}"
    ;;
  
  vet)
    echo "${BLUE}🔍 Running go vet...${NC}"
    go vet ./...
    echo "${GREEN}✅ Go vet passed${NC}"
    ;;
  
  mod:tidy)
    echo "${BLUE}📦 Tidying modules...${NC}"
    go mod tidy
    echo "${GREEN}✅ Modules tidied${NC}"
    ;;
  
  mod:verify)
    echo "${BLUE}📦 Verifying modules...${NC}"
    go mod verify
    echo "${GREEN}✅ Modules verified${NC}"
    ;;
  
  db:up)
    echo "${BLUE}🗄️  Running SQL migrations up...${NC}"
    go run . migration sql up
    ;;
  
  db:down)
    echo "${BLUE}🗄️  Running SQL migrations down...${NC}"
    go run . migration sql down
    ;;
  
  db:mongo:up)
    echo "${BLUE}🗄️  Running MongoDB migrations up...${NC}"
    go run . migration mongo up
    ;;
  
  mongo:shell)
    echo "${BLUE}🗄️  Connecting to MongoDB...${NC}"
    if command_exists mongosh; then
      mongosh "mongodb://localhost:27017"
    else
      echo "${RED}❌ mongosh not found. Install MongoDB tools or use Docker.${NC}"
      exit 1
    fi
    ;;
  
  postgres:shell)
    echo "${BLUE}🗄️  Connecting to PostgreSQL...${NC}"
    if command_exists psql; then
      psql -h localhost -U postgres
    else
      echo "${RED}❌ psql not found. Install PostgreSQL client.${NC}"
      exit 1
    fi
    ;;
  
  clean)
    echo "${BLUE}🧹 Cleaning up...${NC}"
    rm -f coverage.out coverage.html
    go clean -testcache
    find . -name "*.test" -delete
    echo "${GREEN}✅ Cleanup complete${NC}"
    ;;
  
  help|--help|-h)
    print_usage
    ;;
  
  *)
    echo "${RED}❌ Unknown command: $1${NC}"
    echo ""
    print_usage
    exit 1
    ;;
esac
