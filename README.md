# Go-PSTE-Boilerplate

A production-ready, modular monolithic boilerplate built with Go implementing clean architecture principles with **Plug, Swap, Toggle, Extract** capabilities.

[![Go Version](https://img.shields.io/badge/Go-1.24.7-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## PSTE Architecture

**Plug** • **Swap** • **Toggle** • **Extract**

### Core Capabilities

- 🔌 **Plug** - Pluggable infrastructure components:
  - HTTP frameworks (Echo, Gin, Fiber, fasthttp, net/http)
  - Cache backends (Redis, Memory)
  - File storage (S3, GCS, Local, Noop)
  - Email services (SMTP, Mailgun)
  - Event bus implementations (Memory, RabbitMQ, Redpanda)
  - Worker backends (Asynq, RabbitMQ, Cron Scheduler)

- 🔄 **Swap** - Swap implementations per module:
  - Database backends (PostgreSQL, MongoDB)
  - Service implementations (v1, noop, disable)
  - Handler implementations (v1, noop, disable)
  - Repository implementations (SQL, Mongo, noop)

- 🎛️ **Toggle** - Toggle features and implementations via feature flags
- 📦 **Extract** - Design for easy extraction into microservices

### Additional Features

- 🔐 **Complete Authentication** - JWT, Session-based, and Basic Auth
- 🛡️ **Middleware Support** - Authentication, authorization, and role-based access
- 📦 **Modular Architecture** - Domain-per-module with isolated boundaries
- 🎛️ **Feature Flags** - Enable/disable features through configuration
- 🔄 **Database Migrations** - Goose (SQL) and mongosh (MongoDB)
- 🧩 **Shared Kernel** - Events, Errors, Context, UoW, Validator
- 🔗 **Anti-Corruption Layer** - Clean cross-module communication via ACL
- 🔍 **Dependency Linter** - Enforces module isolation rules

## Quick Start

### Prerequisites

- Go 1.24.7+
- PostgreSQL 14+
- MongoDB 6.0+ (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/kamil5b/go-pste-boilerplate.git
cd go-pste-boilerplate

# Install dependencies
go mod tidy

# Configure your database
# Edit config/config.yaml with your credentials

# Run the application
go run .
```

### CLI Commands

```bash
# Application
go run .                      # Run migrations and start HTTP + gRPC servers
go run . server               # Start servers only (skip migrations)
go run . worker               # Start worker only

# Database Migrations
go run . migration sql up     # Apply SQL migrations
go run . migration sql down   # Rollback SQL migrations
go run . migration mongo up   # Apply MongoDB migrations

# Protocol Buffers
make proto                    # Generate protobuf code for all modules
make proto-product            # Generate protobuf code for product module

# Development
make deps-check               # Check module dependency violations
make test                     # Run all tests
make lint                     # Run linter
```

## Configuration

### Application Config (`config/config.yaml`)

```yaml
environment: development

app:
  server:
    port: "8080"        # HTTP server port
    grpc_port: "9090"   # gRPC server port

  database:
    sql:
      db_url: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
    mongo:
      mongo_url: "mongodb://localhost:27017"
      mongo_db: "myapp_db"

  jwt:
    secret: "your-secret-key"
```

### Feature Flags (`config/featureflags.yaml`)

```yaml
http_handler: echo  # echo | gin

handler:
  authentication: v1  # v1 | disable
  product: v1
  user: v1

service:
  authentication: v1
  product: v1
  user: v1

repository:
  authentication: postgres  # postgres | mongo | disable
  product: postgres
  user: postgres
```

## Project Structure

```
go-pste-boilerplate/
├── cmd/
│   ├── bootstrap/          # Application bootstrapping
│   └── lint-deps/          # Dependency linter tool
├── config/                 # Configuration files
├── docs/                   # Documentation
├── internal/
│   ├── app/               # Application core (DI, config, HTTP setup)
│   ├── infrastructure/    # Database connections, external services
│   ├── modules/           # Business modules (auth, product, user)
│   │   └── <module>/
│   │       ├── domain/    # Module's private domain types
│   │       │   └── proto/ # Protocol Buffer definitions (source)
│   │       ├── proto/     # Generated protobuf code
│   │       │   ├── v1/    # Version-specific generated code
│   │       │   └── adapters/ # Domain/Proto converters
│   │       ├── acl/       # Anti-Corruption Layer adapters
│   │       ├── handler/   # HTTP and gRPC handlers (v1, noop)
│   │       ├── service/   # Business logic (v1, noop)
│   │       └── repository/# Data access (sql, mongo, noop)
│   ├── shared/            # Shared kernel (cross-cutting concerns)
│   │   ├── context/       # Framework-agnostic HTTP context
│   │   ├── errors/        # Domain error types
│   │   ├── events/        # Event bus for inter-module communication
│   │   ├── uow/           # Unit of Work interface
│   │   └── validator/     # Request validation
│   └── transports/        # HTTP and gRPC framework adapters
└── pkg/                   # Shared utilities
```

## API Endpoints

### HTTP Endpoints

#### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | User login |
| POST | `/auth/register` | User registration |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/validate` | Validate token |

### Authentication (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/logout` | User logout |
| GET | `/auth/profile` | Get user profile |
| PUT | `/auth/password` | Change password |
| GET | `/auth/sessions` | List active sessions |
| DELETE | `/auth/sessions/:id` | Revoke specific session |
| DELETE | `/auth/sessions` | Revoke all sessions |

### Products (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/product` | List products |
| POST | `/product` | Create product |

### Users (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/user` | List users |
| POST | `/user` | Create user |
| GET | `/user/:id` | Get user by ID |
| PUT | `/user/:id` | Update user |
| DELETE | `/user/:id` | Delete user |

### gRPC Services

The application exposes gRPC services alongside HTTP endpoints for high-performance communication.

#### Product Service (Port 9090)

| Method | Service | Description |
|--------|---------|-------------|
| Create | `product.v1.ProductService/Create` | Create a new product |
| Get | `product.v1.ProductService/Get` | Get product by ID |
| List | `product.v1.ProductService/List` | List all products |
| Update | `product.v1.ProductService/Update` | Update existing product |
| Delete | `product.v1.ProductService/Delete` | Delete product |

**Test with grpcurl:**
```bash
# List available services
grpcurl -plaintext localhost:9090 list

# Create product
grpcurl -plaintext -d '{"name":"Test Product","description":"A test"}' \
  localhost:9090 product.v1.ProductService/Create

# Get product
grpcurl -plaintext -d '{"id":"123"}' \
  localhost:9090 product.v1.ProductService/Get
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24.7 |
| HTTP Frameworks | Echo v4, Gin v1, Fiber, fasthttp, net/http |
| RPC Framework | gRPC with Protocol Buffers |
| SQL Database | PostgreSQL (sqlx) |
| NoSQL Database | MongoDB |
| Cache | Redis, Memory |
| File Storage | S3, GCS, Local, Noop |
| Email | SMTP, Mailgun |
| Event Bus | Memory, RabbitMQ, Redpanda |
| Workers | Asynq, RabbitMQ, Cron Scheduler |
| Authentication | JWT (golang-jwt/jwt/v5) |
| Migrations | Goose, mongosh |
| Logging | Zerolog |

## Strict Import Regulations

To maintain module isolation and prevent dependency chaos, PSTE enforces strict dependency boundaries:

### Module Isolation Rules

- ✅ **Modules can import**: Own domain, shared kernel, own ACL
- ❌ **Modules cannot import**: Other modules directly, /app/**, /infrastructure/**
- ✅ **ACL (Anti-Corruption Layer) can import**: Own domain + other module domains (by design)
- ✅ **Shared kernel imports**: Only stdlib and external packages (never modules)

### Layer-Specific Rules

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| **Domain** | Shared kernel only | Other modules, services |
| **Handler** | Own domain, shared | Other modules, other handlers |
| **Service** | Own domain, shared, own ACL | Other modules (except via ACL) |
| **Repository** | Own domain, shared | Other modules |
| **ACL** | Own + other module domains | Other layer implementations |

### Dependency Linter

A built-in linter enforces these rules automatically:

```bash
go run cmd/lint-deps/main.go        # Check violations
go run cmd/lint-deps/main.go -v     # Verbose output
```

The linter catches:
- Cross-module imports
- Deprecated domain paths
- Invalid /app and /infrastructure imports
- Shared kernel violations
- Cyclic dependencies

## Architecture Deep Dive

### What is PSTE?

PSTE is a synthesis of proven architectural patterns designed for enterprise-grade, long-lived applications:

| Pattern | How PSTE Uses It |
|---------|-----------------|
| **Clean Architecture** | Separates concerns into layers (Domain → Service → Handler → Repository) |
| **Modular Monolith** | Single deployment with multiple independent modules |
| **Domain-Driven Design** | Each module has its own domain models and boundaries |
| **Dependency Injection** | Central container manages object creation and wiring |
| **Strategy/Factory Pattern** | Multiple implementations selectable at runtime |
| **Event-Driven Architecture** | Async inter-module communication via event bus |
| **Strict Import Rules** | Enforced boundaries prevent coupling |

### PSTE Strengths

✅ **Flexibility** - Swap any component without code changes  
✅ **Testability** - Clean layers + DI make unit testing straightforward  
✅ **Modularity** - Clear boundaries prevent architectural decay  
✅ **Scalability** - Extract modules to microservices when needed  
✅ **Maintainability** - Domain isolation makes code easier to understand  
✅ **Future-proof** - Designed for evolution, not fixed architecture  
✅ **Team autonomy** - Teams own modules, clear ownership boundaries  

### PSTE Trade-offs

⚖️ **Complexity** - More files and abstractions than simpler patterns  
⚖️ **Learning curve** - Developers need to understand multiple patterns  
⚖️ **Boilerplate** - More code for simple operations (CRUD endpoints)  
⚖️ **Runtime overhead** - Interface indirection + DI container add latency  
⚖️ **Implicit dependencies** - Events make dependencies less visible  
⚖️ **Over-engineering risk** - Easy to add abstraction when not needed  

### When PSTE Shines

✅ Medium to large teams (5+ engineers)  
✅ Products with evolving requirements  
✅ Multi-domain business logic  
✅ Teams planning microservices migration  
✅ Projects lasting 2+ years  

### When PSTE May Be Overkill

⚠️ Startups/MVP phase (use simpler patterns first)  
⚠️ Simple CRUD APIs (Clean Architecture alone may suffice)  
⚠️ Single-person teams  
⚠️ Throw-away prototypes  
⚠️ Real-time/ultra-high-performance systems requiring minimal latency  

## Documentation

For detailed documentation, see [Technical Documentation](docs/TECHNICAL_DOCUMENTATION.md).

## Roadmap – Project Checklist

### Completed ✅

- [x] Architecture & Infrastructure setup (including .gitkeep placeholders)
- [x] Product CRUD: HTTP Echo → v1 → SQL repository
- [x] SQL & MongoDB repository support with migrations
- [x] Gin framework integration
- [x] Utilize Request & Response models
- [x] User CRUD implementation
- [x] Authentication system: JWT, Basic Auth, Session-based (untested)
- [x] Middleware integration (untested)
- [x] Shared Kernel (`internal/shared/`) - Events, Errors, Context, UoW, Validator
- [x] Domain-per-Module Pattern - Each module owns its domain types
- [x] Anti-Corruption Layer (ACL) - Clean cross-module communication
- [x] Dependency Linter (`cmd/lint-deps/`) - Enforces module isolation
- [x] Shared Context Interface (`sharedctx.Context`) - Framework-agnostic handlers
- [x] Redis integration (caching)
- [x] **Worker Support** - Asynq, RabbitMQ, and Redpanda integration
- [x] **Email Services** - SMTP and Mailgun support with worker integration
- [x] **Storage Support** - Local, AWS S3, S3-Compatible (MinIO), Google Cloud Storage
- [x] **gRPC & Protocol Buffers** - Full gRPC support with dual HTTP/gRPC handlers
- [x] **Proto Generation** - Automated script for generating protobuf code
- [x] **Unit Tests** - Comprehensive test coverage for core modules and shared kernel
  - Product module: 81-100% coverage (handler, service, gRPC, proto adapters)
  - User module: 84% service coverage
  - Auth module: 77.4% service coverage (login, register, tokens, sessions, passwords)
  - Shared kernel: 90-100% coverage (context, email, worker, validator, cache, errors)
  - Logger: 91.3% coverage

### In Progress 🚧

- [ ] Unit tests for HTTP handlers (Auth and User modules)
- [ ] Integration tests for database repositories
- [ ] End-to-end tests for complete workflows

### Planned 📋

- [ ] WebSocket integration
- [ ] OpenTelemetry integration for distributed tracing
- [ ] API documentation generation (Swagger/OpenAPI for REST, reflection for gRPC)

## Contributing

1. Follow the module structure outlined in the documentation
2. Use feature flags for new components
3. Implement both PostgreSQL and MongoDB repositories when applicable
4. Add migrations for database schema changes
5. **Run dependency linter before committing**: `make deps-check`
6. **Generate protobuf code after .proto changes**: `make proto` or `make proto-<module>`
7. Use ACL pattern for cross-module communication
8. Place `.proto` files in `domain/proto/v1/` directory
9. Update documentation for significant changes

## License

See [LICENSE](LICENSE) file for details.
