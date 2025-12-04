# Go-PSTE-Monolith - Technical Documentation

> **Version:** 2.1.0  
> **Last Updated:** December 4, 2025  
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
16. [Worker Support](#worker-support)
17. [Development Guide](#development-guide)
18. [Dependency Linter](#dependency-linter)
19. [Microservices Readiness](#microservices-readiness)
20. [Roadmap](#roadmap)

---

## Overview

Go-PSTE-Monolith is a production-ready, modular monolithic application built with Go. It implements clean architecture principles with pluggable components, allowing teams to:

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
| Logging | Logrus (sirupsen/logrus) |
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
github.com/kamil5b/go-ptse-monolith/
├── main.go                          # Application entry point
├── go.mod                           # Go module definition
├── config/
│   ├── config.yaml                  # Application configuration
│   └── featureflags.yaml            # Feature flag configuration
├── cmd/
│   ├── bootstrap/
│   │   ├── bootstrap.server.go      # Server initialization
│   │   ├── bootstrap.migration.go   # Migration runner
│   │   └── bootstrap.worker.go      # Worker initialization
│   └── lint-deps/
│       └── main.go                  # Dependency linter CLI tool
├── internal/
│   ├── app/
│   │   ├── core/
│   │   │   ├── config.go            # Config structs & loader
│   │   │   ├── container.go         # Dependency injection container
│   │   │   └── feature_flag.go      # Feature flag structs & loader
│   │   ├── http/
│   │   │   ├── echo.go              # Echo server setup
│   │   │   ├── gin.go               # Gin server setup
│   │   │   ├── fiber.go             # Fiber server setup
│   │   │   ├── fasthttp.go          # FastHTTP server setup
│   │   │   ├── nethttp.go           # net/http server setup
│   │   │   ├── helpers.go           # Middleware helpers
│   │   │   └── routes.go            # Route definitions
│   │   └── worker/
│   │       ├── manager.go           # Worker manager
│   │       └── registrar.go         # Module task registrar
│   ├── shared/                      # Shared Kernel
│   │   ├── cache/
│   │   │   ├── cache.go             # Cache interface definition
│   │   │   ├── errors.go            # Cache error types
│   │   │   ├── memory.go            # In-memory cache implementation
│   │   │   └── mocks/               # Cache mocks for testing
│   │   ├── context/
│   │   │   ├── context.go           # Shared HTTP context interface
│   │   │   ├── context_key.go       # Context key definitions
│   │   │   ├── util.go              # Context utilities
│   │   │   └── mocks/               # Context mocks for testing
│   │   ├── email/
│   │   │   ├── email.go             # Email service interface & types
│   │   │   ├── noop.go              # NoOp email service
│   │   │   └── mocks/               # Email mocks for testing
│   │   ├── errors/
│   │   │   ├── errors.go            # Domain error types
│   │   │   ├── http.go              # HTTP status mapping
│   │   │   └── validation.go        # Validation error helpers
│   │   ├── events/
│   │   │   ├── event.go             # EventBus interface & Event struct
│   │   │   ├── memory_bus.go        # In-memory EventBus implementation
│   │   │   ├── errors.go            # Event-related errors
│   │   │   └── mocks/               # Event bus mocks for testing
│   │   ├── model/
│   │   │   ├── request.go           # Common request models
│   │   │   └── response.go          # Common response models
│   │   ├── storage/
│   │   │   ├── storage.go           # Storage service interface
│   │   │   ├── errors.go            # Storage error types
│   │   │   └── mocks/               # Storage mocks for testing
│   │   ├── uow/
│   │   │   └── unit_of_work.go      # Unit of Work interface
│   │   ├── validator/
│   │   │   └── validator.go         # Request validation utilities
│   │   └── worker/
│   │       ├── worker.go            # TaskPayload, TaskHandler, Client, Server, etc.
│   │       └── mocks/               # Worker mocks for testing
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
│   │   ├── email/                   # Email providers
│   │   │   ├── smtp/                # SMTP email service
│   │   │   ├── mailgun/             # Mailgun email service
│   │   │   └── template/            # Email template loader
│   │   ├── storage/                 # File storage
│   │   │   ├── local/               # Local filesystem storage
│   │   │   ├── s3/                  # AWS S3 & S3-compatible storage
│   │   │   ├── gcs/                 # Google Cloud Storage
│   │   │   └── noop/                # NoOp storage for testing
│   │   └── worker/                  # Worker implementations
│   │       ├── asynq/               # Asynq (Redis) worker
│   │       ├── rabbitmq/            # RabbitMQ worker
│   │       ├── redpanda/            # Redpanda/Kafka worker
│   │       ├── cron_scheduler.go    # Cron job scheduler
│   │       ├── retry_policy.go      # Retry policy utilities
│   │       └── noop.go              # NoOp worker for testing
│   ├── modules/
│   │   ├── auth/
│   │   │   ├── domain/              # Module-specific domain
│   │   │   │   ├── model.go         # Auth entities
│   │   │   │   ├── interfaces.go    # Handler/Service/Repo/ACL interfaces
│   │   │   │   ├── request.go       # Request DTOs
│   │   │   │   ├── response.go      # Response DTOs
│   │   │   │   └── events.go        # Domain events
│   │   │   ├── acl/                 # Anti-Corruption Layer
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
│   │   │   ├── domain/              # Module-specific domain
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
│   │   │   ├── domain/              # ⭐ Module-specific domain
│   │   │   │   ├── model.go
│   │   │   │   ├── interfaces.go
│   │   │   │   ├── request.go
│   │   │   │   ├── response.go
│   │   │   │   └── events.go
│   │   │   ├── handler/v1/
│   │   │   ├── service/v1/
│   │   │   ├── repository/sql/
│   │   │   └── worker/              # User module worker tasks
│   │   │       ├── tasks.go         # Task definitions & payloads
│   │   │       ├── handlers.go      # Task handlers
│   │   │       └── registrar.go     # Module task registrar
│   │   └── unitofwork/              # Unit of Work implementations
│   │       ├── default.unitofwork.go
│   │       ├── sql.unitofwork.go
│   │       └── mongo.unitofwork.go
│   ├── transports/
│   │   ├── http/
│   │   │   ├── route.go                       # Route type definition
│   │   │   ├── echo/
│   │   │   │   ├── adapter.echo.go
│   │   │   │   └── context.echo.go
│   │   │   ├── gin/
│   │   │   │   ├── adapter.gin.go
│   │   │   │   └── context.gin.go
│   │   │   ├── nethttp/                       # native net/http adapters
│   │   │   │   ├── adapter.nethttp.go
│   │   │   │   └── context.nethttp.go
│   │   │   ├── fasthttp/                      # fasthttp adapters (github.com/valyala/fasthttp)
│   │   │   │   ├── adapter.fasthttp.go
│   │   │   │   └── context.fasthttp.go
│   │   │   └── fiber/                         # Fiber adapters (github.com/gofiber/fiber)
│   │   │       ├── adapter.fiber.go
│   │   │       └── context.fiber.go
│   │   ├── grpc/                              # gRPC transport (planned)
│   │   │   └── .gitkeep
│   │   └── websocket/                         # WebSocket transport (planned)
│   │       └── .gitkeep
│   └── proto/                                 # gRPC protobuf definitions (planned)
│       └── .gitkeep
└── scripts/                             # Utility scripts
    ├── db.sh
    ├── dev.sh
    ├── setup.sh
    ├── health-check.sh
    ├── new-module.sh
    ├── generate_mocks_from_source.sh
    └── lint-deps-check.sh
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
git clone https://github.com/kamil5b/go-ptse-monolith.git
cd go-ptse-monolith

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
| `go run . worker` | Start worker server for async task processing |
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

  worker:
    enabled: false
    backend: "asynq"  # asynq, rabbitmq, redpanda, disable
    asynq:
      redis_url: "redis://localhost:6379"
      concurrency: 10
      max_retries: 3
      default_timeout: "300s"
    rabbitmq:
      url: "amqp://guest:guest@localhost:5672/"
      exchange: "tasks"
      queue: "tasks_queue"
      worker_count: 10
      prefetch_count: 1
    redpanda:
      brokers:
        - "localhost:9092"
      topic: "tasks"
      consumer_group: "workers"
      partition_count: 3
      replication_factor: 1
      worker_count: 10

  email:
    enabled: false
    provider: "noop"  # smtp, mailgun, noop
    smtp:
      host: "smtp.gmail.com"
      port: 587
      username: "your-email@gmail.com"
      password: "your-app-password"
      from_addr: "noreply@example.com"
      from_name: "MyApp"
    mailgun:
      domain: "mg.example.com"
      api_key: "key-xxxx"
      from_addr: "noreply@example.com"
      from_name: "MyApp"

  storage:
    enabled: false
    local:
      base_path: "./uploads"
      max_file_size: 104857600  # 100MB in bytes
      allow_public_access: false
      public_url: "http://localhost:8080/files"
    s3:
      region: "us-east-1"
      bucket: "my-bucket"
      access_key_id: "${AWS_ACCESS_KEY_ID}"
      secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
      endpoint: ""  # Leave empty for AWS, set for MinIO/Spaces
      use_ssl: true
      path_style: false  # Set to true for MinIO
      presigned_url_ttl: 3600
      server_side_encryption: false
      storage_class: "STANDARD"
    gcs:
      project_id: "my-project"
      bucket: "my-bucket"
      credentials_file: ""  # Path to service account JSON
      credentials_json: "${GCS_CREDENTIALS}"  # Or use environment variable
      storage_class: "STANDARD"
      location: "US"
      metadata_cache: true
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
    Worker   WorkerConfig   `yaml:"worker"`
    Email    EmailConfig    `yaml:"email"`
    Storage  StorageConfig  `yaml:"storage"`
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

type WorkerConfig struct {
    Enabled  bool                 `yaml:"enabled"`
    Backend  string               `yaml:"backend"` // asynq, rabbitmq, redpanda, disable
    Asynq    AsynqWorkerConfig    `yaml:"asynq"`
    RabbitMQ RabbitMQWorkerConfig `yaml:"rabbitmq"`
    Redpanda RedpandaWorkerConfig `yaml:"redpanda"`
}

type EmailConfig struct {
    Enabled  bool          `yaml:"enabled"`
    Provider string        `yaml:"provider"` // smtp, mailgun, noop
    SMTP     SMTPConfig    `yaml:"smtp"`
    Mailgun  MailgunConfig `yaml:"mailgun"`
}

type StorageConfig struct {
    Enabled bool               `yaml:"enabled"`
    Local   LocalStorageConfig `yaml:"local"`
    S3      S3StorageConfig    `yaml:"s3"`
    GCS     GCSStorageConfig   `yaml:"gcs"`
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

worker:
  enabled: false
  backend: disable  # asynq | rabbitmq | redpanda | disable
  tasks:
    email_notifications: false
    data_export: false
    report_generation: false
    image_processing: false

email:
  enabled: false
  provider: noop  # smtp | mailgun | noop

storage:
  enabled: false
  backend: noop  # local | s3 | gcs | s3-compatible | noop
  s3:
    enable_encryption: false
    storage_class: "STANDARD"
    presigned_url_ttl: 3600
  gcs:
    storage_class: "STANDARD"
    metadata_cache: true
```

### Feature Flag Options

| Component | Options | Description |
|-----------|---------|-------------|
| `http_handler` | `echo`, `gin`, `nethttp`, `fasthttp`, `fiber` | HTTP framework selection |
| `cache` | `redis`, `memory`, `disable` | Cache backend (redis or in-memory) |
| `handler.*` | `v1`, `disable` | Handler version or disabled |
| `service.*` | `v1`, `disable` | Service version or disabled |
| `repository.*` | `postgres`, `mongo`, `disable` | Database backend |
| `worker.enabled` | `true`, `false` | Enable/disable worker system |
| `worker.backend` | `asynq`, `rabbitmq`, `redpanda`, `disable` | Worker queue backend |
| `worker.tasks.*` | `true`, `false` | Enable/disable specific task types |
| `email.enabled` | `true`, `false` | Enable/disable email service |
| `email.provider` | `smtp`, `mailgun`, `noop` | Email provider selection |
| `storage.enabled` | `true`, `false` | Enable/disable storage service |
| `storage.backend` | `local`, `s3`, `gcs`, `s3-compatible`, `noop` | Storage backend selection |

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
- **Features:** CRUD operations, worker tasks (welcome emails, data export, reports)
- **Repository:** PostgreSQL
- **Workers:** Send welcome email, password reset, data export, monthly emails

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
├── cache/
│   ├── cache.go             # Cache interface definition
│   ├── errors.go            # Cache error types
│   ├── memory.go            # In-memory cache implementation
│   └── mocks/               # Cache mocks for testing
├── context/
│   ├── context.go           # Framework-agnostic HTTP context interface
│   ├── context_key.go       # Context key definitions
│   ├── util.go              # Context utilities
│   └── mocks/               # Context mocks for testing
├── email/
│   ├── email.go             # EmailService interface & types
│   ├── noop.go              # NoOp email service implementation
│   └── mocks/               # Email mocks for testing
├── errors/
│   ├── errors.go            # Domain error types
│   ├── validation.go        # Validation error handling
│   └── http.go              # HTTP status code mapping
├── events/
│   ├── event.go             # Event and EventBus interfaces
│   ├── memory_bus.go        # In-memory EventBus implementation
│   ├── errors.go            # Event-related errors
│   └── mocks/               # Event bus mocks for testing
├── model/
│   ├── request.go           # Common request models
│   └── response.go          # Common response models
├── storage/
│   ├── storage.go           # StorageService interface
│   ├── errors.go            # Storage error types
│   └── mocks/               # Storage mocks for testing
├── uow/
│   └── unit_of_work.go      # Unit of Work interface
├── validator/
│   └── validator.go         # Request validation utilities
└── worker/
    ├── worker.go            # TaskPayload, TaskHandler, Client, Server, Scheduler interfaces
    └── mocks/               # Worker mocks for testing
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
    userdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
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
    authacl "github.com/kamil5b/go-ptse-monolith/internal/modules/auth/acl"
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
   - Imports "github.com/kamil5b/go-ptse-monolith/internal/modules/user/repository/sql"
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
import sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"

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

### Current Readiness Score: 9/10

The architecture has been significantly improved to support future microservices migration.

### Readiness Assessment

| Aspect | Score | Status | Notes |
|--------|-------|--------|-------|
| **Module Isolation** | ✅ 10/10 | Complete | Domain-per-module, no cross-module imports |
| **Dependency Direction** | ✅ 9/10 | Complete | ACL pattern, dependency linter enforced |
| **Database per Module** | 🟡 7/10 | Partial | Shared DB, but separate tables per module |
| **API Contracts** | ✅ 8/10 | Good | Clean request/response DTOs per module |
| **Configuration** | ✅ 9/10 | Good | Feature flags support module-level config |
| **Event-Driven** | ✅ 8/10 | Good | EventBus ready with worker integration |
| **Async Processing** | ✅ 9/10 | Good | Workers with Asynq, RabbitMQ, Redpanda |
| **Testing** | 🟡 6/10 | Partial | Mock generation scripts, needs more tests |

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
✅ **Worker Infrastructure**: Asynq, RabbitMQ, Redpanda support for async processing  
✅ **Feature Flags**: Enable/disable modules independently  
✅ **Repository Pattern**: Database access abstracted behind interfaces  
✅ **Dependency Linter**: Enforces clean boundaries  
✅ **Storage Abstraction**: S3, GCS, local filesystem support  
✅ **Email Abstraction**: SMTP, Mailgun support with NoOp for testing  

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
- [x] **Shared Kernel** (`internal/shared/`) - Events, Errors, Context, UoW, Validator, Email, Storage, Worker
- [x] **Domain-per-Module Pattern** - Each module owns its domain types
- [x] **Anti-Corruption Layer (ACL)** - Clean cross-module communication
- [x] **Dependency Linter** (`cmd/lint-deps/`) - Enforces module isolation
- [x] **Shared Context Interface** (`sharedctx.Context`) - Framework-agnostic handlers
- [x] **Redis Integration** - Caching with Redis & in-memory fallback
- [x] **Worker Support** - Asynq, RabbitMQ, and Redpanda integration with cron scheduler
- [x] **Email Services** - SMTP and Mailgun providers with NoOp for testing
- [x] **Storage Services** - Local filesystem, AWS S3, S3-compatible (MinIO), and Google Cloud Storage
- [x] **Additional HTTP Frameworks** - net/http, FastHTTP, and Fiber support

### Planned 📋
- [ ] Unit Tests (Priority: High)
- [ ] gRPC & Protocol Buffers support
- [ ] WebSocket integration
- [ ] OpenTelemetry integration for distributed tracing
- [ ] Database-per-module schema separation
- [ ] API Gateway setup (Kong/Traefik)
- [ ] Kubernetes deployment manifests

---

## Worker Support

The application supports multiple task queue and worker backends for asynchronous job processing. Workers enable decoupling of long-running tasks from HTTP request/response cycles and support scheduled jobs, retries, and distributed processing.

### Supported Backends

| Backend | Use Case | Features | Production Ready |
|---------|----------|----------|------------------|
| **Asynq** | Task queue | Redis-backed, retry logic, scheduling, priority queues | ✅ Yes |
| **RabbitMQ** | Message broker | Advanced routing, persistent queues, dead-letter exchanges | ✅ Yes |
| **Redpanda** | Streaming platform | Kafka-compatible, high throughput, built-in retries | ✅ Yes |

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP Handler Layer                            │
│                  (HTTP Request/Response)                         │
└────────────────┬──────────────────────────────────────────────┘
                 │
                 │ Enqueue Task
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Task Queue Backend                             │
│   (Asynq / RabbitMQ / Redpanda)                                 │
│                                                                  │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│   │   Task A     │  │   Task B     │  │   Task C     │         │
│   │ (Priority 1) │  │ (Priority 2) │  │ (Scheduled)  │         │
│   └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                 │
                 │ Dequeue Tasks
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Worker Pool                                   │
│                                                                  │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐          │
│   │Worker 1 │  │Worker 2 │  │Worker N │  │Scheduler│          │
│   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘          │
└────────┼────────────┼────────────┼────────────┼────────────────┘
         │            │            │            │
         ▼            ▼            ▼            ▼
    Service Logic (Event Publishing, Database Updates, etc.)
```

### Project Structure

```
internal/
├── shared/
│   └── worker/
│       └── worker.go              # Shared types & interfaces (TaskPayload, TaskHandler,
│                                  # Client, Server, Scheduler, CronExpression, etc.)
├── infrastructure/
│   └── worker/
│       ├── noop.go                # No-op implementations
│       ├── cron_scheduler.go      # Cron scheduler implementation
│       ├── retry_policy.go        # Retry policy utilities
│       ├── asynq/
│       │   ├── client.go          # Asynq task client
│       │   └── server.go          # Asynq worker server
│       ├── rabbitmq/
│       │   ├── client.go          # RabbitMQ producer
│       │   └── server.go          # RabbitMQ consumer
│       └── redpanda/
│           ├── client.go          # Redpanda producer
│           └── server.go          # Redpanda consumer
└── modules/
    └── <module>/
        └── worker/
            ├── tasks.go           # Task definitions
            ├── handlers.go        # Task handlers
            └── registrar.go       # Module task registrar
```

### Configuration

### config/config.yaml

```yaml
app:
  worker:
    enabled: true
    backend: asynq  # asynq | rabbitmq | redpanda
    
    # Asynq configuration
    asynq:
      redis_url: "redis://localhost:6379"
      concurrency: 10
      max_retries: 3
      default_timeout: 300s  # 5 minutes
    
    # RabbitMQ configuration
    rabbitmq:
      url: "amqp://guest:guest@localhost:5672/"
      exchange: "tasks"
      queue: "tasks_queue"
      worker_count: 10
      prefetch_count: 1
    
    # Redpanda configuration
    redpanda:
      brokers:
        - "localhost:9092"
      topic: "tasks"
      consumer_group: "workers"
      partition_count: 3
      replication_factor: 1
      worker_count: 10
```

### Feature Flags

### config/featureflags.yaml

```yaml
worker:
  enabled: true
  backend: asynq  # asynq | rabbitmq | redpanda | disable
  
  # Task-level feature flags
  tasks:
    email_notifications: true
    data_export: true
    report_generation: true
    image_processing: true
```

### Shared Worker Types & Interfaces

All worker types and interfaces are defined in `internal/shared/worker/worker.go` so modules can use them without importing infrastructure:

```go
// internal/shared/worker/worker.go
package worker

import (
    "context"
    "time"
)

// TaskPayload defines the structure of task data
type TaskPayload map[string]interface{}

// TaskHandler processes a task
type TaskHandler func(ctx context.Context, payload TaskPayload) error

// TaskDefinition defines a task that a module provides
type TaskDefinition struct {
    TaskName string
    Handler  TaskHandler
}

// CronJobDefinition defines a cron job that a module provides
type CronJobDefinition struct {
    JobID          string
    TaskName       string
    CronExpression CronExpression
    Payload        map[string]interface{}
}

// CronExpression represents a simplified cron expression
type CronExpression struct {
    Minute  int // 0-59 or -1 for any
    Hour    int // 0-23 or -1 for any
    Day     int // 1-31 or -1 for any
    Month   int // 1-12 or -1 for any
    Weekday int // 0-6 (Sun-Sat) or -1 for any
}

// Client enqueues tasks
type Client interface {
    Enqueue(ctx context.Context, taskName string, payload TaskPayload, options ...Option) error
    EnqueueDelayed(ctx context.Context, taskName string, payload TaskPayload, delay time.Duration, options ...Option) error
    Close() error
}

// Server runs workers to process tasks
type Server interface {
    RegisterHandler(taskName string, handler TaskHandler) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}

// Scheduler is responsible for scheduling recurring tasks
type Scheduler interface {
    AddJob(id, taskName string, schedule interface{}, payload TaskPayload) error
    RemoveJob(id string) error
    EnableJob(id string) error
    DisableJob(id string) error
    Start(ctx context.Context) error
    Stop() error
}

// Option defines task options (priority, retry, timeout, etc.)
type Option interface{}

// Helper functions for CronExpression
func EveryMinute() CronExpression
func EveryHour() CronExpression
func Daily(hour, minute int) CronExpression
func Weekly(weekday, hour, minute int) CronExpression
func Monthly(day, hour, minute int) CronExpression
```

### Asynq Implementation

#### Task Enqueueing

```go
// internal/modules/user/worker/tasks.go
package worker

import (
    "context"
    "encoding/json"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

const (
    TaskSendWelcomeEmail = "user:send_welcome_email"
    TaskExportUserData   = "user:export_user_data"
)

type SendWelcomeEmailPayload struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
}

func (p SendWelcomeEmailPayload) MarshalJSON() ([]byte, error) {
    return json.Marshal(p)
}

// EnqueueWelcomeEmail enqueues a welcome email task
func EnqueueWelcomeEmail(ctx context.Context, client sharedworker.Client, userID, email, name string) error {
    payload := SendWelcomeEmailPayload{
        UserID: userID,
        Email:  email,
        Name:   name,
    }
    
    // Enqueue with options (priority, max retries, timeout)
    return client.Enqueue(ctx, TaskSendWelcomeEmail, payload)
}
```

#### Task Handlers

```go
// internal/modules/user/worker/handlers.go
package worker

import (
    "context"
    "encoding/json"
    "fmt"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
    userdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
)

type UserWorkerHandler struct {
    emailService userdomain.EmailService
    userRepo     userdomain.Repository
}

func NewUserWorkerHandler(
    emailService userdomain.EmailService,
    userRepo userdomain.Repository,
) *UserWorkerHandler {
    return &UserWorkerHandler{
        emailService: emailService,
        userRepo:     userRepo,
    }
}

// HandleSendWelcomeEmail processes the welcome email task
func (h *UserWorkerHandler) HandleSendWelcomeEmail(ctx context.Context, payload sharedworker.TaskPayload) error {
    var p SendWelcomeEmailPayload
    
    // Unmarshal payload
    data, _ := json.Marshal(payload)
    if err := json.Unmarshal(data, &p); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    // Get user details
    user, err := h.userRepo.GetByID(ctx, p.UserID)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }
    
    // Send welcome email
    if err := h.emailService.SendWelcomeEmail(ctx, user.Email, user.Name); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    return nil
}

// HandleExportUserData processes the user data export task
func (h *UserWorkerHandler) HandleExportUserData(ctx context.Context, payload workerlib.TaskPayload) error {
    var p ExportUserDataPayload
    
    data, _ := json.Marshal(payload)
    if err := json.Unmarshal(data, &p); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    // Generate export file
    user, err := h.userRepo.GetByID(ctx, p.UserID)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }
    
    // Store export in storage system
    // (implementation depends on storage backend)
    
    // Send notification
    return nil
}
```

#### Asynq Server Implementation

```go
// internal/infrastructure/worker/asynq/server.go
package asynqworker

import (
    "context"
    "fmt"
    
    "github.com/hibiken/asynq"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
    "github.com/kamil5b/go-ptse-monolith/internal/shared/context/logger"
)

type AsynqServer struct {
    srv      *asynq.Server
    mux      *asynq.ServeMux
    handlers map[string]sharedworker.TaskHandler
    log      logger.Logger
}

func NewAsynqServer(redisURL string, concurrency int, log logger.Logger) *AsynqServer {
    return &AsynqServer{
        srv: asynq.NewServer(
            asynq.RedisClientOpt{Addr: redisURL},
            asynq.Config{
                Concurrency: concurrency,
                Queues: map[string]int{
                    "critical": 6,
                    "default":  3,
                    "low":      1,
                },
            },
        ),
        mux:      asynq.NewServeMux(),
        handlers: make(map[string]sharedworker.TaskHandler),
        log:      log,
    }
}

func (s *AsynqServer) RegisterHandler(taskName string, handler sharedworker.TaskHandler) error {
    s.handlers[taskName] = handler
    s.mux.HandleFunc(taskName, func(ctx context.Context, t *asynq.Task) error {
        payload := sharedworker.TaskPayload(t.Payload())
        return handler(ctx, payload)
    })
    s.log.Info("Registered handler for task", "task", taskName)
    return nil
}

func (s *AsynqServer) Start(ctx context.Context) error {
    s.log.Info("Starting Asynq worker server")
    return s.srv.Start(s.mux)
}

func (s *AsynqServer) Stop(ctx context.Context) error {
    s.log.Info("Stopping Asynq worker server")
    s.srv.Stop()
    s.srv.WaitForShutdown()
    return nil
}
```

#### Asynq Client Implementation

```go
// internal/infrastructure/worker/asynq/client.go
package asynqworker

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/hibiken/asynq"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

type AsynqClient struct {
    client *asynq.Client
}

func NewAsynqClient(redisURL string) *AsynqClient {
    return &AsynqClient{
        client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisURL}),
    }
}

func (c *AsynqClient) Enqueue(
    ctx context.Context,
    taskName string,
    payload sharedworker.TaskPayload,
    options ...sharedworker.Option,
) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    task := asynq.NewTask(taskName, data)
    _, err = c.client.EnqueueContext(ctx, task)
    return err
}

func (c *AsynqClient) EnqueueDelayed(
    ctx context.Context,
    taskName string,
    payload sharedworker.TaskPayload,
    delay time.Duration,
    options ...sharedworker.Option,
) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    task := asynq.NewTask(taskName, data)
    _, err = c.client.EnqueueContext(
        ctx,
        task,
        asynq.ProcessIn(delay),
    )
    return err
}

func (c *AsynqClient) Close() error {
    return c.client.Close()
}
```

### RabbitMQ Implementation

```go
// internal/infrastructure/worker/rabbitmq/client.go
package rabbitmqworker

import (
    "context"
    "encoding/json"
    "time"
    
    amqp "github.com/rabbitmq/amqp091-go"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

type RabbitMQClient struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    exchange string
    queue    string
}

func NewRabbitMQClient(url, exchange, queue string) (*RabbitMQClient, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    
    // Declare exchange
    if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
        return nil, err
    }
    
    // Declare queue
    if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
        return nil, err
    }
    
    return &RabbitMQClient{
        conn:     conn,
        channel:  ch,
        exchange: exchange,
        queue:    queue,
    }, nil
}

func (c *RabbitMQClient) Enqueue(
    ctx context.Context,
    taskName string,
    payload sharedworker.TaskPayload,
    options ...sharedworker.Option,
) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    return c.channel.PublishWithContext(
        ctx,
        c.exchange,
        taskName, // routing key
        true,     // mandatory
        false,    // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        data,
            Persistent:  true,
        },
    )
}

func (c *RabbitMQClient) EnqueueDelayed(
    ctx context.Context,
    taskName string,
    payload sharedworker.TaskPayload,
    delay time.Duration,
    options ...sharedworker.Option,
) error {
    // RabbitMQ requires plugin for delayed delivery
    // For now, just enqueue immediately
    return c.Enqueue(ctx, taskName, payload, options...)
}

func (c *RabbitMQClient) Close() error {
    if err := c.channel.Close(); err != nil {
        return err
    }
    return c.conn.Close()
}
```

### Redpanda/Kafka Implementation

```go
// internal/infrastructure/worker/redpanda/client.go
package redpandaworker

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/segmentio/kafka-go"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

type RedpandaClient struct {
    writer *kafka.Writer
    topic  string
}

func NewRedpandaClient(brokers []string, topic string) *RedpandaClient {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(brokers...),
        Topic:    topic,
        Balancer: &kafka.LeastBytes{},
    }
    
    return &RedpandaClient{
        writer: writer,
        topic:  topic,
    }
}

func (c *RedpandaClient) Enqueue(
    ctx context.Context,
    taskName string,
    payload sharedworker.TaskPayload,
    options ...sharedworker.Option,
) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    return c.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(taskName),
        Value: data,
    })
}

func (c *RedpandaClient) EnqueueDelayed(
    ctx context.Context,
    taskName string,
    payload sharedworker.TaskPayload,
    delay time.Duration,
    options ...sharedworker.Option,
) error {
    // Redpanda doesn't natively support delayed delivery
    // Could use scheduled processing topic or external scheduler
    return c.Enqueue(ctx, taskName, payload, options...)
}

func (c *RedpandaClient) Close() error {
    return c.writer.Close()
}
```

### Integrating Workers into Service Layer

```go
// internal/modules/user/service/v1/service_v1.user.go
package servicev1

import (
    "context"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
    userdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
    userworker "github.com/kamil5b/go-ptse-monolith/internal/modules/user/worker"
)

type UserService struct {
    repository  userdomain.Repository
    workerClient sharedworker.Client
    eventBus    events.EventBus
}

func NewUserService(
    repo userdomain.Repository,
    client sharedworker.Client,
    bus events.EventBus,
) *UserService {
    return &UserService{
        repository:   repo,
        workerClient: client,
        eventBus:     bus,
    }
}

func (s *UserService) Create(ctx context.Context, req *userdomain.CreateUserRequest) (*userdomain.User, error) {
    user := &userdomain.User{
        ID:    uuid.New().String(),
        Email: req.Email,
        Name:  req.Name,
    }
    
    if err := s.repository.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Enqueue welcome email task (async)
    _ = userworker.EnqueueWelcomeEmail(
        ctx,
        s.workerClient,
        user.ID,
        user.Email,
        user.Name,
    )
    
    // Publish domain event
    if s.eventBus != nil {
        _ = s.eventBus.Publish(ctx, &userdomain.UserCreatedEvent{
            UserID:  user.ID,
            Email:   user.Email,
            Created: time.Now(),
        })
    }
    
    return user, nil
}
```

### Wiring Workers in Container

```go
// internal/app/core/container.go (excerpt)
package core

import (
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
    asynqworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/asynq"
    rabbitmqworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/rabbitmq"
    redpandaworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/redpanda"
    infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
    userworker "github.com/kamil5b/go-ptse-monolith/internal/modules/user/worker"
)

func buildWorkerClient(config Config, featureFlags FeatureFlags) (sharedworker.Client, error) {
    if !featureFlags.Worker.Enabled {
        return infraworker.NewNoOpClient(), nil
    }
    
    switch featureFlags.Worker.Backend {
    case "asynq":
        return asynqworker.NewAsynqClient(config.App.Worker.Asynq.RedisURL), nil
    case "rabbitmq":
        return rabbitmqworker.NewRabbitMQClient(
            config.App.Worker.RabbitMQ.URL,
            config.App.Worker.RabbitMQ.Exchange,
            config.App.Worker.RabbitMQ.Queue,
        )
    case "redpanda":
        return redpandaworker.NewRedpandaClient(
            config.App.Worker.Redpanda.Brokers,
            config.App.Worker.Redpanda.Topic,
        ), nil
    default:
        return infraworker.NewNoOpClient(), nil
    }
}

func buildWorkerServer(config Config, featureFlags FeatureFlags, container *Container) (sharedworker.Server, error) {
    if !featureFlags.Worker.Enabled {
        return infraworker.NewNoOpServer(), nil
    }
    
    var workerServer sharedworker.Server
    var err error
    
    switch featureFlags.Worker.Backend {
    case "asynq":
        workerServer = asynqworker.NewAsynqServer(
            config.App.Worker.Asynq.RedisURL,
            config.App.Worker.Asynq.Concurrency,
            container.Logger,
        )
    case "rabbitmq":
        // RabbitMQ server setup
    case "redpanda":
        // Redpanda server setup
    }
    
    // Register handlers
    userWorkerHandler := userworker.NewUserWorkerHandler(container.UserRepo)
    if err := workerServer.RegisterHandler(
        userworker.TaskSendWelcomeEmail,
        userWorkerHandler.HandleSendWelcomeEmail,
    ); err != nil {
        return nil, err
    }
    
    return workerServer, nil
}
```

### Running Workers

```bash
# Start worker server for Asynq
go run . worker

# Start worker server with specific backend
WORKER_BACKEND=rabbitmq go run . worker

# Start with custom concurrency
WORKER_CONCURRENCY=20 go run . worker
```

### Worker Command Handler

```go
// cmd/bootstrap/bootstrap.server.go (excerpt)
package bootstrap

func Server() error {
    config := LoadConfig()
    featureFlags := LoadFeatureFlags()
    
    if featureFlags.Worker.Enabled {
        // Start worker server
        workerServer, err := container.WorkerServer()
        if err != nil {
            return err
        }
        
        // Run in separate goroutine
        go func() {
            if err := workerServer.Start(context.Background()); err != nil {
                log.Fatal(err)
            }
        }()
    }
    
    // Start HTTP server...
}
```

### Best Practices

1. **Idempotent Tasks**: Ensure tasks can safely run multiple times (retries)
   ```go
   // ✅ Good: Check if already processed
   processed, _ := repo.CheckIfProcessed(ctx, userID, "welcome_email")
   if processed {
       return nil
   }
   ```

2. **Graceful Shutdown**: Always gracefully stop workers
   ```go
   sigChan := make(chan os.Signal, 1)
   signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
   <-sigChan
   workerServer.Stop(ctx)
   ```

3. **Meaningful Error Handling**: Distinguish retryable vs non-retryable errors
   ```go
   if err := externalAPI.Call(); err != nil {
       if isRetryable(err) {
           return err // Will retry
       } else {
           return &NonRetryableError{err} // Won't retry
       }
   }
   ```

4. **Task Monitoring**: Log task progress
   ```go
   log.Info("Processing task",
       "task_id", taskID,
       "user_id", userID,
       "retries", retryCount,
   )
   ```

5. **Payload Validation**: Always validate and unmarshal payloads safely
   ```go
   if err := json.Unmarshal(payload, &p); err != nil {
       return fmt.Errorf("invalid payload: %w", err)
   }
   ```

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Tasks not being processed | Check worker server is running and handlers are registered |
| High memory usage | Reduce concurrency or check for memory leaks in handlers |
| Tasks stuck in queue | Check task handler error handling and retry configuration |
| Connection timeouts | Verify broker connectivity and network configuration |

---

## Email Services

The application supports multiple email providers for sending transactional and notification emails. Email services are integrated with the worker system to enable asynchronous email sending via background tasks.

### Supported Providers

| Provider | Use Case | Features | Production Ready |
|----------|----------|----------|------------------|
| **SMTP** | Standard email | Native SMTP protocol, TLS/auth, attachments | ✅ Yes |
| **Mailgun** | Email API service | REST API, templates, tracking, deliverability | ✅ Yes |
| **NoOp** | Development/Testing | Mock implementation, logs to stdout | ✅ Yes |

### Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                   Service Layer                              │
│                  (Business Logic)                            │
└────────────────┬─────────────────────────────────────────────┘
                 │ Enqueue Email Task
                 ▼
┌──────────────────────────────────────────────────────────────┐
│                   Worker Queue                               │
│         (Asynq / RabbitMQ / Redpanda)                       │
└────────────────┬─────────────────────────────────────────────┘
                 │ Dequeue & Process
                 ▼
┌──────────────────────────────────────────────────────────────┐
│              Email Service Layer                             │
│   (SMTP / Mailgun / NoOp)                                    │
└────────────────┬─────────────────────────────────────────────┘
                 │ Send
                 ▼
┌──────────────────────────────────────────────────────────────┐
│            Email Provider Backend                            │
│     (SMTP Server / Mailgun API / STDOUT)                     │
└──────────────────────────────────────────────────────────────┘
```

### Project Structure

```
internal/
├── shared/
│   └── email/
│       ├── email.go         # EmailService interface
│       └── noop.go          # NoOp implementation
└── infrastructure/
    └── email/
        ├── smtp/
        │   └── smtp.go      # SMTP implementation
        └── mailgun/
            └── mailgun.go   # Mailgun implementation
```

### Configuration

#### config/config.yaml

```yaml
app:
  email:
    enabled: true
    provider: "smtp"  # smtp | mailgun | noop
    
    # SMTP configuration
    smtp:
      host: "smtp.gmail.com"
      port: 587
      username: "your-email@gmail.com"
      password: "your-app-password"
      from_addr: "noreply@example.com"
      from_name: "MyApp"
    
    # Mailgun configuration
    mailgun:
      domain: "mg.example.com"
      api_key: "key-xxxx"
      from_addr: "noreply@example.com"
      from_name: "MyApp"
```

#### config/featureflags.yaml

```yaml
email:
  enabled: true
  provider: "noop"  # smtp | mailgun | noop
```

### Email Service Interface

```go
// internal/shared/email/email.go
package email

import "context"

type Email struct {
    To            []string              // Recipients
    CC            []string              // Carbon copy
    BCC           []string              // Blind carbon copy
    Subject       string                // Email subject
    TextBody      string                // Plain text body
    HTMLBody      string                // HTML body
    Attachments   []Attachment          // File attachments
    TemplateData  map[string]interface{} // Template variables
    ReplyTo       []string              // Reply-to addresses
    Headers       map[string]string     // Custom headers
}

type Attachment struct {
    Filename string
    Content  []byte
    MimeType string
}

// EmailService sends emails via configured provider
type EmailService interface {
    // Send sends a single email
    Send(ctx context.Context, email *Email) error
    
    // SendBatch sends multiple emails
    SendBatch(ctx context.Context, emails []*Email) error
    
    // SendTemplate sends email using a provider template
    SendTemplate(ctx context.Context, email *Email, templateName string) error
    
    // ValidateEmail validates email format
    ValidateEmail(ctx context.Context, email string) bool
    
    // Health checks service connectivity
    Health(ctx context.Context) error
}
```

### SMTP Implementation

```go
// internal/infrastructure/email/smtp/smtp.go
package smtp

import "context"

type SMTPConfig struct {
    Host     string // SMTP host (e.g., smtp.gmail.com)
    Port     int    // SMTP port (e.g., 587 for TLS)
    Username string // SMTP username
    Password string // SMTP password
    FromAddr string // From email address
    FromName string // From name
}

// SMTPEmailService sends emails via SMTP
type SMTPEmailService struct {
    config SMTPConfig
    addr   string // Cached host:port
}

func NewSMTPEmailService(config SMTPConfig) *SMTPEmailService {
    return &SMTPEmailService{
        config: config,
        addr:   fmt.Sprintf("%s:%d", config.Host, config.Port),
    }
}

// Features:
// - TLS/SMTP authentication
// - HTML and plain text bodies
// - CC/BCC support
// - Custom headers
// - File attachments via MIME
// - Reply-To addresses
// - Email validation via regex
```

### Mailgun Implementation

```go
// internal/infrastructure/email/mailgun/mailgun.go
package mailgun

import "context"

type MailgunConfig struct {
    Domain   string // Mailgun domain
    APIKey   string // Mailgun API key
    FromAddr string // From email address
    FromName string // From name
}

// MailgunEmailService sends emails via Mailgun API
type MailgunEmailService struct {
    config   MailgunConfig
    mg       mailgun.Mailgun
    fromAddr string
}

func NewMailgunEmailService(config MailgunConfig) *MailgunEmailService {
    mg := mailgun.NewMailgun(config.Domain, config.APIKey)
    return &MailgunEmailService{
        config:   config,
        mg:       mg,
        fromAddr: fmt.Sprintf("%s <%s>", config.FromName, config.FromAddr),
    }
}

// Features:
// - Mailgun REST API
// - Template support
// - Batch sending
// - Custom headers
// - File attachments
// - Email validation
// - Delivery tracking
```

### Integration with Workers

Email services are typically used with the worker system for asynchronous sending:

```go
// internal/modules/user/worker/handlers.go
package worker

import (
    "context"
    "github.com/kamil5b/go-ptse-monolith/internal/shared/email"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

type UserWorkerHandler struct {
    userRepository userdomain.Repository
    emailService   email.EmailService
}

// HandleSendWelcomeEmail processes welcome email task
func (h *UserWorkerHandler) HandleSendWelcomeEmail(ctx context.Context, payload sharedworker.TaskPayload) error {
    var p SendWelcomeEmailPayload
    
    // Unmarshal and validate payload
    data, _ := json.Marshal(payload)
    if err := json.Unmarshal(data, &p); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    // Get user
    user, err := h.userRepository.GetByID(ctx, p.UserID)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }
    
    // Send email
    emailMsg := &email.Email{
        To:       []string{user.Email},
        Subject:  "Welcome to Our Platform!",
        HTMLBody: fmt.Sprintf("<h1>Welcome %s!</h1>", user.Name),
        TextBody: fmt.Sprintf("Welcome %s!", user.Name),
    }
    
    if err := h.emailService.Send(ctx, emailMsg); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    return nil
}
```

### Usage Example

```go
// Enqueue a welcome email task
container := NewContainer(featureFlags, config, db, mongoClient)

payload := worker.SendWelcomeEmailPayload{
    UserID: "user-123",
    Email:  "user@example.com",
    Name:   "John Doe",
}

err := container.WorkerClient.Enqueue(ctx, worker.TaskSendWelcomeEmail, payload)
```

The worker will:
1. Dequeue the task
2. Call the handler
3. The handler retrieves user details and sends email via the configured provider
4. Handles retries automatically on failure

### Feature Flag Controls

Email services can be disabled or switched via feature flags without code changes:

```yaml
# Enable SMTP in production
email:
  enabled: true
  provider: "smtp"

# Disable in development
email:
  enabled: false
  provider: "noop"
```

---

## Storage Services

The application supports multiple file storage backends for user uploads, exports, and media management. Storage services are designed to work seamlessly with the worker system for asynchronous file operations and provide a unified interface across different storage providers.

### Supported Backends

| Backend | Use Case | Features | Production Ready |
|---------|----------|----------|------------------|
| **Local Filesystem** | Development/Testing | Simple file storage on disk | ✅ Yes |
| **AWS S3** | Cloud storage | Scalable, highly available, CDN integration | ✅ Yes |
| **S3-Compatible** | MinIO, DigitalOcean Spaces | S3-compatible API providers | ✅ Yes |
| **Google Cloud Storage (GCS)** | Google Cloud integration | Native GCS support, bucket operations | ✅ Yes |
| **NoOp** | Testing | Mock implementation | ✅ Yes |

### Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                   Service Layer                              │
│                  (Business Logic)                            │
└────────────────┬─────────────────────────────────────────────┘
                 │ Upload/Download/Delete
                 ▼
┌──────────────────────────────────────────────────────────────┐
│              Storage Service Layer                           │
│   (Local / AWS S3 / GCS / S3-Compatible)                     │
└────────────────┬─────────────────────────────────────────────┘
                 │ Read/Write/Delete Operations
                 ▼
┌──────────────────────────────────────────────────────────────┐
│            Storage Backend                                   │
│   (Filesystem / AWS S3 / GCS / MinIO)                        │
└──────────────────────────────────────────────────────────────┘
```

### Project Structure

```
internal/
├── shared/
│   └── storage/
│       ├── storage.go         # StorageService interface
│       └── errors.go          # Storage error types
└── infrastructure/
    └── storage/
        ├── noop/
        │   └── storage_noop.go         # NoOp implementation
        ├── local/
        │   └── storage_local.go        # Local filesystem
        ├── s3/
        │   └── storage_s3.go           # AWS S3 & compatible
        └── gcs/
            └── storage_gcs.go          # Google Cloud Storage
```

### Configuration

#### config/config.yaml

```yaml
app:
  storage:
    enabled: true
    local:
      base_path: "./uploads"              # Base directory for local storage
      max_file_size: 104857600             # 100MB in bytes
      allow_public_access: false           # Serve files via HTTP
      public_url: "http://localhost:8080/files"
    
    s3:
      region: "us-east-1"                 # AWS region
      bucket: "my-bucket"                 # S3 bucket name
      access_key_id: "${S3_ACCESS_KEY}"   # Environment variable
      secret_access_key: "${S3_SECRET}"   # Environment variable
      endpoint: ""                         # Leave empty for AWS, set for MinIO/Spaces
      use_ssl: true
      path_style: false                    # Use path-style URLs (true for MinIO)
      
    gcs:
      project_id: "my-project"            # GCP project ID
      bucket: "my-bucket"                 # GCS bucket name
      credentials_file: "/path/to/creds.json" # Service account JSON
      credentials_json: "${GCS_CREDENTIALS}"  # Or use env var
```

#### config/featureflags.yaml

```yaml
storage:
  enabled: true
  backend: "local"  # local | s3 | gcs | s3-compatible | noop
  
  # Backend-specific feature flags
  s3:
    enable_encryption: true           # Enable server-side encryption
    storage_class: "STANDARD"         # Storage class (STANDARD, GLACIER, etc)
    presigned_url_ttl: 3600          # Presigned URL validity in seconds
  
  gcs:
    storage_class: "STANDARD"         # Storage class
    metadata_cache: true              # Cache object metadata
```

### Storage Service Interface

```go
// internal/shared/storage/storage.go
package storage

import (
	"context"
	"io"
	"time"
)

// StorageObject represents metadata about a stored object
type StorageObject struct {
	Name          string                 // Object name/path
	Size          int64                  // File size in bytes
	ContentType   string                 // MIME type
	ETag          string                 // Entity tag (MD5 or hash)
	LastModified  time.Time              // Last modification time
	Metadata      map[string]string      // Custom metadata
	PresignedURL  string                 // Presigned URL (if applicable)
}

// UploadOptions configures upload behavior
type UploadOptions struct {
	ContentType     string                 // MIME type
	Metadata        map[string]string      // Custom metadata
	CacheControl    string                 // Cache-Control header
	ContentEncoding string                 // Content-Encoding header
	ACL             string                 // Access control (private/public-read)
	ServerSideEncryption bool               // Enable encryption
}

// StorageService provides unified file storage operations
type StorageService interface {
	// Upload stores a file and returns metadata
	Upload(ctx context.Context, path string, reader io.Reader, opts *UploadOptions) (*StorageObject, error)

	// UploadBytes stores bytes and returns metadata
	UploadBytes(ctx context.Context, path string, data []byte, opts *UploadOptions) (*StorageObject, error)

	// Download retrieves a file
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// GetBytes retrieves file contents as bytes
	GetBytes(ctx context.Context, path string) ([]byte, error)

	// GetObject retrieves object metadata
	GetObject(ctx context.Context, path string) (*StorageObject, error)

	// Delete removes a file
	Delete(ctx context.Context, path string) error

	// DeletePrefix removes all objects with given prefix
	DeletePrefix(ctx context.Context, prefix string) error

	// Exists checks if object exists
	Exists(ctx context.Context, path string) (bool, error)

	// ListObjects lists objects in a path prefix
	ListObjects(ctx context.Context, prefix string, recursive bool) ([]*StorageObject, error)

	// GetPresignedURL generates a temporary public URL (if supported)
	GetPresignedURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// Copy copies an object within storage
	Copy(ctx context.Context, sourcePath, destPath string) (*StorageObject, error)

	// Health checks the health of the storage service
	Health(ctx context.Context) error
}
```

### Local Filesystem Implementation

```go
// internal/infrastructure/storage/local/storage_local.go
package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
)

type LocalStorageConfig struct {
	BasePath            string        // Base directory path
	MaxFileSize         int64         // Maximum file size in bytes
	AllowPublicAccess   bool          // Serve files via HTTP
	PublicURL           string        // Public URL prefix
	CreateMissingDirs   bool          // Auto-create directories
	FilePermissions     os.FileMode   // File permissions (default: 0644)
	DirPermissions      os.FileMode   // Directory permissions (default: 0755)
}

// LocalStorageService stores files in local filesystem
type LocalStorageService struct {
	config LocalStorageConfig
}

func NewLocalStorageService(config LocalStorageConfig) (*LocalStorageService, error) {
	// Validate and create base directory
	if err := os.MkdirAll(config.BasePath, config.DirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorageService{config: config}, nil
}

func (s *LocalStorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	fullPath := filepath.Join(s.config.BasePath, path)

	// Create directory if needed
	if s.config.CreateMissingDirs {
		if err := os.MkdirAll(filepath.Dir(fullPath), s.config.DirPermissions); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Check size limit
	if written > s.config.MaxFileSize {
		os.Remove(fullPath)
		return nil, fmt.Errorf("file exceeds maximum size: %d > %d", written, s.config.MaxFileSize)
	}

	// Change permissions
	if err := os.Chmod(fullPath, s.config.FilePermissions); err != nil {
		return nil, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Get file info
	info, _ := os.Stat(fullPath)

	result := &storage.StorageObject{
		Name:         path,
		Size:         written,
		ContentType:  opts.ContentType,
		LastModified: info.ModTime(),
		Metadata:     opts.Metadata,
	}

	// Set presigned URL if public access enabled
	if s.config.AllowPublicAccess {
		result.PresignedURL = fmt.Sprintf("%s/%s", s.config.PublicURL, path)
	}

	return result, nil
}

func (s *LocalStorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.config.BasePath, path)
	return os.Open(fullPath)
}

func (s *LocalStorageService) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.config.BasePath, path)
	return os.Remove(fullPath)
}

func (s *LocalStorageService) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.config.BasePath, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// ... additional methods (GetObject, ListObjects, Copy, Health, etc.)
```

### AWS S3 & S3-Compatible Implementation

```go
// internal/infrastructure/storage/s3/storage_s3.go
package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
)

type S3StorageConfig struct {
	Region                string // AWS region
	Bucket                string // S3 bucket name
	AccessKeyID           string // AWS access key
	SecretAccessKey       string // AWS secret key
	Endpoint              string // Custom endpoint (for MinIO, etc)
	UseSSL                bool   // Use SSL for endpoint
	PathStyle             bool   // Use path-style URLs (true for MinIO)
	PresignedURLTTL       int    // Presigned URL validity in seconds
	ServerSideEncryption  bool   // Enable encryption
	StorageClass          string // Storage class (STANDARD, GLACIER, etc)
}

// S3StorageService stores files in AWS S3 or compatible services
type S3StorageService struct {
	config        S3StorageConfig
	client        *s3.Client
	presigner     *s3.PresignClient
	uploader      *manager.Uploader
	downloader    *manager.Downloader
}

func NewS3StorageService(cfg S3StorageConfig) (*S3StorageService, error) {
	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.PathStyle
		}
	})

	return &S3StorageService{
		config:     cfg,
		client:     client,
		presigner:  s3.NewPresignClient(client),
		uploader:   manager.NewUploader(client),
		downloader: manager.NewDownloader(client),
	}, nil
}

func (s *S3StorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	// Prepare put object options
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
		Body:   reader,
	}

	// Set content type
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	// Set storage class
	if s.config.StorageClass != "" {
		input.StorageClass = types.StorageClass(s.config.StorageClass)
	}

	// Set server-side encryption
	if s.config.ServerSideEncryption {
		input.ServerSideEncryption = types.ServerSideEncryptionAes256
	}

	// Upload file
	output, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
	}

	// Get object metadata
	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	// Generate presigned URL
	presignInput := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	}

	presignOutput, err := s.presigner.PresignGetObject(ctx, presignInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(s.config.PresignedURLTTL) * time.Second
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &storage.StorageObject{
		Name:         path,
		Size:         *headOutput.ContentLength,
		ContentType:  *headOutput.ContentType,
		ETag:         *output.ETag,
		LastModified: *headOutput.LastModified,
		Metadata:     headOutput.Metadata,
		PresignedURL: presignOutput.URL,
	}, nil
}

func (s *S3StorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return output.Body, nil
}

func (s *S3StorageService) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	return err
}

func (s *S3StorageService) GetPresignedURL(
	ctx context.Context,
	path string,
	expiration time.Duration,
) (string, error) {
	presignInput := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	}

	presignOutput, err := s.presigner.PresignGetObject(ctx, presignInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = expiration
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignOutput.URL, nil
}

// ... additional methods (ListObjects, Copy, Health, etc.)
```

### Google Cloud Storage Implementation

```go
// internal/infrastructure/storage/gcs/storage_gcs.go
package gcs

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	storagepkg "github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
)

type GCSStorageConfig struct {
	ProjectID        string // GCP project ID
	Bucket           string // GCS bucket name
	CredentialsFile  string // Path to service account JSON
	CredentialsJSON  string // Inline JSON credentials
	StorageClass     string // Storage class (STANDARD, NEARLINE, COLDLINE, ARCHIVE)
	Location         string // Bucket location
	MetadataCache    bool   // Cache object metadata
	PresignedURLTTL  int    // Presigned URL validity in seconds
}

// GCSStorageService stores files in Google Cloud Storage
type GCSStorageService struct {
	config GCSStorageConfig
	client *storage.Client
	bucket *storage.BucketHandle
}

func NewGCSStorageService(ctx context.Context, cfg GCSStorageConfig) (*GCSStorageService, error) {
	var client *storage.Client
	var err error

	// Create GCS client with credentials
	if cfg.CredentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(cfg.CredentialsFile))
	} else if cfg.CredentialsJSON != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorageService{
		config: cfg,
		client: client,
		bucket: client.Bucket(cfg.Bucket),
	}, nil
}

func (s *GCSStorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storagepkg.UploadOptions,
) (*storagepkg.StorageObject, error) {
	object := s.bucket.Object(path)
	writer := object.NewWriter(ctx)

	// Set content type
	if opts.ContentType != "" {
		writer.ContentType = opts.ContentType
	}

	// Set metadata
	if opts.Metadata != nil {
		writer.Metadata = opts.Metadata
	}

	// Set cache control
	if opts.CacheControl != "" {
		writer.CacheControl = opts.CacheControl
	}

	// Copy data
	if _, err := io.Copy(writer, reader); err != nil {
		return nil, fmt.Errorf("failed to write to GCS: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Get object attributes
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	// Generate presigned URL (requires additional setup with key)
	// For now, return object metadata
	return &storagepkg.StorageObject{
		Name:         path,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         attrs.ETag,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
	}, nil
}

func (s *GCSStorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	reader, err := s.bucket.Object(path).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}
	return reader, nil
}

func (s *GCSStorageService) Delete(ctx context.Context, path string) error {
	return s.bucket.Object(path).Delete(ctx)
}

func (s *GCSStorageService) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.bucket.Object(path).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	return err == nil, err
}

func (s *GCSStorageService) GetObject(ctx context.Context, path string) (*storagepkg.StorageObject, error) {
	attrs, err := s.bucket.Object(path).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return &storagepkg.StorageObject{
		Name:         attrs.Name,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         attrs.ETag,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
	}, nil
}

func (s *GCSStorageService) ListObjects(
	ctx context.Context,
	prefix string,
	recursive bool,
) ([]*storagepkg.StorageObject, error) {
	query := &storage.Query{Prefix: prefix}
	if !recursive {
		query.Delimiter = "/"
	}

	it := s.bucket.Objects(ctx, query)
	var objects []*storagepkg.StorageObject

	for {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		objects = append(objects, &storagepkg.StorageObject{
			Name:         attrs.Name,
			Size:         attrs.Size,
			ContentType:  attrs.ContentType,
			ETag:         attrs.ETag,
			LastModified: attrs.Updated,
			Metadata:     attrs.Metadata,
		})
	}

	return objects, nil
}

// ... additional methods (Copy, Health, etc.)
```

### MinIO Configuration Example

MinIO is an S3-compatible object storage server. To use MinIO with the S3 implementation:

```yaml
# config/config.yaml
app:
  storage:
    s3:
      region: "us-east-1"
      bucket: "my-bucket"
      access_key_id: "minioadmin"
      secret_access_key: "minioadmin"
      endpoint: "http://localhost:9000"  # MinIO endpoint
      use_ssl: false
      path_style: true                   # MinIO requires path-style URLs
      presigned_url_ttl: 3600
```

Then use the S3 backend with MinIO configuration.

### Integration with Modules

Storage services can be injected into handlers and services for file upload/download operations:

```go
// internal/modules/product/handler/v1/handler_v1.product.go
package handlerv1

import (
	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

type ProductHandler struct {
	service   productdomain.ProductService
	storage   storage.StorageService
}

// UploadProductImage handles image uploads
func (h *ProductHandler) UploadProductImage(c sharedctx.Context) error {
	productID := c.Param("product_id")
	
	// Get uploaded file from request
	file, err := c.GetFile("image") // framework-specific
	if err != nil {
		return c.JSON(400, map[string]string{"error": "no file uploaded"})
	}

	// Upload to storage
	storagePath := fmt.Sprintf("products/%s/%s", productID, file.Filename)
	obj, err := h.storage.Upload(c.GetContext(), storagePath, file, &storage.UploadOptions{
		ContentType: file.ContentType,
		ACL:         "public-read",
		Metadata: map[string]string{
			"product_id": productID,
			"uploaded_by": c.GetUserID(),
		},
	})
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	// Return presigned URL or public URL
	return c.JSON(200, map[string]string{
		"url": obj.PresignedURL,
		"size": fmt.Sprintf("%d", obj.Size),
	})
}

// GetProductImage serves product images
func (h *ProductHandler) GetProductImage(c sharedctx.Context) error {
	productID := c.Param("product_id")
	imageName := c.Param("image_name")
	
	storagePath := fmt.Sprintf("products/%s/%s", productID, imageName)
	
	// Download from storage
	reader, err := h.storage.Download(c.GetContext(), storagePath)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "image not found"})
	}
	defer reader.Close()

	// Stream response (framework-specific)
	return c.Stream(200, "image/jpeg", reader)
}
```

### Wiring Storage in Container

```go
// internal/app/core/container.go (excerpt)
package core

import (
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/local"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/s3"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/storage/gcs"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
)

func buildStorageService(config Config, featureFlags FeatureFlags) (storage.StorageService, error) {
	if !featureFlags.Storage.Enabled {
		return &storage.NoOpStorageService{}, nil
	}

	switch featureFlags.Storage.Backend {
	case "local":
		return local.NewLocalStorageService(local.LocalStorageConfig{
			BasePath:          config.App.Storage.Local.BasePath,
			MaxFileSize:       config.App.Storage.Local.MaxFileSize,
			AllowPublicAccess: config.App.Storage.Local.AllowPublicAccess,
			PublicURL:         config.App.Storage.Local.PublicURL,
			CreateMissingDirs: true,
		})

	case "s3":
		return s3.NewS3StorageService(s3.S3StorageConfig{
			Region:              config.App.Storage.S3.Region,
			Bucket:              config.App.Storage.S3.Bucket,
			AccessKeyID:         config.App.Storage.S3.AccessKeyID,
			SecretAccessKey:     config.App.Storage.S3.SecretAccessKey,
			Endpoint:            config.App.Storage.S3.Endpoint,
			UseSSL:              config.App.Storage.S3.UseSSL,
			PathStyle:           config.App.Storage.S3.PathStyle,
			PresignedURLTTL:     config.App.Storage.S3.PresignedURLTTL,
			ServerSideEncryption: featureFlags.Storage.S3.EnableEncryption,
			StorageClass:        featureFlags.Storage.S3.StorageClass,
		})

	case "gcs":
		return gcs.NewGCSStorageService(context.Background(), gcs.GCSStorageConfig{
			ProjectID:        config.App.Storage.GCS.ProjectID,
			Bucket:           config.App.Storage.GCS.Bucket,
			CredentialsFile:  config.App.Storage.GCS.CredentialsFile,
			StorageClass:     featureFlags.Storage.GCS.StorageClass,
			MetadataCache:    featureFlags.Storage.GCS.MetadataCache,
		})

	default:
		return &storage.NoOpStorageService{}, nil
	}
}

// In NewContainer, inject storage service
container.StorageService = buildStorageService(config, featureFlags)
```

### Error Handling

```go
// internal/shared/storage/errors.go
package storage

import "fmt"

type StorageErrorType string

const (
	ErrTypeNotFound      StorageErrorType = "STORAGE_NOT_FOUND"
	ErrTypeSizeLimitExceeded = "STORAGE_SIZE_LIMIT_EXCEEDED"
	ErrTypePermissionDenied  = "STORAGE_PERMISSION_DENIED"
	ErrTypeInvalidPath       = "STORAGE_INVALID_PATH"
	ErrTypeServiceError      = "STORAGE_SERVICE_ERROR"
)

type StorageError struct {
	Type      StorageErrorType
	Message   string
	Err       error
	HTTPCode  int
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Helper constructors
func NotFound(path string) *StorageError {
	return &StorageError{
		Type:     ErrTypeNotFound,
		Message:  fmt.Sprintf("object not found: %s", path),
		HTTPCode: 404,
	}
}

func SizeLimitExceeded(max int64) *StorageError {
	return &StorageError{
		Type:     ErrTypeSizeLimitExceeded,
		Message:  fmt.Sprintf("file size exceeds limit of %d bytes", max),
		HTTPCode: 413,
	}
}
```

### Best Practices

1. **Validate File Types**: Check MIME types and extensions before upload
   ```go
   allowedTypes := map[string]bool{
       "image/jpeg": true,
       "image/png": true,
       "image/gif": true,
   }
   
   if !allowedTypes[opts.ContentType] {
       return nil, fmt.Errorf("file type not allowed")
   }
   ```

2. **Generate Safe Paths**: Use UUIDs to prevent directory traversal
   ```go
   storagePath := fmt.Sprintf("uploads/%s/%s", 
       uuid.New().String(), 
       filepath.Base(filename),
   )
   ```

3. **Implement Cleanup**: Delete files after associated records are removed
   ```go
   if err := s.repo.Delete(ctx, id); err != nil {
       return err
   }
   // Delete associated files
   _ = s.storage.DeletePrefix(ctx, fmt.Sprintf("products/%s/", id))
   ```

4. **Use Async Uploads**: Enqueue large file operations as workers
   ```go
   // Enqueue export task (async)
   s.workerClient.Enqueue(ctx, "export:data_export", map[string]interface{}{
       "user_id": userID,
       "format": "csv",
   })
   ```

5. **Monitor Storage**: Health checks and metrics
   ```go
   if err := s.storage.Health(ctx); err != nil {
       log.Error("storage service unhealthy", "error", err)
   }
   ```

### Troubleshooting

| Issue | Solution |
|-------|----------|
| 403 Forbidden on S3 | Check IAM permissions and bucket policy |
| MinIO connection failed | Verify endpoint URL and credentials |
| Large file uploads timeout | Increase timeout and use multipart uploads |
| Storage quota exceeded | Check bucket size limits and cleanup old files |
| Presigned URLs not working | Verify TTL and backend time synchronization |

---

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
