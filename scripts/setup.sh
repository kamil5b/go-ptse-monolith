#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Project Setup Script
# ============================================================================
# Initializes the project with all necessary dependencies and tools
# Usage: ./scripts/setup.sh
# ============================================================================

echo "🚀 Setting up Go Modular Monolith project..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check Go version
echo "${BLUE}📦 Checking Go version...${NC}"
GO_VERSION=$(go version | awk '{print $3}')
echo "   Go version: $GO_VERSION"
echo ""

# Download Go modules
echo "${BLUE}📦 Downloading Go modules...${NC}"
go mod download
go mod tidy
echo "   ✅ Modules downloaded and tidied"
echo ""

# Install development tools
echo "${BLUE}🛠️  Installing development tools...${NC}"

tools=(
  "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  "github.com/golang/mock/mockgen@latest"
  "goose.db/cmd/goose@latest"
)

for tool in "${tools[@]}"; do
  echo "   Installing $tool..."
  go install "$tool" 2>/dev/null || echo "   ⚠️  Could not install $tool"
done
echo "   ✅ Development tools installed"
echo ""

# Create necessary directories
echo "${BLUE}📁 Creating necessary directories...${NC}"
mkdir -p internal/infrastructure/db/sql/migration
mkdir -p internal/infrastructure/db/mongo/migration
mkdir -p config
echo "   ✅ Directories created"
echo ""

# Check database configuration
echo "${BLUE}🗄️  Checking database configuration...${NC}"
if [ -f "config/config.yaml" ]; then
  echo "   ✅ config/config.yaml exists"
else
  echo "   ⚠️  config/config.yaml not found"
fi
echo ""

# Generate mocks
echo "${BLUE}🎭 Generating mocks...${NC}"
if command -v mockgen &> /dev/null; then
  bash ./scripts/generate_mocks_from_source.sh
  echo "   ✅ Mocks generated"
else
  echo "   ⚠️  mockgen not found, skipping mock generation"
fi
echo ""

# Run linter checks
echo "${BLUE}✨ Running linter checks...${NC}"
if command -v golangci-lint &> /dev/null; then
  golangci-lint run --deadline=5m ./internal/... ./cmd/... 2>/dev/null || echo "   ⚠️  Some lint warnings found"
  echo "   ✅ Linter checks completed"
else
  echo "   ⚠️  golangci-lint not found, skipping linter"
fi
echo ""

# Check module dependencies
echo "${BLUE}🔗 Checking module dependencies...${NC}"
if go run cmd/lint-deps/main.go; then
  echo "   ✅ No cross-module dependency violations"
else
  echo "   ❌ Cross-module dependency violations found"
fi
echo ""

echo "${GREEN}✅ Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Update config/config.yaml with your database credentials"
echo "  2. Run: go run . migration sql up    (apply SQL migrations)"
echo "  3. Run: go run .                     (start the server)"
echo ""
echo "Useful commands:"
echo "  • make run                (run with migrations)"
echo "  • make server             (run server only)"
echo "  • make test               (run all tests)"
echo "  • make lint               (run linter)"
echo "  • make deps-check         (check module dependencies)"
echo "  • ./scripts/dev.sh        (development helper)"
echo "  • ./scripts/new-module.sh (scaffold new module)"
echo ""
