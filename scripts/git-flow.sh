#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Git Workflow Helper Script
# ============================================================================
# Automates common git workflow tasks for feature development
# Usage: ./scripts/git-flow.sh [command]
# ============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_usage() {
  cat << EOF
${BOLD}Git Workflow Helper${NC}

${BLUE}Usage:${NC}
  ./scripts/git-flow.sh [command]

${BLUE}Commands:${NC}
  ${BOLD}start${NC} <feature-name>       Start a new feature branch
  ${BOLD}status${NC}                      Show branch status and diff
  ${BOLD}pre-commit${NC}                  Run pre-commit checks
  ${BOLD}push${NC}                        Push current branch
  ${BOLD}publish${NC}                     Push with upstream tracking
  ${BOLD}rebase${NC}                      Rebase on main branch
  ${BOLD}sync${NC}                        Fetch and rebase on main
  ${BOLD}diff${NC}                        Show diff with main
  ${BOLD}commits${NC}                     Show commits on this branch
  ${BOLD}merge${NC}                       Merge to main and cleanup
  ${BOLD}cleanup${NC}                     Delete local/remote branches

${BLUE}Examples:${NC}
  ./scripts/git-flow.sh start add-payment-module
  ./scripts/git-flow.sh pre-commit
  ./scripts/git-flow.sh publish
  ./scripts/git-flow.sh sync
  ./scripts/git-flow.sh merge
EOF
}

# Get current branch
get_current_branch() {
  git rev-parse --abbrev-ref HEAD
}

# Check if on main branch
check_not_main() {
  if [ "$(get_current_branch)" = "main" ]; then
    echo "${RED}❌ You are on the main branch${NC}"
    exit 1
  fi
}

# Main command handling
case "${1:-help}" in
  start)
    if [ -z "${2:-}" ]; then
      echo "${RED}❌ Usage: ./scripts/git-flow.sh start <feature-name>${NC}"
      exit 1
    fi
    
    FEATURE_NAME="$2"
    
    # Validate branch name
    if ! [[ "$FEATURE_NAME" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$ ]]; then
      echo "${RED}❌ Invalid branch name: $FEATURE_NAME${NC}"
      echo "   Use lowercase, hyphens, and numbers (e.g., 'add-payment-module')"
      exit 1
    fi
    
    echo "${BLUE}🌿 Starting feature: $FEATURE_NAME${NC}"
    git fetch origin
    git checkout -b "feature/$FEATURE_NAME" origin/main
    echo "${GREEN}✅ Feature branch created: feature/$FEATURE_NAME${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Implement your feature"
    echo "  2. Run: ./scripts/git-flow.sh pre-commit"
    echo "  3. Run: ./scripts/git-flow.sh publish"
    echo "  4. Open a Pull Request on GitHub"
    ;;
  
  status)
    CURRENT_BRANCH=$(get_current_branch)
    echo "${BLUE}Current branch: ${BOLD}$CURRENT_BRANCH${NC}"
    echo ""
    
    # Show branch info
    git status
    echo ""
    
    # Show commits ahead of main
    COMMIT_COUNT=$(git rev-list --count main.."$CURRENT_BRANCH" 2>/dev/null || echo "0")
    if [ "$COMMIT_COUNT" -gt 0 ]; then
      echo "${YELLOW}📊 Commits ahead of main: $COMMIT_COUNT${NC}"
    fi
    ;;
  
  pre-commit)
    check_not_main
    
    echo "${BLUE}🔍 Running pre-commit checks...${NC}"
    echo ""
    
    # Run health check
    echo "${BLUE}📊 Running health check...${NC}"
    if ./scripts/health-check.sh; then
      echo "${GREEN}✅ Health check passed${NC}"
    else
      echo "${RED}❌ Health check failed${NC}"
      exit 1
    fi
    echo ""
    
    # Run tests
    echo "${BLUE}🧪 Running tests...${NC}"
    if go test -v -race ./...; then
      echo "${GREEN}✅ Tests passed${NC}"
    else
      echo "${RED}❌ Tests failed${NC}"
      exit 1
    fi
    echo ""
    
    # Check dependencies
    echo "${BLUE}🔗 Checking dependencies...${NC}"
    if ./scripts/lint-deps-check.sh; then
      echo "${GREEN}✅ Dependency check passed${NC}"
    else
      echo "${RED}❌ Dependency violations found${NC}"
      exit 1
    fi
    echo ""
    
    echo "${GREEN}✅ All pre-commit checks passed!${NC}"
    ;;
  
  push)
    check_not_main
    
    CURRENT_BRANCH=$(get_current_branch)
    echo "${BLUE}📤 Pushing branch: $CURRENT_BRANCH${NC}"
    git push origin "$CURRENT_BRANCH"
    echo "${GREEN}✅ Branch pushed${NC}"
    ;;
  
  publish)
    check_not_main
    
    CURRENT_BRANCH=$(get_current_branch)
    echo "${BLUE}📤 Publishing branch: $CURRENT_BRANCH${NC}"
    git push -u origin "$CURRENT_BRANCH"
    echo "${GREEN}✅ Branch published${NC}"
    echo ""
    echo "Create a Pull Request at:"
    echo "  https://github.com/$(git config --get remote.origin.url | sed 's/.*:\(.*\)\/\(.*\)\.git/\1\/\2/')/pull/new/$CURRENT_BRANCH"
    ;;
  
  rebase)
    check_not_main
    
    echo "${BLUE}🔄 Rebasing on main...${NC}"
    git fetch origin
    git rebase origin/main
    echo "${GREEN}✅ Rebased on main${NC}"
    ;;
  
  sync)
    check_not_main
    
    echo "${BLUE}🔄 Syncing with main...${NC}"
    git fetch origin
    git rebase origin/main
    echo "${GREEN}✅ Synced with main${NC}"
    
    # Offer to push
    read -p "Push updated branch? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      CURRENT_BRANCH=$(get_current_branch)
      git push --force-with-lease origin "$CURRENT_BRANCH"
      echo "${GREEN}✅ Branch pushed${NC}"
    fi
    ;;
  
  diff)
    CURRENT_BRANCH=$(get_current_branch)
    if [ "$CURRENT_BRANCH" = "main" ]; then
      echo "${YELLOW}⚠️  You are on main branch${NC}"
      exit 0
    fi
    
    echo "${BLUE}📊 Changes vs main:${NC}"
    git diff main..."$CURRENT_BRANCH" --stat
    echo ""
    read -p "Show full diff? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      git diff main..."$CURRENT_BRANCH"
    fi
    ;;
  
  commits)
    CURRENT_BRANCH=$(get_current_branch)
    if [ "$CURRENT_BRANCH" = "main" ]; then
      echo "${YELLOW}⚠️  You are on main branch${NC}"
      exit 0
    fi
    
    echo "${BLUE}📋 Commits on this branch:${NC}"
    git log main.."$CURRENT_BRANCH" --oneline --graph
    ;;
  
  merge)
    check_not_main
    
    CURRENT_BRANCH=$(get_current_branch)
    
    echo "${BLUE}🔀 Merging $CURRENT_BRANCH to main...${NC}"
    echo ""
    
    # Verify branch is pushed
    if ! git rev-parse origin/"$CURRENT_BRANCH" > /dev/null 2>&1; then
      echo "${RED}❌ Branch not pushed to remote${NC}"
      echo "   Run: ./scripts/git-flow.sh publish"
      exit 1
    fi
    
    # Show what will be merged
    echo "${YELLOW}Changes to be merged:${NC}"
    git log main.."$CURRENT_BRANCH" --oneline
    echo ""
    
    # Confirm merge
    read -p "Continue with merge? (y/n) " -n 1 -r
    echo
    if ! [[ $REPLY =~ ^[Yy]$ ]]; then
      echo "Merge cancelled"
      exit 0
    fi
    
    # Perform merge
    git checkout main
    git pull origin main
    git merge --ff-only "origin/$CURRENT_BRANCH"
    echo "${GREEN}✅ Merged to main${NC}"
    
    # Cleanup
    echo "${BLUE}🧹 Cleaning up branch...${NC}"
    git branch -d "$CURRENT_BRANCH"
    git push origin --delete "$CURRENT_BRANCH"
    echo "${GREEN}✅ Branch deleted${NC}"
    ;;
  
  cleanup)
    echo "${BLUE}🧹 Cleaning up branches...${NC}"
    echo ""
    
    # Remove local merged branches
    echo "${BLUE}Removing local merged branches...${NC}"
    git branch --merged | grep -v "main\|master\|\*" | xargs -r git branch -d
    
    # Remove remote tracking references for deleted branches
    echo "${BLUE}Removing stale remote references...${NC}"
    git fetch -p origin
    
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
