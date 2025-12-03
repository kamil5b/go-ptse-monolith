# Go Modular Monolith - Technical Documentation

> **Version:** 2.0.0  
> **Last Updated:** December 2, 2025  
> **Go Version:** 1.24.7

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Project Structure](#project-structure)
4. [Getting Started](#getting-started)
5. [Configuration](#configuration)
6. [Feature Flags](#feature-flags)
7. [Modules](#modules)
8. [Shared Kernel](#shared-kernel)
9. [Domain Layer](#domain-layer)
10. [Anti-Corruption Layer (ACL)](#anti-corruption-layer-acl)
11. [Infrastructure Layer](#infrastructure-layer)
12. [Transport Layer](#transport-layer)
13. [Authentication & Middleware](#authentication--middleware)
14. [Database & Migrations](#database--migrations)
15. [API Reference](#api-reference)
16. [Development Guide](#development-guide)
17. [Dependency Linter](#dependency-linter)
18. [Microservices Readiness](#microservices-readiness)
19. [Roadmap](#roadmap)

---

## Overview

Go Modular Monolith is a production-ready, modular monolithic application built with Go. It implements clean architecture principles with pluggable components, allowing teams to:

- **Switch HTTP frameworks** (Echo/Gin) via configuration
- **Switch HTTP frameworks** (Echo, Gin, net/http, fasthttp, Fiber) via configuration
- **Swap database backends** (PostgreSQL/MongoDB) per module
- **Version handlers, services, and repositories** independently
- **Enable/disable features** through feature flags
- **Maintain strict module isolation** with domain-per-module pattern
- **Communicate between modules** via Anti-Corruption Layer (ACL) or Event Bus
- **Prepare for microservices migration** with enforced dependency boundaries

### Key Technologies

| Component | Technology |
|-----------|------------|
| Language | Go 1.24.7 |
| HTTP Frameworks | Echo v4, Gin v1, net/http (stdlib), fasthttp, Fiber |
| SQL Database | PostgreSQL (via sqlx) |
| NoSQL Database | MongoDB |
| Migrations | Goose (SQL), mongosh (MongoDB) |
| Authentication | JWT (golang-jwt/jwt/v5) |
| Logging | Zerolog |
| Password Hashing | golang.org/x/crypto/bcrypt |
| UUID Generation | github.com/google/uuid |
| Validation | go-playground/validator/v10 |

---

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                      Transport Layer                        │
│  (Echo / Gin / net/http / FastHTTP / Fiber Adapters)        │
├─────────────────────────────────────────────────────────────┤
│                      Handler Layer                          │
│              (HTTP Request/Response Handling)               │
├─────────────────────────────────────────────────────────────┤
│                      Service Layer                          │
│                  (Business Logic)                           │
├─────────────────────────────────────────────────────────────┤
│                    Repository Layer                         │
│              (Data Access - SQL/MongoDB)                    │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure Layer                      │
│           (Database Connections, External Services)         │
└─────────────────────────────────────────────────────────────┘
```

### Module Isolation Pattern

Each module is fully isolated with its own domain types. Cross-module communication uses:
- **Anti-Corruption Layer (ACL)** for synchronous, transactional operations
- **Event Bus** for asynchronous, eventually consistent operations

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Shared Kernel                                  │
│  (internal/shared: events, errors, context, uow, validator)             │
└─────────────────────────────────────────────────────────────────────────┘
         ▲                    ▲                    ▲
         │                    │                    │
┌────────┴────────┐  ┌───────┴────────┐  ┌───────┴────────┐
│   Auth Module   │  │ Product Module │  │  User Module   │
├─────────────────┤  ├────────────────┤  ├────────────────┤
│  domain/        │  │  domain/       │  │  domain/       │
│  handler/       │  │  handler/      │  │  handler/      │
│  service/       │  │  service/      │  │  service/      │
│  repository/    │  │  repository/   │  │  repository/   │
│  acl/  ◄────────┼──┼────────────────┼──┤                │
│  middleware/    │  │                │  │                │
└─────────────────┘  └────────────────┘  └────────────────┘
         │                                       ▲
         └───────── ACL translates ─────────────┘
```

### Dependency Flow

```
main.go
   │
   ▼
bootstrap/
   │
   ├── LoadConfig()
   ├── LoadFeatureFlags()
   │
   ▼
Container (DI)
   │
   ├── EventBus (shared)
   ├── Repository (SQL/MongoDB)
   ├── ACL Adapters (cross-module)
   ├── Service (v1/noop)
   └── Handler (v1/noop)
   │
   ▼
HTTP Server (Echo/Gin/net/http/FastHTTP/Fiber)
   │
   ▼
Routes → Handlers → Services → Repositories → Database
                        │
                        ├── ACL (for cross-module)
                        └── EventBus (for async events)
```

### Dependency Rules

| Source | Can Import | Cannot Import |
|--------|------------|---------------|
| Module domain | Shared kernel only | Other modules |
| Module handler | Own domain, shared | Other modules |
| Module service | Own domain, shared, own ACL | Other module implementations |
| Module repository | Own domain, shared | Other modules |
| Module ACL | Own domain, OTHER module domains | - |
| Shared kernel | Standard library, external packages | Any module |

---

## Project Structure

```
go-modular-monolith/
├── main.go                          # Application entry point
├── go.mod                           # Go module definition
├── config/
│   ├── config.yaml                  # Application configuration
│   └── featureflags.yaml            # Feature flag configuration
├── cmd/
│   ├── bootstrap/
│   │   ├── bootstrap.server.go      # Server initialization
│   │   └── bootstrap.migration.go   # Migration runner
│   └── lint-deps/
│       └── main.go                  # Dependency linter CLI tool
├── internal/
│   ├── app/
│   │   ├── core/
│   │   │   ├── config.go            # Config structs & loader
│   │   │   ├── container.go         # Dependency injection container
│   │   │   └── feature_flag.go      # Feature flag structs & loader
│   │   └── http/
│   │       ├── echo.go              # Echo server setup
│   │       ├── gin.go               # Gin server setup
│   │       ├── helpers.go           # Middleware helpers
│   │       └── routes.go            # Route definitions
│   ├── shared/                      # ⭐ NEW: Shared Kernel
│   │   ├── cache/
│   │   │   ├── cache.go             # Cache interface definition
│   │   │   ├── errors.go            # Cache error types
│   │   │   └── memory.go            # In-memory cache implementation
│   │   ├── context/
│   │   │   └── context.go           # Shared HTTP context interface
│   │   ├── errors/
│   │   │   ├── errors.go            # Domain error types
│   │   │   ├── http.go              # HTTP status mapping
│   │   │   └── validation.go        # Validation error helpers
│   │   ├── events/
│   │   │   ├── event.go             # EventBus interface & Event struct
│   │   │   ├── memory_bus.go        # In-memory EventBus implementation
│   │   │   └── errors.go            # Event-related errors
│   │   ├── uow/
│   │   │   └── unit_of_work.go      # Unit of Work interface
│   │   └── validator/
│   │       └── validator.go         # Request validation utilities
│   ├── infrastructure/
│   │   ├── cache/
│   │   │   └── redis.go             # Redis cache implementation
│   │   ├── db/
│   │   │   ├── sql/
│   │   │   │   ├── sql.go           # PostgreSQL connection
│   │   │   │   └── migration/       # SQL migration files
│   │   │   ├── mongo/
│   │   │   │   ├── mongo.go         # MongoDB connection
│   │   │   │   └── migration/       # MongoDB migration scripts
│   │   │   └── uow/
│   │   │       └── unit_of_work.go  # UoW implementation
│   │   ├── cache/
│   │   │   └── redis.go             # Redis cache implementation
│   │   ├── logger/                  # Logger infrastructure
│   │   └── storage/                 # File storage (planned)
│   ├── modules/
│   │   ├── auth/
│   │   │   ├── domain/              # ⭐ NEW: Module-specific domain
│   │   │   │   ├── model.go         # Auth entities
│   │   │   │   ├── interfaces.go    # Handler/Service/Repo/ACL interfaces
│   │   │   │   ├── request.go       # Request DTOs
│   │   │   │   ├── response.go      # Response DTOs
│   │   │   │   └── events.go        # Domain events
│   │   │   ├── acl/                 # ⭐ NEW: Anti-Corruption Layer
│   │   │   │   └── user_creator.go  # ACL adapter for user module
│   │   │   ├── handler/
│   │   │   │   ├── v1/              # v1 handler implementation
│   │   │   │   └── noop/            # No-op/disabled handler
│   │   │   ├── middleware/          # Auth middleware (JWT, Session, Basic)
│   │   │   ├── service/
│   │   │   │   ├── v1/              # v1 service implementation
│   │   │   │   └── noop/            # No-op/disabled service
│   │   │   └── repository/
│   │   │       ├── sql/             # PostgreSQL repository
│   │   │       ├── mongo/           # MongoDB repository
│   │   │       └── noop/            # No-op repository
│   │   ├── product/
│   │   │   ├── domain/              # ⭐ NEW: Module-specific domain
│   │   │   │   ├── model.go
│   │   │   │   ├── interfaces.go
│   │   │   │   ├── request.go
│   │   │   │   ├── response.go
│   │   │   │   └── events.go
│   │   │   ├── handler/
│   │   │   │   ├── v1/
│   │   │   │   └── noop/
│   │   │   ├── service/
│   │   │   │   ├── v1/
│   │   │   │   └── noop/
│   │   │   └── repository/
│   │   │       ├── sql/
│   │   │       ├── mongo/
│   │   │       └── noop/
│   │   ├── user/
│   │   │   ├── domain/              # ⭐ NEW: Module-specific domain
│   │   │   │   ├── model.go
│   │   │   │   ├── interfaces.go
│   │   │   │   ├── request.go
│   │   │   │   ├── response.go
│   │   │   │   └── events.go
│   │   │   ├── handler/v1/
│   │   │   ├── service/v1/
│   │   │   └── repository/sql/
│   │   └── unitofwork/              # Unit of Work implementations
│   │       ├── default.unitofwork.go
│   │       ├── sql.unitofwork.go
│   │       └── mongo.unitofwork.go
│   ├── transports/
│   │   └── http/
│   │       ├── echo/
│   │       │   ├── adapter.echo.go
│   │       │   └── context.echo.go
│   │       ├── gin/
│   │       │   ├── adapter.gin.go
│   │       │   └── context.gin.go
│   │       ├── nethttp/                       # native net/http adapters
│   │       │   ├── adapter.nethttp.go
│   │       │   └── context.nethttp.go
│   │       ├── fasthttp/                      # fasthttp adapters (github.com/valyala/fasthttp)
│   │       │   ├── adapter.fasthttp.go
│   │       │   └── context.fasthttp.go
│   │       └── fiber/                         # Fiber adapters (github.com/gofiber/fiber)
│   │           ├── adapter.fiber.go
│   │           └── context.fiber.go
│   └── proto/                       # gRPC protobuf definitions (planned)
└── pkg/
    ├── constant/                    # Shared constants
    ├── logger/
    │   └── logger.go                # Shared logger utilities
    ├── model/
    │   ├── request.go               # Common request models
    │   └── response.go              # Common response models
    ├── routes/
    │   └── route.go                 # Route struct definition
    └── util/
        └── context.util.go          # Context utilities
```

---

## Getting Started

### Prerequisites

- Go 1.24.7+
- PostgreSQL 14+
- MongoDB 6.0+ (optional)
- mongosh CLI (for MongoDB migrations)

### Installation

```bash
# Clone the repository
git clone https://github.com/kamil5b/go-modular-monolith.git
cd go-modular-monolith

# Install dependencies
go mod tidy

# Configure database connection
# Edit config/config.yaml with your database credentials
```

### Running the Application

```bash
# Default: Run migrations and start server
go run .

# Run only the server (skip migrations)
go run . server

# Run SQL migrations
go run . migration sql up
go run . migration sql down

# Run MongoDB migrations
go run . migration mongo up
```

### CLI Commands

| Command | Description |
|---------|-------------|
| `go run .` | Run SQL migrations (up) then start server |
| `go run . server` | Start server only |
| `go run . migration sql up` | Apply SQL migrations |
| `go run . migration sql down` | Rollback SQL migrations |
| `go run . migration mongo up` | Apply MongoDB migrations |

---

## Configuration

### config/config.yaml

```yaml
environment: development  # development | production

app:
  server:
    port: "8080"          # HTTP server port
    grpc_port: "9090"     # gRPC server port (planned)

  database:
    sql:
      db_url: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
    mongo:
      mongo_url: "mongodb://localhost:27017"
      mongo_db: "myapp_db"

  redis:
    host: "localhost"
    port: "6379"
    password: ""
    db: 0
    max_retries: 3
    pool_size: 10
    min_idle_conns: 5

  jwt:
    secret: "supersecretkey"  # JWT signing secret
    access_token_duration: "15m"
    refresh_token_duration: "168h"

  auth:
    type: "jwt"  # jwt, session, basic, none
    session_cookie: "session_token"
    bcrypt_cost: 10
```

### Configuration Struct

```go
type Config struct {
    Environment string    `yaml:"environment"`
    App         AppConfig `yaml:"app"`
}

type AppConfig struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    JWT      JWTConfig      `yaml:"jwt"`
    Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
    Port     string `yaml:"port"`
    GRPCPort string `yaml:"grpc_port"`
}

type DatabaseConfig struct {
    SQL   SQLConfig   `yaml:"sql"`
    Mongo MongoConfig `yaml:"mongo"`
}

type RedisConfig struct {
    Host         string `yaml:"host"`
    Port         string `yaml:"port"`
    Password     string `yaml:"password"`
    DB           int    `yaml:"db"`
    MaxRetries   int    `yaml:"max_retries"`
    PoolSize     int    `yaml:"pool_size"`
    MinIdleConns int    `yaml:"min_idle_conns"`
}
```

---

## Feature Flags

Feature flags allow dynamic component selection without code changes.

### config/featureflags.yaml

```yaml
http_handler: echo  # echo | gin | nethttp | fasthttp | fiber

cache: redis  # redis | memory | disable

handler:
  authentication: v1   # v1 | disable
  product: v1          # v1 | disable
  user: v1             # v1 | disable

service:
  authentication: v1   # v1 | disable
  product: v1          # v1 | disable
  user: v1             # v1 | disable

repository:
  authentication: postgres  # postgres | mongo | disable
  product: postgres         # postgres | mongo | disable
  user: postgres            # postgres | mongo | disable
```

### Feature Flag Options

| Component | Options | Description |
|-----------|---------|-------------|
| `http_handler` | `echo`, `gin`, `nethttp`, `fasthttp`, `fiber` | HTTP framework selection |
| `cache` | `redis`, `memory`, `disable` | Cache backend (redis or in-memory) |
| `handler.*` | `v1`, `disable` | Handler version or disabled |
| `service.*` | `v1`, `disable` | Service version or disabled |
| `repository.*` | `postgres`, `mongo`, `disable` | Database backend |

### How It Works

The `Container` in `internal/app/core/container.go` reads feature flags and instantiates the appropriate implementations:

```go
// Repository selection example
switch featureFlag.Repository.Product {
case "mongo":
    productRepository = repoMongo.NewMongoRepository(mongo, "appdb")
case "postgres":
    productRepository = repoSQL.NewSQLRepository(db)
default:
    // No-op or disabled
}
```

---

## Modules

### Module Structure

Each business module follows a **domain-per-module** pattern with isolated domain types:

```
modules/<module>/
├── domain/                              # Module's private domain (NEW)
│   ├── model.go                         # Domain entities
│   ├── interfaces.go                    # Handler, Service, Repository interfaces
│   ├── request.go                       # Request DTOs
│   ├── response.go                      # Response DTOs
│   └── events.go                        # Domain events
├── acl/                                 # Anti-Corruption Layer (if needed)
│   └── <external>_<adapter>.go          # Adapters for external module dependencies
├── handler/
│   ├── v1/handler_v1.<module>.go        # Version 1 implementation
│   └── noop/handler_noop.<module>.go    # No-op implementation
├── service/
│   ├── v1/service_v1.<module>.go        # Version 1 implementation
│   └── noop/service_noop.<module>.go    # No-op implementation
└── repository/
    ├── sql/repository_sql.<module>.go   # PostgreSQL implementation
    ├── mongo/repository_mongo.<module>.go # MongoDB implementation
    └── noop/repository_noop.<module>.go  # No-op implementation
```

### Domain-Per-Module Pattern

Each module owns its domain types, preventing cross-module coupling:

```go
// internal/modules/product/domain/interfaces.go
package domain

type ProductHandler interface {
    Create(c sharedctx.Context) error
    Get(c sharedctx.Context) error
    List(c sharedctx.Context) error
    Update(c sharedctx.Context) error
    Delete(c sharedctx.Context) error
}
```

**Benefits:**
- **Module Isolation**: Each module can evolve independently
- **No Cyclic Dependencies**: Modules don't import each other's domain packages
- **Microservice-Ready**: Each module can be extracted as a separate service
- **Clear Ownership**: Domain types belong to the module that uses them

### Current Modules

#### Product Module
- **Status:** ✅ Complete
- **Features:** CRUD operations
- **Repository:** PostgreSQL, MongoDB

#### User Module
- **Status:** ✅ Complete  
- **Features:** CRUD operations
- **Repository:** PostgreSQL

#### Auth Module
- **Status:** ✅ Complete (untested)
- **Features:** JWT authentication, session management, Basic Auth, middleware
- **Repository:** PostgreSQL, MongoDB

---

## Shared Kernel

The shared kernel (`internal/shared/`) contains cross-cutting concerns that all modules can depend on. This is the **only** package that modules are allowed to import from outside their own domain.

### Package Structure

```
internal/shared/
├── context/
│   └── context.go       # Framework-agnostic HTTP context interface
├── errors/
│   ├── errors.go        # Domain error types
│   ├── validation.go    # Validation error handling
│   └── http.go          # HTTP status code mapping
├── events/
│   ├── event.go         # Event and EventBus interfaces
│   ├── memory_bus.go    # In-memory EventBus implementation
│   └── errors.go        # Event-related errors
├── uow/
│   └── unit_of_work.go  # Unit of Work interface
└── validator/
    └── validator.go     # Request validation utilities
```

### Shared Context (`sharedctx`)

The `sharedctx.Context` interface provides a framework-agnostic abstraction for HTTP handlers:

```go
// internal/shared/context/context.go
package sharedctx

type Context interface {
    // Request binding
    BindJSON(obj any) error
    BindURI(obj any) error
    BindQuery(obj any) error
    BindHeader(obj any) error
    Bind(obj any) error

    // Response
    JSON(code int, v any) error

    // Parameters
    Param(name string) string
    GetUserID() string
    Get(key string) any
    GetContext() context.Context

    // Auth-specific (for middleware)
    Set(key string, value any)
    GetHeader(key string) string
    SetHeader(key, value string)
    GetCookie(name string) (string, error)
    SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
    RemoveCookie(name string)
    GetClientIP() string
    GetUserAgent() string
}
```

**Usage in Handlers:**
```go
// All modules use sharedctx.Context instead of per-module Context
type ProductHandler interface {
    Create(c sharedctx.Context) error
    Get(c sharedctx.Context) error
}
```

### Domain Errors

Structured error types for consistent error handling:

```go
// internal/shared/errors/errors.go
package sharederrors

type ErrorType string

const (
    ErrTypeNotFound       ErrorType = "NOT_FOUND"
    ErrTypeValidation     ErrorType = "VALIDATION"
    ErrTypeUnauthorized   ErrorType = "UNAUTHORIZED"
    ErrTypeForbidden      ErrorType = "FORBIDDEN"
    ErrTypeConflict       ErrorType = "CONFLICT"
    ErrTypeInternal       ErrorType = "INTERNAL"
)

type DomainError struct {
    Type    ErrorType
    Message string
    Err     error
}

// Helper constructors
func NotFound(resource, id string) *DomainError
func Validation(message string) *DomainError
func Unauthorized(message string) *DomainError
func Forbidden(message string) *DomainError
func Conflict(message string) *DomainError
func Internal(err error) *DomainError
```

### Event Bus

Asynchronous inter-module communication without direct dependencies:

```go
// internal/shared/events/event.go
package events

type Event interface {
    EventName() string
    OccurredAt() time.Time
}

type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventName string, handler EventHandler) error
}

type EventHandler func(ctx context.Context, event Event) error
```

**In-Memory Implementation:**
```go
// internal/shared/events/memory_bus.go
type InMemoryEventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus
func (b *InMemoryEventBus) Publish(ctx context.Context, event Event) error
func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) error
```

**Domain Events Example:**
```go
// internal/modules/user/domain/events.go
type UserCreated struct {
    UserID    string
    Email     string
    Name      string
    Timestamp time.Time
}

func (e UserCreated) EventName() string    { return "user.created" }
func (e UserCreated) OccurredAt() time.Time { return e.Timestamp }
```

### Caching Layer

The caching layer (`internal/shared/cache/`) provides a unified interface for cache operations across all modules, supporting both Redis and in-memory implementations.

**Cache Interface:**
```go
// internal/shared/cache/cache.go
package cache

type Cache interface {
	// Get retrieves a string value from cache
	Get(ctx context.Context, key string) (string, error)

	// GetBytes retrieves bytes value from cache
	GetBytes(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with optional expiration
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// SetNX stores a value only if key does not exist
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)

	// Delete removes one or more keys from cache
	Delete(ctx context.Context, keys ...string) error

	// Exists checks if one or more keys exist in cache (returns count of existing keys)
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets a timeout on a key
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// TTL returns the remaining time to live of a key
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Increment increments the number stored at key
	Increment(ctx context.Context, key string, increment int64) (int64, error)

	// Decrement decrements the number stored at key
	Decrement(ctx context.Context, key string, decrement int64) (int64, error)

	// Health checks the health of the cache
	Health(ctx context.Context) error
}
```

**Implementations:**

1. **Redis Cache** (`internal/infrastructure/cache/redis.go`):
   - Production-grade caching with connection pooling
   - Configurable timeouts and pool sizes
   - Automatic health checks
   - Fallback to in-memory cache on connection failure

2. **In-Memory Cache** (`internal/shared/cache/memory.go`):
   - Ideal for testing and development
   - Automatic expiration cleanup
   - Goroutine-safe with RWMutex

**Usage in Modules:**
```go
// Services can inject Cache to cache computed results
type ProductService struct {
    repository productDomain.Repository
    cache      cache.Cache
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*productDomain.Product, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("product:%s", id)
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil {
        // Parse and return cached product
        return parseProduct(cached)
    }

    // Cache miss or error, fetch from repository
    product, err := s.repository.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache the result (1 hour expiration)
    s.cache.Set(ctx, cacheKey, product.String(), time.Hour)
    return product, nil
}
```

---

## Anti-Corruption Layer (ACL)

When a module needs to interact with another module, it uses an **Anti-Corruption Layer** to maintain isolation. The ACL translates between module boundaries without creating direct dependencies.

### Pattern Overview

```
┌─────────────────────────────────────────────────────────────┐
│ Auth Module                                                  │
│                                                              │
│  ┌──────────┐    ┌─────────────────┐    ┌───────────────┐  │
│  │ Service  │───>│ UserCreator     │───>│ ACL Adapter   │──┼──> User Repository
│  │          │    │ (interface)     │    │               │  │
│  └──────────┘    └─────────────────┘    └───────────────┘  │
│                                                              │
│  Auth module defines the interface it needs.                │
│  ACL adapter wraps the actual implementation.               │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Example

**1. Module defines its interface (what it needs):**
```go
// internal/modules/auth/domain/interfaces.go
package domain

// UserCreator is an ACL interface for creating users during registration
// This decouples auth from direct user module dependency
type UserCreator interface {
    CreateUser(ctx context.Context, name, email string) (string, error)
}
```

**2. ACL adapter implements the interface:**
```go
// internal/modules/auth/acl/user_creator.go
package acl

import (
    userdomain "go-modular-monolith/internal/modules/user/domain"
)

// UserCreatorAdapter adapts the user repository to auth's UserCreator interface
type UserCreatorAdapter struct {
    userRepo userdomain.UserRepository
}

func NewUserCreatorAdapter(userRepo userdomain.UserRepository) *UserCreatorAdapter {
    return &UserCreatorAdapter{userRepo: userRepo}
}

func (a *UserCreatorAdapter) CreateUser(ctx context.Context, name, email string) (string, error) {
    user := &userdomain.User{
        ID:        generateID(),
        Name:      name,
        Email:     email,
        CreatedAt: time.Now(),
        CreatedBy: "system",
    }
    if err := a.userRepo.Create(ctx, user); err != nil {
        return "", err
    }
    return user.ID, nil
}
```

**3. Container wires the ACL:**
```go
// internal/app/core/container.go
import (
    authacl "go-modular-monolith/internal/modules/auth/acl"
)

// Create ACL adapter
userCreator := authacl.NewUserCreatorAdapter(userRepository)

// Inject into auth service
authService = authServiceV1.NewService(
    authRepository,
    userCreator, // ACL adapter, not direct user repository
    sessionDuration,
    jwtSecret,
)
```

### ACL vs Events

| Aspect | ACL | Events |
|--------|-----|--------|
| **Communication** | Synchronous | Asynchronous |
| **Consistency** | Strong (transactional) | Eventual |
| **Coupling** | Interface-level | Event contract |
| **Use Case** | Registration (need user ID immediately) | Notifications, analytics |
| **Testing** | Easy to mock | Requires event bus setup |

### When to Use ACL

✅ **Use ACL when:**
- You need synchronous, transactional operations
- The calling module needs an immediate response
- You want explicit, type-safe interfaces
- Testing with mocks is important

❌ **Use Events when:**
- Operations can be eventually consistent
- Multiple modules need to react to the same event
- You want fire-and-forget behavior
- Modules should be completely decoupled

---

## Dependency Linter

The project includes a custom dependency linter (`cmd/lint-deps/main.go`) that enforces module isolation rules.

### Running the Linter

```bash
go run cmd/lint-deps/main.go
```

### Rules Enforced

1. **No Cross-Module Imports**: Modules cannot import from other modules' packages
2. **ACL Exception**: `acl/` folders are allowed to import other modules (they're the translation layer)
3. **Shared Kernel Allowed**: All modules can import from `internal/shared/`

### Example Output

```
✅ Checking dependencies...

❌ Violation in internal/modules/auth/service/v1/service_v1.auth.go:
   - Imports "go-modular-monolith/internal/modules/user/repository/sql"
   - Module "auth" should not import from module "user"

Fix: Use an ACL adapter or events for cross-module communication.
```

When properly configured:
```
✅ Checking dependencies...
✅ No cross-module dependency violations found!
```

### Linter Configuration

The linter automatically:
- Scans all `.go` files in `internal/modules/`
- Identifies module boundaries by folder names
- Allows imports from `internal/shared/`
- Allows imports within ACL folders to other modules

---

## Domain Layer

### Shared Context Interface

All modules use `sharedctx.Context` from the shared kernel for framework-agnostic HTTP handling:

```go
import sharedctx "go-modular-monolith/internal/shared/context"

type ProductHandler interface {
    Create(c sharedctx.Context) error
    Get(c sharedctx.Context) error
    List(c sharedctx.Context) error
    Update(c sharedctx.Context) error
    Delete(c sharedctx.Context) error
}
```

---

## Microservices Readiness

### Current Readiness Score: 8/10

The architecture has been significantly improved to support future microservices migration.

### Readiness Assessment

| Aspect | Score | Status | Notes |
|--------|-------|--------|-------|
| **Module Isolation** | ✅ 10/10 | Complete | Domain-per-module, no cross-module imports |
| **Dependency Direction** | ✅ 9/10 | Complete | ACL pattern, dependency linter enforced |
| **Database per Module** | 🟡 7/10 | Partial | Shared DB, but separate tables per module |
| **API Contracts** | ✅ 8/10 | Good | Clean request/response DTOs per module |
| **Configuration** | ✅ 8/10 | Good | Feature flags support module-level config |
| **Event-Driven** | 🟡 7/10 | Partial | EventBus ready, not fully utilized |
| **Testing** | 🔴 4/10 | Needs Work | Unit tests not implemented yet |

### Migration Path

When ready to migrate to microservices:

```
┌─────────────────────────────────────────────────────────────────┐
│                    CURRENT: Modular Monolith                    │
│                                                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                     │
│  │ Product  │  │   User   │  │   Auth   │  (Shared Process)   │
│  │ Module   │  │  Module  │  │  Module  │                     │
│  └──────────┘  └──────────┘  └──────────┘                     │
│        │            │             │                            │
│        └────────────┴─────────────┘                            │
│                     │                                          │
│              Shared Database                                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FUTURE: Microservices                        │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Product Svc  │  │  User Svc    │  │  Auth Svc    │         │
│  │              │  │              │  │              │         │
│  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │         │
│  │ │Product DB│ │  │ │ User DB  │ │  │ │ Auth DB  │ │         │
│  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│         │                 │                 │                  │
│         └─────────────────┴─────────────────┘                  │
│                           │                                    │
│                    Message Queue / API Gateway                 │
└─────────────────────────────────────────────────────────────────┘
```

### Step-by-Step Migration

1. **Extract Module as Service**
   ```bash
   # Each module folder becomes its own service
   internal/modules/product/ → product-service/
   ```

2. **Replace ACL with HTTP/gRPC**
   ```go
   // Before (ACL in monolith)
   type UserCreatorAdapter struct {
       userRepo userdomain.UserRepository
   }
   
   // After (HTTP client in microservice)
   type UserServiceClient struct {
       baseURL string
       client  *http.Client
   }
   ```

3. **Replace In-Memory EventBus**
   ```go
   // Before (in-memory)
   bus := events.NewInMemoryEventBus()
   
   // After (distributed)
   bus := events.NewKafkaEventBus(brokers)
   // or
   bus := events.NewRabbitMQEventBus(amqpURL)
   ```

4. **Database per Service**
   - Each service gets its own database
   - Use database migrations from `internal/infrastructure/db/*/migration/`

### What's Already Microservice-Ready

✅ **Module Isolation**: Each module is self-contained with its own domain  
✅ **ACL Pattern**: Cross-module communication via adapters (easy to swap for HTTP clients)  
✅ **Event Bus Interface**: In-memory implementation can be swapped for Kafka/RabbitMQ  
✅ **Feature Flags**: Enable/disable modules independently  
✅ **Repository Pattern**: Database access abstracted behind interfaces  
✅ **Dependency Linter**: Enforces clean boundaries  

### What Needs Work for Microservices

🔴 **Database Separation**: Currently shared DB, need per-module schemas  
🔴 **Distributed Tracing**: Add OpenTelemetry for cross-service tracing  
🔴 **API Gateway**: Need gateway for routing and aggregation  
🔴 **Service Discovery**: Add Consul/etcd for service registration  
🔴 **Circuit Breakers**: Add resilience patterns for inter-service calls  

---

## Roadmap

### Completed ✅
- [x] Architecture & Infrastructure setup
- [x] Product CRUD: HTTP Echo → v1 → SQL repository
- [x] SQL & MongoDB repository support with migrations
- [x] Gin framework integration
- [x] Request & Response models
- [x] User CRUD implementation
- [x] Authentication system: JWT, Basic Auth, Session-based (untested)
- [x] Middleware integration (untested)
- [x] **Shared Kernel** (`internal/shared/`) - Events, Errors, Context, UoW, Validator
- [x] **Domain-per-Module Pattern** - Each module owns its domain types
- [x] **Anti-Corruption Layer (ACL)** - Clean cross-module communication
- [x] **Dependency Linter** (`cmd/lint-deps/`) - Enforces module isolation
- [x] **Shared Context Interface** (`sharedctx.Context`) - Framework-agnostic handlers
- [x] **Redis Integration** - Caching with Redis & in-memory fallback

### Planned 📋
- [ ] Unit Tests (Priority: High)
- [ ] Worker support: Asynq, RabbitMQ, Redpanda
- [ ] Storage support: S3-Compatible, GCS, MinIO, Local
- [ ] gRPC & Protocol Buffers support
- [ ] WebSocket integration
- [ ] OpenTelemetry integration for distributed tracing
- [ ] Database-per-module schema separation
- [ ] API Gateway setup (Kong/Traefik)
- [ ] Kubernetes deployment manifests

---

## Contributing

### Development Guidelines

1. **Follow Module Structure**: Use `domain/` folder for module-specific types
2. **Use Feature Flags**: Configure new components via `config/featureflags.yaml`
3. **Implement Multiple Repositories**: Support PostgreSQL and MongoDB when applicable
4. **Add Migrations**: Place in `internal/infrastructure/db/*/migration/`
5. **Update Documentation**: Keep this file current with changes

### Service Layer Best Practices

When implementing service methods that use Unit of Work transactions, follow the **Panic Recovery Pattern** with defer to ensure proper context cleanup:

```go
func (s *ServiceV1) Create(ctx context.Context, req *domain.CreateRequest, createdBy string) (result *domain.Entity, err error) {
	// Start transaction context
	ctx = s.uow.StartContext(ctx)
	
	// Defer cleanup with error and panic recovery
	defer s.uow.DeferErrorContext(ctx, err)
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("panic: %s", x)
			case error:
				err = fmt.Errorf("panic: %w", x)
			default:
				err = fmt.Errorf("panic: %v", x)
			}
		}
	}()

	// Your business logic here
	entity := &domain.Entity{
		Name: req.Name,
		CreatedAt: time.Now().UTC(),
		CreatedBy: createdBy,
	}
	
	if err = s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	// Publish domain events
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.EntityCreatedEvent{
			EntityID:  entity.ID,
			CreatedBy: createdBy,
		})
	}

	result = entity
	return
}
```

**Key Points:**
- Use **named return values** (`result *domain.Entity, err error`) for clean deferred cleanup
- First defer handles UnitOfWork cleanup and error propagation
- Second defer handles panic recovery, converting panics to errors
- Always check errors and return early
- Publish domain events after successful operations
- Return early from defers so cleanup always happens at function exit

**Benefits:**
- ✅ Guaranteed context cleanup (commit on success, rollback on error)
- ✅ Panic safety - panics are caught and converted to errors
- ✅ Transaction atomicity - all-or-nothing semantics
- ✅ Cleaner code - error handling and cleanup in one place
- ✅ Consistent pattern across all services

### Dependency Rules

1. **Run Linter Before Commit**:
   ```bash
   go run cmd/lint-deps/main.go
   ```

2. **Cross-Module Communication**:
   - Use **ACL** for synchronous operations (create ACL adapter in `acl/` folder)
   - Use **Events** for asynchronous operations (publish to EventBus)

3. **Allowed Imports**:
   - ✅ `internal/shared/*` - Shared kernel packages
   - ✅ Same module packages - `internal/modules/<module>/*`
   - ✅ ACL folders can import other modules - `internal/modules/<module>/acl/`
   - ❌ Cross-module imports - `internal/modules/<other-module>/*`

### Adding a New Module

1. Create folder structure:
   ```
   internal/modules/<new-module>/
   ├── domain/
   │   ├── model.go
   │   ├── interfaces.go
   │   ├── request.go
   │   ├── response.go
   │   └── events.go
   ├── handler/v1/
   ├── service/v1/
   └── repository/sql/
   ```

2. Add feature flags to `config/featureflags.yaml`
3. Wire up in `internal/app/core/container.go`
4. Add routes in `internal/app/http/routes.go`
5. Run dependency linter to verify isolation

---

## License

See [LICENSE](../LICENSE) file for details.
