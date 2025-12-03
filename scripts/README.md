# Developer Experience Scripts

This directory contains helpful scripts to improve the development experience for the Go Modular Monolith project.

## Quick Start

### 1. Initial Setup
```bash
chmod +x scripts/*.sh
./scripts/setup.sh
```

This will:
- Install Go dependencies
- Install development tools (golangci-lint, mockgen, goose)
- Generate mocks from interfaces
- Run linter checks
- Verify module dependencies
- Check configuration files

### 2. Start Development
```bash
./scripts/dev.sh help        # See all available commands
./scripts/dev.sh run         # Run the application
./scripts/db.sh up           # Start databases (PostgreSQL + MongoDB)
```

## Available Scripts

### 📦 `setup.sh`
**Initializes the project with all dependencies and tools**

```bash
./scripts/setup.sh
```

**What it does:**
- Verifies Go version compatibility
- Downloads Go modules
- Installs development tools
- Creates necessary directories
- Generates mocks
- Runs linter checks
- Verifies module dependency isolation

**When to use:** First time setup or after major dependency changes

---

### 🛠️ `dev.sh`
**Quick access to common development commands**

```bash
./scripts/dev.sh [command]
```

**Available commands:**

| Command | Description |
|---------|-------------|
| `run` | Run application (with migrations) |
| `server` | Start server only |
| `test [pattern]` | Run tests (optional filter) |
| `test:unit` | Run unit tests only |
| `test:cover` | Run tests with coverage report |
| `lint` | Run linter (golangci-lint) |
| `deps` | Check module dependencies |
| `mocks` | Generate mocks from interfaces |
| `fmt` | Format code (gofmt) |
| `vet` | Run go vet |
| `mod:tidy` | Tidy Go modules |
| `mod:verify` | Verify module integrity |
| `db:up` | Run SQL migrations up |
| `db:down` | Run SQL migrations down |
| `db:mongo:up` | Run MongoDB migrations up |
| `mongo:shell` | Connect to MongoDB shell |
| `postgres:shell` | Connect to PostgreSQL shell |
| `clean` | Remove generated files and caches |
| `help` | Show help message |

For Git workflow commands, use: `./scripts/git-flow.sh help`

**Examples:**
```bash
./scripts/dev.sh run
./scripts/dev.sh test ./internal/modules/user
./scripts/dev.sh test:cover
./scripts/dev.sh lint
./scripts/dev.sh db:up
```

---

### 🗄️ `db.sh`
**Manages local databases using Docker Compose**

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
| `postgres:dump` | Dump PostgreSQL database |
| `postgres:restore <file>` | Restore PostgreSQL database |
| `clean` | Remove containers and volumes |

**Prerequisites:**
- Docker and Docker Compose installed
- (Script auto-creates `docker-compose.yml` if needed)

**Examples:**
```bash
./scripts/db.sh up              # Start databases
./scripts/db.sh postgres:shell  # Connect to PostgreSQL
./scripts/db.sh postgres:dump   # Backup database
./scripts/db.sh down            # Stop databases
```

**Connection strings (after running `db.sh up`):**
- **PostgreSQL:** `postgresql://postgres:postgres@localhost:5432/appdb`
- **MongoDB:** `mongodb://admin:admin@localhost:27017/appdb?authSource=admin`

---

### 📦 `new-module.sh`
**Scaffolds a new module with correct structure and boilerplate**

```bash
./scripts/new-module.sh <module_name>
```

**Prerequisites:**
- Module name must be lowercase and alphanumeric (e.g., `order`, `invoice`, `shipping`)

**Example:**
```bash
./scripts/new-module.sh order
```

**Creates:**
```
internal/modules/order/
├── domain/
│   ├── model.go
│   ├── interfaces.go
│   ├── request.go
│   ├── response.go
│   ├── events.go
│   └── mocks/
├── handler/
│   ├── v1/handler_v1.order.go
│   └── noop/handler_noop.order.go
├── service/
│   ├── v1/service_v1.order.go
│   └── noop/service_noop.order.go
├── repository/
│   ├── sql/repository_sql.order.go
│   ├── mongo/repository_mongo.order.go
│   └── noop/repository_noop.order.go
└── acl/
    └── .gitkeep
```

**Next steps after creating a module:**
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

### 🔗 `lint-deps-check.sh`
**Enhanced module dependency linter with detailed violation reports**

```bash
./scripts/lint-deps-check.sh [--verbose] [--fix]
```

**Options:**
- `--verbose`: Show all checked files and detailed analysis
- `--fix`: (reserved for future enhancements)

**What it checks:**
- ✅ Cross-module imports (should not exist)
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
./scripts/lint-deps-check.sh
./scripts/lint-deps-check.sh --verbose  # Detailed analysis
./scripts/lint-deps-check.sh --verbose | grep "❌"  # Show violations only
```

---

### 💚 `health-check.sh`
**Comprehensive project health check**

```bash
./scripts/health-check.sh [--fix]
```

**Checks:**
- Go environment and version
- Module dependencies (go.mod, go.sum)
- Code formatting (gofmt)
- Code quality (golangci-lint, go vet)
- Module structure (required directories)
- Dependency isolation (cross-module imports)
- Configuration files

**Options:**
- `--fix`: Automatically fix fixable issues

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

**When to run:** Before committing code or deploying

---

### 🎭 `generate_mocks_from_source.sh`
**Generates mocks from interfaces using mockgen**

```bash
./scripts/generate_mocks_from_source.sh
```

**What it does:**
- Scans all `.go` files in `internal/`
- Finds interface declarations
- Generates mocks using mockgen in source mode
- Places mocks in `<module>/domain/mocks/`

**When to use:** After defining new interfaces or run via `./scripts/dev.sh mocks`

---

### 🌿 `git-flow.sh`
**Automates common git workflow tasks for feature development**

```bash
./scripts/git-flow.sh [command]
```

**Available commands:**

| Command | Description |
|---------|-------------|
| `start <name>` | Create a new feature branch |
| `status` | Show branch status and changes |
| `pre-commit` | Run all checks before commit |
| `push` | Push current branch |
| `publish` | Push with upstream tracking |
| `sync` | Fetch and rebase on main |
| `diff` | Show diff with main branch |
| `commits` | Show commits on this branch |
| `merge` | Merge to main and cleanup |
| `cleanup` | Delete merged local/remote branches |

**Git Workflow:**

1. **Start a feature:**
   ```bash
   ./scripts/git-flow.sh start add-payment-module
   ```

2. **Implement your feature**
   ```bash
   # Make changes...
   git add .
   git commit -m "Add payment module"
   ```

3. **Pre-commit checks:**
   ```bash
   ./scripts/git-flow.sh pre-commit
   ```
   This runs:
   - Health checks (formatting, linting, structure)
   - Full test suite
   - Dependency isolation checks

4. **Publish to remote:**
   ```bash
   ./scripts/git-flow.sh publish
   ```

5. **Keep in sync with main:**
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
   - Merge using fast-forward
   - Delete local and remote branch

**Features:**
- Validates branch names (lowercase, hyphens, numbers)
- Automatic upstream tracking
- Pre-commit checks (health, tests, dependencies)
- Safe merging with confirmation
- Automatic cleanup of merged branches

**Example workflow:**
```bash
./scripts/git-flow.sh start implement-caching
# ... make changes ...
./scripts/git-flow.sh pre-commit     # Verify everything works
./scripts/git-flow.sh publish        # Push to GitHub
# ... open Pull Request, get review ...
./scripts/git-flow.sh merge          # Merge to main
```

---

## Development Workflow

### Start of Day
```bash
# Update code
git pull origin main

# Verify everything works
./scripts/health-check.sh

# Start databases
./scripts/db.sh up

# Run the app
./scripts/dev.sh run
```

### Start a New Feature
```bash
# Create feature branch
./scripts/git-flow.sh start add-new-feature

# Start databases if not running
./scripts/db.sh up

# Begin development
./scripts/dev.sh run
```

### During Development
```bash
# Run tests frequently
./scripts/dev.sh test ./internal/modules/your-module

# Check code quality
./scripts/dev.sh lint

# Verify module isolation
./scripts/lint-deps-check.sh --verbose

# Format code
./scripts/dev.sh fmt

# Keep in sync with main
./scripts/git-flow.sh sync
```

### Creating a New Module
```bash
# Scaffold the module
./scripts/new-module.sh payment

# Implement the module
# (edit files as needed)

# Verify isolation
./scripts/lint-deps-check.sh --verbose

# Test it
./scripts/dev.sh test

# Check health
./scripts/health-check.sh
```

### Before Committing
```bash
# Run all pre-commit checks (health, tests, dependencies)
./scripts/git-flow.sh pre-commit

# If checks pass, commit
git add .
git commit -m "Descriptive message"
```

### Submitting Code
```bash
# Publish feature branch
./scripts/git-flow.sh publish

# (Create Pull Request on GitHub)

# Keep branch in sync while waiting for review
./scripts/git-flow.sh sync

# When approved, merge to main
./scripts/git-flow.sh merge
```

### End of Day
```bash
# Save your work
git push

# Clean up
./scripts/dev.sh clean

# Stop databases
./scripts/db.sh down
```

---

## Setup Instructions

### Make Scripts Executable
```bash
chmod +x scripts/*.sh
```

### Initial Project Setup
```bash
./scripts/setup.sh
```

### Subsequent Uses
```bash
# Just use the scripts!
./scripts/dev.sh help
./scripts/db.sh help
```

---

## Requirements

### Required
- Go 1.24.7+ (checked by setup.sh)
- bash 4.0+ (for script compatibility)

### Optional (installed by setup.sh)
- golangci-lint - Code linting
- mockgen - Mock generation
- goose - Database migrations

### Optional (for specific commands)
- Docker + Docker Compose - For `db.sh`
- PostgreSQL client (psql) - For PostgreSQL shell access
- MongoDB tools (mongosh) - For MongoDB shell access

---

## Troubleshooting

### Script Permission Errors
```bash
chmod +x scripts/*.sh
```

### Docker not found
- Install [Docker Desktop](https://www.docker.com/products/docker-desktop)
- Or install [docker-compose](https://docs.docker.com/compose/install/)

### golangci-lint errors during setup
- These are warnings, not failures
- Install manually: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### Database connection errors
- Ensure `./scripts/db.sh up` has completed successfully
- Check with `./scripts/db.sh ps`
- Check logs with `./scripts/db.sh logs`

### Module isolation violations
- Review the output of `./scripts/lint-deps-check.sh --verbose`
- Use ACL pattern or Event Bus for cross-module communication
- See `TECHNICAL_DOCUMENTATION.md` for examples

---

## Contributing

When adding new development scripts:
1. Follow the naming convention: `lowercase-with-dashes.sh`
2. Add color output for better UX (use the color definitions)
3. Include help/usage information
4. Document in this README
5. Make scripts idempotent (safe to run multiple times)
6. Provide clear error messages

---

## See Also

- `TECHNICAL_DOCUMENTATION.md` - Full project architecture and patterns
- `ROADMAP_CHECKLIST.md` - Project roadmap and status
- `Makefile` - Alternative command runner
- `go.mod` - Project dependencies

