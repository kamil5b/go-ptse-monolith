# Developer Scripts Guide

## Overview

The Go Modular Monolith project includes a comprehensive suite of developer experience scripts that automate common tasks, enforce best practices, and streamline the development workflow.

## Quick Links

- **All Scripts Location:** `scripts/` directory
- **Scripts Reference:** Run `./scripts/[script-name].sh help` for any script
- **Full Documentation:** See `scripts/README.md`
- **Quick Cheat Sheet:** See `SCRIPTS_QUICK_REFERENCE.md` (root directory)

## Available Scripts

### 1. setup.sh - Initial Project Setup
**Purpose:** Initializes the project with all dependencies and development tools

```bash
./scripts/setup.sh
```

**What it does:**
- Verifies Go version (1.24.7+)
- Downloads and tidies Go modules
- Installs development tools:
  - golangci-lint (code linting)
  - mockgen (mock generation)
  - goose (database migrations)
- Generates mocks from interfaces
- Runs linter checks
- Verifies module dependency isolation
- Creates necessary directories

**When to use:** Run once after cloning the repository

**Example:**
```bash
git clone <repo-url>
cd github.com/kamil5b/go-ptse-monolith
./scripts/setup.sh
```

---

### 2. dev.sh - Development Commands
**Purpose:** Quick access to 19+ daily development commands

```bash
./scripts/dev.sh [command] [options]
```

**Available commands:**

| Command | Description |
|---------|-------------|
| `run` | Run application (with migrations) |
| `server` | Start server only (skip migrations) |
| `test [pattern]` | Run tests with optional filter |
| `test:unit` | Run unit tests only |
| `test:cover` | Run tests with coverage report |
| `lint` | Run golangci-lint code quality checks |
| `deps` | Check module dependencies for violations |
| `mocks` | Generate mocks from interfaces |
| `fmt` | Format code with gofmt |
| `vet` | Run go vet static analyzer |
| `mod:tidy` | Tidy Go modules |
| `mod:verify` | Verify module integrity |
| `db:up` | Apply SQL migrations (up) |
| `db:down` | Rollback SQL migrations (down) |
| `db:mongo:up` | Apply MongoDB migrations |
| `mongo:shell` | Connect to MongoDB shell |
| `postgres:shell` | Connect to PostgreSQL shell |
| `clean` | Clean generated files and caches |
| `help` | Show help message |

**Examples:**
```bash
./scripts/dev.sh run                          # Run the app
./scripts/dev.sh test ./internal/modules/user # Test specific module
./scripts/dev.sh test:cover                   # Run tests with coverage
./scripts/dev.sh lint                         # Check code quality
./scripts/dev.sh db:up                        # Apply migrations
./scripts/dev.sh postgres:shell               # Connect to PostgreSQL
```

**Benefits:**
- Consistent command naming
- Colored output for readability
- Built-in error handling
- Help system for all commands

---

### 3. db.sh - Database Management
**Purpose:** Manage local databases using Docker Compose

```bash
./scripts/db.sh [command]
```

**Available commands:**

| Command | Description |
|---------|-------------|
| `up` | Start all databases (PostgreSQL + MongoDB) |
| `down` | Stop all databases |
| `restart` | Restart all databases |
| `logs` | Show database logs |
| `ps` | Show running containers |
| `postgres:shell` | Connect to PostgreSQL CLI |
| `mongo:shell` | Connect to MongoDB shell |
| `postgres:dump` | Backup PostgreSQL database |
| `postgres:restore <file>` | Restore PostgreSQL database |
| `clean` | Remove containers and volumes |

**Prerequisites:**
- Docker and Docker Compose installed
- Script auto-creates `docker-compose.yml` if missing

**Examples:**
```bash
./scripts/db.sh up                 # Start databases
./scripts/db.sh postgres:shell     # Connect to PostgreSQL
./scripts/db.sh postgres:dump      # Backup database
./scripts/db.sh postgres:restore backups/postgres_20250103_120000.sql
./scripts/db.sh mongo:shell        # Connect to MongoDB
./scripts/db.sh down               # Stop databases
```

**Connection Strings (after `db.sh up`):**
- PostgreSQL: `postgresql://postgres:postgres@localhost:5432/appdb`
- MongoDB: `mongodb://admin:admin@localhost:27017/appdb?authSource=admin`

---

### 4. git-flow.sh - Git Workflow Automation
**Purpose:** Automate feature branch workflow with validation

```bash
./scripts/git-flow.sh [command]
```

**Available commands:**

| Command | Description |
|---------|-------------|
| `start <name>` | Create a new feature branch |
| `status` | Show branch status and changes |
| `pre-commit` | Run all pre-commit checks |
| `push` | Push current branch |
| `publish` | Push with upstream tracking |
| `sync` | Fetch and rebase on main |
| `diff` | Show diff with main branch |
| `commits` | Show commits on this branch |
| `merge` | Merge to main and cleanup |
| `cleanup` | Delete merged local/remote branches |

**Feature Development Workflow:**

1. **Start a feature:**
```bash
./scripts/git-flow.sh start add-payment-module
```

2. **Make your changes:**
```bash
# Edit files, commit as usual
git add .
git commit -m "Add payment module"
```

3. **Run pre-commit checks:**
```bash
./scripts/git-flow.sh pre-commit
```
This runs:
- Health checks (formatting, linting, structure)
- Full test suite
- Module dependency isolation checks

4. **Publish to remote:**
```bash
./scripts/git-flow.sh publish
```

5. **Keep in sync while waiting for review:**
```bash
./scripts/git-flow.sh sync
```

6. **Merge to main:**
```bash
./scripts/git-flow.sh merge
```
This will:
- Verify branch is pushed
- Show what will be merged
- Confirm with you before merging
- Use fast-forward merge
- Delete local and remote branch

**Features:**
- Validates branch names (lowercase, alphanumeric, hyphens)
- Automatic upstream tracking
- Pre-commit validation (health, tests, dependencies)
- Safe merging with confirmation
- Automatic cleanup of merged branches

---

### 5. new-module.sh - Module Scaffolding
**Purpose:** Scaffold new modules with correct architecture

```bash
./scripts/new-module.sh <module_name>
```

**Prerequisites:**
- Module name must be lowercase alphanumeric (e.g., `order`, `payment`, `invoice`)

**Example:**
```bash
./scripts/new-module.sh payment
```

**Creates complete structure:**
```
internal/modules/payment/
├── domain/
│   ├── model.go           # Domain entities
│   ├── interfaces.go      # Handler/Service/Repository interfaces
│   ├── request.go         # Request DTOs
│   ├── response.go        # Response DTOs
│   ├── events.go          # Domain events
│   └── mocks/
├── handler/
│   ├── v1/handler_v1.payment.go    # v1 implementation
│   └── noop/handler_noop.payment.go # No-op implementation
├── service/
│   ├── v1/service_v1.payment.go    # v1 implementation
│   └── noop/service_noop.payment.go # No-op implementation
├── repository/
│   ├── sql/repository_sql.payment.go      # SQL implementation
│   ├── mongo/repository_mongo.payment.go  # MongoDB implementation
│   └── noop/repository_noop.payment.go    # No-op implementation
└── acl/
    └── .gitkeep  # Anti-Corruption Layer folder
```

**Next steps:**
1. Define domain models in `domain/model.go`
2. Define interfaces in `domain/interfaces.go`
3. Implement handlers in `handler/v1/`
4. Implement services in `service/v1/`
5. Implement repositories in `repository/sql/` and `repository/mongo/`
6. Wire up in `internal/app/core/container.go`
7. Add routes in `internal/app/http/routes.go`
8. Add feature flags in `config/featureflags.yaml`
9. Verify isolation: `./scripts/lint-deps-check.sh --verbose`

---

### 6. lint-deps-check.sh - Module Dependency Verification
**Purpose:** Enhanced linter for checking module isolation and cross-module dependencies

```bash
./scripts/lint-deps-check.sh [--verbose]
```

**Options:**
- `--verbose`: Show detailed analysis of all checked files

**What it checks:**
- ✅ Cross-module imports (violations)
- ✅ ACL imports (allowed to import other modules)
- ✅ Shared kernel imports (allowed everywhere)
- ✅ Infrastructure imports (module-level restrictions)

**Example output:**
```
Module Dependency Linter
Analyzing cross-module dependencies...

✅ No cross-module dependency violations found!

Module Structure:
  • auth
  • product
  • user
```

**When violations are found:**
The script provides actionable guidance:
- Use **Anti-Corruption Layer (ACL)** for synchronous operations
- Use **Event Bus** for asynchronous operations
- See `TECHNICAL_DOCUMENTATION.md` for ACL pattern details

**Examples:**
```bash
./scripts/lint-deps-check.sh                      # Quick check
./scripts/lint-deps-check.sh --verbose            # Detailed analysis
./scripts/lint-deps-check.sh --verbose | grep "❌" # Show only violations
```

**Enforced Rules:**
1. No cross-module imports (except ACL)
2. ACL can import from other modules
3. All modules can import from shared kernel
4. Domain types are module-specific (not shared)

---

### 7. health-check.sh - Project Health Assessment
**Purpose:** Comprehensive project health check with auto-fix capability

```bash
./scripts/health-check.sh [--fix]
```

**Options:**
- `--fix`: Automatically fix fixable issues

**Checks:**

| Check | Details |
|-------|---------|
| Go environment | Verifies Go version compatibility |
| Dependencies | Validates go.mod and go.sum |
| Code formatting | Checks gofmt compliance |
| Code quality | Runs golangci-lint and go vet |
| Module structure | Verifies required directories exist |
| Dependency isolation | Ensures no cross-module violations |
| Configuration | Validates config.yaml and featureflags.yaml |

**Example output:**
```
Project Health Check

🔍 Go Environment
  Go version: 1.24.7

📦 Dependencies
  Checking go.mod tidiness... ✅
  Checking go.sum consistency... ✅

...

=== Health Check Summary ===
  Score: 18/18 (100%)

✅ Project is healthy!
```

**When to run:**
- Before committing code
- Before deploying
- In CI/CD pipelines
- As part of team workflows

**Examples:**
```bash
./scripts/health-check.sh          # Check health
./scripts/health-check.sh --fix    # Auto-fix issues
```

---

### 8. generate_mocks_from_source.sh - Mock Generation
**Purpose:** Generate mocks from interface definitions using mockgen

```bash
./scripts/generate_mocks_from_source.sh
```

**What it does:**
- Scans all `.go` files in `internal/` folder
- Finds interface declarations
- Generates mocks using mockgen in source mode
- Places mocks in `<module>/domain/mocks/` folder

**When to use:**
- After defining new interfaces
- Via `./scripts/dev.sh mocks` command

**Example:**
```bash
# Define an interface in internal/modules/payment/domain/interfaces.go
type PaymentRepository interface {
    Create(ctx context.Context, payment *Payment) error
}

# Generate mocks
./scripts/generate_mocks_from_source.sh
# or
./scripts/dev.sh mocks

# Mock will be in: internal/modules/payment/domain/mocks/mock_interfaces.go
```

---

## Common Workflows

### Morning: Start Development
```bash
git pull origin main
./scripts/health-check.sh
./scripts/db.sh up
./scripts/dev.sh run
```

### Create a Feature
```bash
./scripts/git-flow.sh start add-new-feature
# ... implement feature ...
./scripts/git-flow.sh pre-commit
./scripts/git-flow.sh publish
```

### Create a New Module
```bash
./scripts/new-module.sh payment
# ... implement module ...
./scripts/lint-deps-check.sh --verbose
./scripts/dev.sh test
./scripts/health-check.sh
```

### Before Committing
```bash
./scripts/git-flow.sh pre-commit  # Runs health check + tests + deps check
git add .
git commit -m "Your message"
./scripts/git-flow.sh publish
```

### End of Day
```bash
./scripts/dev.sh clean
./scripts/db.sh down
```

---

## Best Practices

### 1. Always Run Pre-Commit Checks
```bash
./scripts/git-flow.sh pre-commit  # Before pushing code
```

### 2. Keep Module Isolation
```bash
./scripts/lint-deps-check.sh --verbose  # Regularly verify
```

### 3. Monitor Project Health
```bash
./scripts/health-check.sh  # Weekly or before releases
```

### 4. Use Feature Branches
```bash
./scripts/git-flow.sh start feature-name  # Always use git-flow for consistency
```

### 5. Test Frequently
```bash
./scripts/dev.sh test  # During development
./scripts/dev.sh test:cover  # Before commits
```

---

## Troubleshooting

### Issue: Script Permission Denied
```bash
chmod +x scripts/*.sh
```

### Issue: Docker Not Found
- Install Docker Desktop: https://www.docker.com/products/docker-desktop
- Or install docker-compose: https://docs.docker.com/compose/install/

### Issue: golangci-lint Not Found
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Issue: Database Connection Error
```bash
./scripts/db.sh ps      # Check if databases are running
./scripts/db.sh logs    # View database logs
./scripts/db.sh down    # Stop all
./scripts/db.sh up      # Start all
```

### Issue: Module Isolation Violations
1. Run: `./scripts/lint-deps-check.sh --verbose`
2. Review violations
3. Use ACL pattern or Event Bus for cross-module communication
4. See `TECHNICAL_DOCUMENTATION.md` for examples

---

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Setup
  run: ./scripts/setup.sh

- name: Health Check
  run: ./scripts/health-check.sh --fix

- name: Tests
  run: ./scripts/dev.sh test

- name: Dependency Check
  run: ./scripts/lint-deps-check.sh
```

### Pre-commit Hook
```bash
#!/bin/bash
./scripts/git-flow.sh pre-commit || exit 1
```

---

## See Also

- **scripts/README.md** - Complete reference documentation
- **SCRIPTS_QUICK_REFERENCE.md** - Quick command cheat sheet
- **SCRIPTS_COMPLETE_PACKAGE.md** - Executive overview
- **TECHNICAL_DOCUMENTATION.md** - Architecture and patterns that scripts enforce
- **Makefile** - Alternative task runner (still supported)

---

## Pro Tips

### 1. Create Shell Aliases
```bash
alias dev="./scripts/dev.sh"
alias db="./scripts/db.sh"
alias check="./scripts/health-check.sh"

# Then use: dev test, db up, check
```

### 2. Watch Tests During Development
```bash
# Terminal 1:
./scripts/dev.sh run

# Terminal 2:
watch ./scripts/dev.sh test
```

### 3. Keep Branch Updated
```bash
./scripts/git-flow.sh sync  # Run regularly while waiting for review
```

### 4. Automate Pre-commit
```bash
# Add to .git/hooks/pre-commit
#!/bin/bash
./scripts/health-check.sh || exit 1
```

---

## Support

For help with any script:
```bash
./scripts/[script-name].sh help
./scripts/[script-name].sh --help
./scripts/[script-name].sh -h
```

For comprehensive documentation:
```bash
cat scripts/README.md
```

---

**Made with 💚 for developer productivity!**
