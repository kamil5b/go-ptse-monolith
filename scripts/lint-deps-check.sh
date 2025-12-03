#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Module Dependency Linter with Enhanced Output
# ============================================================================
# Checks for cross-module dependencies and provides detailed violation reports
# Usage: ./scripts/lint-deps-check.sh [--verbose] [--fix]
# ============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Options
VERBOSE=false
FIX=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --verbose)
      VERBOSE=true
      shift
      ;;
    --fix)
      FIX=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

echo "${BOLD}Module Dependency Linter${NC}"
echo "${BLUE}Analyzing cross-module dependencies...${NC}"
echo ""

# Get all modules
MODULES=$(find internal/modules -maxdepth 1 -type d ! -name modules | sed 's|internal/modules/||' | sort)

# Track violations
VIOLATIONS=0
CHECKED_FILES=0

for module in $MODULES; do
  MODULE_PATH="internal/modules/$module"
  
  if [ $VERBOSE = true ]; then
    echo "${BLUE}Checking module: $module${NC}"
  fi
  
  # Find all .go files in the module (excluding mocks and tests)
  while IFS= read -r file; do
    ((CHECKED_FILES++))
    
    # Extract imports from the file
    imports=$(grep -E "^\s*\"go-modular-monolith/internal/modules" "$file" || true)
    
    if [ -z "$imports" ]; then
      [ $VERBOSE = true ] && echo "  ✅ $file"
      continue
    fi
    
    # Check each import
    while IFS= read -r import_line; do
      # Extract module name from import
      imported_module=$(echo "$import_line" | grep -oE 'go-modular-monolith/internal/modules/[^/]+' | sed 's|.*modules/||')
      
      # Skip if import is from same module or from ACL (ACL is allowed to import other modules)
      if [ "$imported_module" = "$module" ] || [[ "$file" == *"/acl/"* ]]; then
        continue
      fi
      
      # This is a violation!
      ((VIOLATIONS++))
      
      echo "${RED}❌ Cross-module dependency violation:${NC}"
      echo "   ${BOLD}File:${NC} $file"
      echo "   ${BOLD}Module:${NC} $module"
      echo "   ${BOLD}Imports from:${NC} $imported_module"
      echo "   ${BOLD}Line:${NC} $import_line"
      echo ""
      
    done <<< "$imports"
    
  done < <(find "$MODULE_PATH" -type f -name "*.go" \
      ! -path "*mocks*" ! -name "*_test.go" \
      ! -path "*noop*" 2>/dev/null)
done

echo "${BLUE}Checking for shared kernel imports...${NC}"

# Verify that all module files only import from shared (cross-module safe)
SHARED_VIOLATIONS=0
for module in $MODULES; do
  MODULE_PATH="internal/modules/$module"
  
  while IFS= read -r file; do
    # ACL files are allowed to import other modules
    if [[ "$file" == *"/acl/"* ]]; then
      continue
    fi
    
    # Check for imports NOT from shared kernel
    bad_imports=$(grep -E "^\s*\"go-modular-monolith/internal/(app|infrastructure)" "$file" || true)
    
    if [ -n "$bad_imports" ]; then
      if [ $VIOLATIONS -eq 0 ]; then
        echo ""
      fi
      
      echo "${RED}⚠️  Unexpected infrastructure import:${NC}"
      echo "   ${BOLD}File:${NC} $file"
      echo "   $bad_imports"
      ((SHARED_VIOLATIONS++))
    fi
    
  done < <(find "$MODULE_PATH" -type f -name "*.go" \
      ! -path "*mocks*" ! -name "*_test.go" 2>/dev/null)
done

echo ""
echo "${BOLD}=== Summary ===${NC}"
echo "Files checked:      $CHECKED_FILES"
echo "Violations found:   $VIOLATIONS"

if [ $VIOLATIONS -eq 0 ]; then
  echo ""
  echo "${GREEN}✅ No cross-module dependency violations found!${NC}"
  echo ""
  echo "${BLUE}Module Structure:${NC}"
  for module in $MODULES; do
    echo "  • $module"
  done
  exit 0
else
  echo ""
  echo "${RED}❌ Cross-module dependency violations detected!${NC}"
  echo ""
  echo "${BLUE}Violation Resolution Tips:${NC}"
  echo "  1. Use Anti-Corruption Layer (ACL) for synchronous operations"
  echo "  2. Use Event Bus for asynchronous operations"
  echo "  3. Define interface in your module's domain package"
  echo "  4. Create adapter in your module's acl/ folder"
  echo "  5. Inject the adapter via the Container"
  echo ""
  echo "${YELLOW}See TECHNICAL_DOCUMENTATION.md for ACL pattern details${NC}"
  exit 1
fi
