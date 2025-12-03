#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Project Health Check Script
# ============================================================================
# Checks project health: dependencies, linting, formatting, and structure
# Usage: ./scripts/health-check.sh [--fix]
# ============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Options
FIX=false

if [ "${1:-}" = "--fix" ]; then
  FIX=true
fi

echo "${BOLD}Project Health Check${NC}"
echo ""

HEALTH_SCORE=0
CHECKS=0

# Helper function to check and report
check_item() {
  local name="$1"
  local command="$2"
  local fix_command="${3:-}"
  
  ((CHECKS++))
  
  echo -n "  Checking $name... "
  
  if eval "$command" > /dev/null 2>&1; then
    echo "${GREEN}✅${NC}"
    ((HEALTH_SCORE++))
  else
    echo "${RED}❌${NC}"
    
    if [ "$FIX" = true ] && [ -n "$fix_command" ]; then
      echo "    ${YELLOW}Fixing...${NC}"
      eval "$fix_command"
      echo "    ${GREEN}Fixed${NC}"
      ((HEALTH_SCORE++))
    fi
  fi
}

# Go version
echo "${BLUE}🔍 Go Environment${NC}"
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "  Go version: $GO_VERSION"
echo ""

# Dependencies
echo "${BLUE}📦 Dependencies${NC}"
check_item "go.mod tidiness" "go mod verify"
check_item "go.sum consistency" "go mod tidy" "go mod tidy"
echo ""

# Code Quality
echo "${BLUE}✨ Code Quality${NC}"

check_item "gofmt" "gofmt -l . | grep -q '^$'" "go fmt ./..."

if command -v golangci-lint &> /dev/null; then
  check_item "golangci-lint" "golangci-lint run ./... --deadline=30s 2>/dev/null"
else
  echo "  ${YELLOW}⚠️  golangci-lint not installed${NC}"
fi

check_item "go vet" "go vet ./..."
echo ""

# Module Structure
echo "${BLUE}🏗️  Module Structure${NC}"

# Check if modules exist
MODULES=$(find internal/modules -maxdepth 1 -type d ! -name modules | sed 's|internal/modules/||')
echo "  Found modules: $(echo $MODULES | tr '\n' ', ' | sed 's/,$//')"

# Check module structure
MODULE_ISSUES=0
for module in $MODULES; do
  MODULE_PATH="internal/modules/$module"
  
  # Check required directories
  if [ ! -d "$MODULE_PATH/domain" ]; then
    echo "  ${RED}❌${NC} Missing domain/ in $module"
    ((MODULE_ISSUES++))
  fi
  
  if [ ! -d "$MODULE_PATH/handler" ]; then
    echo "  ${RED}❌${NC} Missing handler/ in $module"
    ((MODULE_ISSUES++))
  fi
  
  if [ ! -d "$MODULE_PATH/service" ]; then
    echo "  ${RED}❌${NC} Missing service/ in $module"
    ((MODULE_ISSUES++))
  fi
  
  if [ ! -d "$MODULE_PATH/repository" ]; then
    echo "  ${RED}❌${NC} Missing repository/ in $module"
    ((MODULE_ISSUES++))
  fi
done

if [ $MODULE_ISSUES -eq 0 ]; then
  echo "  ${GREEN}✅${NC} All modules have correct structure"
  ((HEALTH_SCORE++))
else
  echo "  ${RED}❌${NC} Found $MODULE_ISSUES structure issues"
fi
((CHECKS++))
echo ""

# Dependency Isolation
echo "${BLUE}🔗 Dependency Isolation${NC}"

if go run cmd/lint-deps/main.go > /dev/null 2>&1; then
  echo "  ${GREEN}✅${NC} No cross-module dependencies"
  ((HEALTH_SCORE++))
else
  echo "  ${RED}❌${NC} Cross-module dependency violations detected"
  if [ "$FIX" = false ]; then
    echo "    ${YELLOW}Run with --fix to see details${NC}"
  fi
fi
((CHECKS++))
echo ""

# Configuration Files
echo "${BLUE}⚙️  Configuration${NC}"

check_item "config.yaml exists" "test -f config/config.yaml"
check_item "featureflags.yaml exists" "test -f config/featureflags.yaml"
echo ""

# Summary
echo "${BOLD}=== Health Check Summary ===${NC}"
PERCENTAGE=$((HEALTH_SCORE * 100 / CHECKS))
echo "  Score: ${BOLD}$HEALTH_SCORE/$CHECKS${NC} ($PERCENTAGE%)"
echo ""

if [ $PERCENTAGE -eq 100 ]; then
  echo "${GREEN}✅ Project is healthy!${NC}"
  exit 0
elif [ $PERCENTAGE -ge 80 ]; then
  echo "${YELLOW}⚠️  Project needs attention${NC}"
  echo ""
  echo "Run: ${BOLD}./scripts/health-check.sh --fix${NC}"
  exit 0
else
  echo "${RED}❌ Project has issues${NC}"
  echo ""
  echo "Run: ${BOLD}./scripts/health-check.sh --fix${NC}"
  exit 1
fi
