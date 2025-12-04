# Go-PSTE-Boilerplate Developer Guide

> **Version:** 1.0.0  
> **Last Updated:** December 4, 2025  
> **Go Version:** 1.24.7

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Project Structure](#project-structure)
3. [Creating a New Module](#creating-a-new-module)
4. [Module Development](#module-development)
5. [Feature Flags](#feature-flags)
6. [Database Operations](#database-operations)
7. [Cross-Module Communication](#cross-module-communication)
8. [Testing](#testing)
9. [Dependency Linter](#dependency-linter)
10. [Debugging & Troubleshooting](#debugging--troubleshooting)
11. [Best Practices](#best-practices)

---

## Quick Start

### Prerequisites

- **Go 1.24.7+** ([Download](https://go.dev/dl/))
- **PostgreSQL 14+** (or use Docker)
- **MongoDB 6.0+** (optional, or use Docker)
- **Git**

### Setup Local Environment

```bash
# Clone the repository
git clone https://github.com/kamil5b/go-pste-boilerplate.git
cd go-pste-boilerplate

# Install Go dependencies
go mod tidy

# Copy config template (if needed)
cp config/config.yaml.example config/config.yaml

# Edit config with your database credentials
nano config/config.yaml

# Run database migrations
go run . migration sql up

# Start the application
go run .
```

### Using Docker for Databases

```bash
# Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=appdb \
  -p 5432:5432 \
  postgres:14

# Start MongoDB
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  mongo:6

# Update config/config.yaml with these credentials
```

### Verify Installation

```bash
# Check dependencies
go mod verify

# Run dependency linter
go run cmd/lint-deps/main.go

# Run migrations
go run . migration sql up

# Start server
go run . server
```

---

## Project Structure

### Top-Level Layout

```
go-pste-boilerplate/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── config/                    # Configuration files
│   ├── config.yaml            # Runtime configuration
│   └── featureflags.yaml      # Feature toggle settings
├── cmd/
│   ├── bootstrap/             # Application initialization
│   └── lint-deps/             # Dependency linter tool
├── internal/
│   ├── app/                   # Application core (DI, config, HTTP setup)
│   ├── infrastructure/        # External services (DB, cache, email, etc.)
│   ├── modules/               # Business modules (auth, product, user, etc.)
│   ├── shared/                # Shared kernel (events, errors, context, UoW)
│   └── transports/            # HTTP framework adapters
├── docs/                      # Documentation
└── scripts/                   # Development scripts
```

### Internal Directory Structure

#### `/internal/app` - Application Core

```
internal/app/
├── core/
│   ├── config.go              # Configuration structs & loader
│   ├── container.go           # Dependency injection container
│   └── feature_flag.go        # Feature flag structs & loader
└── http/
    ├── echo.go                # Echo HTTP server setup
    ├── gin.go                 # Gin HTTP server setup
    ├── fiber.go               # Fiber HTTP server setup
    ├── fasthttp.go            # FastHTTP server setup
    ├── nethttp.go             # net/http server setup
    ├── helpers.go             # Middleware helpers
    └── routes.go              # Route definitions
```

#### `/internal/modules` - Business Modules

```
internal/modules/
├── <module>/
│   ├── domain/
│   │   ├── model.go           # Domain entities
│   │   ├── interfaces.go      # Handler/Service/Repository interfaces
│   │   ├── request.go         # Request DTOs
│   │   ├── response.go        # Response DTOs
│   │   └── events.go          # Domain events
│   ├── acl/                   # Anti-Corruption Layer
│   │   └── <adapter>.go       # Cross-module adapters
│   ├── handler/
│   │   ├── v1/                # Version 1 implementation
│   │   └── noop/              # No-op/disabled implementation
│   ├── service/
│   │   ├── v1/                # Version 1 implementation
│   │   └── noop/              # No-op/disabled implementation
│   ├── repository/
│   │   ├── sql/               # PostgreSQL implementation
│   │   ├── mongo/             # MongoDB implementation
│   │   └── noop/              # No-op/disabled implementation
│   └── middleware/            # Module-specific middleware (optional)
```

#### `/internal/shared` - Shared Kernel

```
internal/shared/
├── context/
│   └── context.go             # Framework-agnostic HTTP context interface
├── errors/
│   ├── errors.go              # Domain error types
│   ├── http.go                # HTTP status code mapping
│   └── validation.go          # Validation error helpers
├── events/
│   ├── event.go               # Event and EventBus interfaces
│   ├── memory_bus.go          # In-memory EventBus implementation
│   └── errors.go              # Event-related errors
├── uow/
│   └── unit_of_work.go        # Unit of Work interface
├── cache/
│   ├── cache.go               # Cache interface
│   ├── memory.go              # In-memory cache implementation
│   └── errors.go              # Cache-related errors
├── validator/
│   └── validator.go           # Request validation utilities
├── model/
│   ├── request.go             # Base request models
│   └── response.go            # Base response models
└── storage/
    └── storage.go             # File storage interface
```

---

## Creating a New Module

### Step 1: Create Module Structure

```bash
# Create module directories
mkdir -p internal/modules/mymodule/{domain,handler,service,repository,acl}
mkdir -p internal/modules/mymodule/{handler/v1,handler/noop}
mkdir -p internal/modules/mymodule/{service/v1,service/noop}
mkdir -p internal/modules/mymodule/{repository/sql,repository/mongo,repository/noop}
```

### Step 2: Define Domain

**`internal/modules/mymodule/domain/model.go`:**
```go
package domain

// MyEntity represents the core domain entity
type MyEntity struct {
    ID    string
    Name  string
    Email string
}
```

**`internal/modules/mymodule/domain/interfaces.go`:**
```go
package domain

import (
    "context"
    sharedctx "go-pste-boilerplate/internal/shared/context"
)

// Handler interface
type Handler interface {
    Create(c sharedctx.Context) error
    Get(c sharedctx.Context) error
    Update(c sharedctx.Context) error
    Delete(c sharedctx.Context) error
}

// Service interface
type Service interface {
    CreateEntity(ctx context.Context, name, email string) (*MyEntity, error)
    GetEntity(ctx context.Context, id string) (*MyEntity, error)
    UpdateEntity(ctx context.Context, entity *MyEntity) error
    DeleteEntity(ctx context.Context, id string) error
}

// Repository interface
type Repository interface {
    Create(ctx context.Context, entity *MyEntity) error
    GetByID(ctx context.Context, id string) (*MyEntity, error)
    Update(ctx context.Context, entity *MyEntity) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]*MyEntity, error)
}
```

**`internal/modules/mymodule/domain/request.go`:**
```go
package domain

// CreateRequest is the request payload for creating an entity
type CreateRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

// UpdateRequest is the request payload for updating an entity
type UpdateRequest struct {
    Name  string `json:"name"`
    Email string `json:"email" validate:"email"`
}
```

**`internal/modules/mymodule/domain/response.go`:**
```go
package domain

// EntityResponse is the response payload
type EntityResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// ListResponse is the list response payload
type ListResponse struct {
    Items []*EntityResponse `json:"items"`
    Total int               `json:"total"`
}
```

### Step 3: Implement Handler

**`internal/modules/mymodule/handler/v1/handler_v1.mymodule.go`:**
```go
package v1

import (
    "go-pste-boilerplate/internal/modules/mymodule/domain"
    "go-pste-boilerplate/internal/shared/context"
    sharederrors "go-pste-boilerplate/internal/shared/errors"
    "go-pste-boilerplate/internal/shared/validator"
)

type Handler struct {
    service domain.Service
}

func NewHandler(service domain.Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Create(c context.Context) error {
    var req domain.CreateRequest
    
    // Bind request body
    if err := c.BindJSON(&req); err != nil {
        return sharederrors.BadRequest("invalid request payload")
    }
    
    // Validate request
    if err := validator.Validate(req); err != nil {
        return err
    }
    
    // Call service
    entity, err := h.service.CreateEntity(c.Context(), req.Name, req.Email)
    if err != nil {
        return err
    }
    
    // Return response
    return c.JSON(201, domain.EntityResponse{
        ID:    entity.ID,
        Name:  entity.Name,
        Email: entity.Email,
    })
}

func (h *Handler) Get(c context.Context) error {
    id := c.GetParam("id")
    
    entity, err := h.service.GetEntity(c.Context(), id)
    if err != nil {
        return err
    }
    
    return c.JSON(200, domain.EntityResponse{
        ID:    entity.ID,
        Name:  entity.Name,
        Email: entity.Email,
    })
}

// ... Implement Update, Delete, List ...
```

**`internal/modules/mymodule/handler/noop/handler_noop.mymodule.go`:**
```go
package noop

import (
    "go-pste-boilerplate/internal/modules/mymodule/domain"
    "go-pste-boilerplate/internal/shared/context"
    sharederrors "go-pste-boilerplate/internal/shared/errors"
)

type NoOpHandler struct{}

func NewNoOpHandler() *NoOpHandler {
    return &NoOpHandler{}
}

func (h *NoOpHandler) Create(c context.Context) error {
    return sharederrors.Forbidden("handler disabled")
}

func (h *NoOpHandler) Get(c context.Context) error {
    return sharederrors.Forbidden("handler disabled")
}

// ... Implement other methods ...
```

### Step 4: Implement Service

**`internal/modules/mymodule/service/v1/service_v1.mymodule.go`:**
```go
package v1

import (
    "context"
    "github.com/google/uuid"
    "go-pste-boilerplate/internal/modules/mymodule/domain"
)

type Service struct {
    repository domain.Repository
}

func NewService(repository domain.Repository) *Service {
    return &Service{repository: repository}
}

func (s *Service) CreateEntity(ctx context.Context, name, email string) (*domain.MyEntity, error) {
    entity := &domain.MyEntity{
        ID:    uuid.New().String(),
        Name:  name,
        Email: email,
    }
    
    if err := s.repository.Create(ctx, entity); err != nil {
        return nil, err
    }
    
    return entity, nil
}

// ... Implement other methods ...
```

### Step 5: Implement Repository

**`internal/modules/mymodule/repository/sql/repository_sql.mymodule.go`:**
```go
package sql

import (
    "context"
    "github.com/jmoiron/sqlx"
    "go-pste-boilerplate/internal/modules/mymodule/domain"
    sharederrors "go-pste-boilerplate/internal/shared/errors"
)

type Repository struct {
    db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, entity *domain.MyEntity) error {
    query := `
        INSERT INTO mymodule_entities (id, name, email)
        VALUES ($1, $2, $3)
    `
    
    _, err := r.db.ExecContext(ctx, query, entity.ID, entity.Name, entity.Email)
    if err != nil {
        return sharederrors.Internal(err)
    }
    
    return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.MyEntity, error) {
    query := `SELECT id, name, email FROM mymodule_entities WHERE id = $1`
    
    var entity domain.MyEntity
    err := r.db.GetContext(ctx, &entity, query, id)
    if err != nil {
        return nil, sharederrors.NotFound("entity", id)
    }
    
    return &entity, nil
}

// ... Implement other methods ...
```

### Step 6: Update Feature Flags

**`config/featureflags.yaml`:**
```yaml
handler:
  mymodule: v1  # v1 | noop | disable

service:
  mymodule: v1  # v1 | noop | disable

repository:
  mymodule: postgres  # postgres | mongo | noop | disable
```

### Step 7: Register in DI Container

**`internal/app/core/container.go`:** (Add to the appropriate section)
```go
// Repository
var mymoduleRepository domain.Repository
switch featureFlag.Repository.Mymodule {
case "postgres":
    mymoduleRepository = repoSQL.NewRepository(db)
case "mongo":
    mymoduleRepository = repoMongo.NewRepository(mongo)
default:
    mymoduleRepository = repoNoop.NewRepository()
}

// Service
var mymoduleService domain.Service
switch featureFlag.Service.Mymodule {
case "v1":
    mymoduleService = serviceV1.NewService(mymoduleRepository)
default:
    mymoduleService = serviceNoop.NewService()
}

// Handler
var mymoduleHandler domain.Handler
switch featureFlag.Handler.Mymodule {
case "v1":
    mymoduleHandler = handlerV1.NewHandler(mymoduleService)
default:
    mymoduleHandler = handlerNoop.NewHandler()
}
```

### Step 8: Add Routes

**`internal/app/http/routes.go`:**
```go
// Add to the appropriate server setup function
mymoduleGroup := router.Group("/mymodule")
{
    mymoduleGroup.POST("", container.MymoduleHandler.Create)
    mymoduleGroup.GET("/:id", container.MymoduleHandler.Get)
    mymoduleGroup.PUT("/:id", container.MymoduleHandler.Update)
    mymoduleGroup.DELETE("/:id", container.MymoduleHandler.Delete)
}
```

---

## Module Development

### Domain Layer

The domain layer contains pure business logic without framework dependencies:

```go
package domain

// Define core entities
type MyEntity struct {
    ID   string
    Data string
}

// Define interfaces (what the module needs)
type Service interface {
    Process(ctx context.Context, entity *MyEntity) error
}

// Define errors
type EntityNotFoundError struct {
    ID string
}

func (e EntityNotFoundError) Error() string {
    return "entity not found: " + e.ID
}
```

### Handler Layer

Handlers receive HTTP requests and delegate to services:

```go
type Handler struct {
    service domain.Service
}

func (h *Handler) HandleRequest(c sharedctx.Context) error {
    // Parse request
    var req domain.Request
    if err := c.BindJSON(&req); err != nil {
        return sharederrors.BadRequest("invalid request")
    }
    
    // Validate
    if err := validator.Validate(req); err != nil {
        return err
    }
    
    // Process via service
    result, err := h.service.Process(c.Context(), req)
    if err != nil {
        return err
    }
    
    // Return response
    return c.JSON(200, result)
}
```

### Service Layer

Services contain business logic and orchestrate repositories/external services:

```go
type Service struct {
    repo   domain.Repository
    cache  cache.Cache
    events events.EventBus
}

func (s *Service) Process(ctx context.Context, req *domain.Request) (*domain.Response, error) {
    // Check cache
    cached, err := s.cache.Get(ctx, req.ID)
    if err == nil {
        return cached, nil
    }
    
    // Fetch data
    entity, err := s.repo.GetByID(ctx, req.ID)
    if err != nil {
        return nil, err
    }
    
    // Process
    entity.Data = req.NewData
    
    // Save
    if err := s.repo.Update(ctx, entity); err != nil {
        return nil, err
    }
    
    // Cache result
    s.cache.Set(ctx, req.ID, entity, 1*time.Hour)
    
    // Publish event
    s.events.Publish(ctx, &domain.EntityUpdated{
        EntityID: entity.ID,
        OccurredAt: time.Now(),
    })
    
    return &domain.Response{ID: entity.ID}, nil
}
```

### Repository Layer

Repositories handle data access without business logic:

```go
type Repository struct {
    db *sqlx.DB
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
    var entity domain.Entity
    err := r.db.GetContext(ctx, &entity, 
        "SELECT * FROM entities WHERE id = $1", id)
    
    if err == sql.ErrNoRows {
        return nil, sharederrors.NotFound("entity", id)
    }
    
    return &entity, err
}

func (r *Repository) Create(ctx context.Context, entity *domain.Entity) error {
    _, err := r.db.ExecContext(ctx,
        "INSERT INTO entities (id, data) VALUES ($1, $2)",
        entity.ID, entity.Data)
    
    return sharederrors.Internal(err)
}
```

---

## Feature Flags

### Anatomy of Feature Flags

Feature flags control runtime behavior without code changes:

```yaml
# config/featureflags.yaml

# HTTP Framework selection
http_handler: echo  # echo | gin | nethttp | fasthttp | fiber

# Handler versioning
handler:
  authentication: v1    # v1 | noop | disable
  product: v1
  user: v1

# Service versioning
service:
  authentication: v1    # v1 | noop | disable
  product: v1
  user: v1

# Repository backend selection
repository:
  authentication: postgres  # postgres | mongo | noop | disable
  product: postgres
  user: postgres
```

### Using Feature Flags in Code

**Reading flags:**
```go
// In container.go
switch featureFlag.HTTP {
case "echo":
    server = newEchoServer()
case "gin":
    server = newGinServer()
case "nethttp":
    server = newNetHTTPServer()
}
```

**Common patterns:**
```go
// Version-based logic
switch featureFlag.Service.Product {
case "v1":
    service = serviceV1.New(repo)
case "v2":
    service = serviceV2.New(repo)
default:
    service = serviceNoop.New() // No-op service
}

// Backend switching
switch featureFlag.Repository.User {
case "postgres":
    repo = repoSQL.New(db)
case "mongo":
    repo = repoMongo.New(mongoClient)
}
```

### Feature Flag Best Practices

1. **Always provide a default** (usually noop)
2. **One feature per flag** (don't combine concerns)
3. **Document all options** in config.yaml comments
4. **Test all flag combinations** during CI/CD
5. **Use consistent naming** (e.g., `v1`, `v2`, `noop`, `disable`)

---

## Database Operations

### Creating a Migration

**SQL Migration (PostgreSQL):**
```bash
# Create new migration file in internal/infrastructure/db/sql/migration/
# File naming: <timestamp>_<description>.sql

# Example: 20250101000001_create_mymodule_table.sql
CREATE TABLE IF NOT EXISTS mymodule_entities (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**MongoDB Migration:**
```bash
# Create migration script in internal/infrastructure/db/mongo/migration/
# File naming: <timestamp>_<description>.js

# Example: 20250101000001_create_mymodule_indexes.js
db.mymodule_entities.createIndex({ email: 1 }, { unique: true });
db.mymodule_entities.createIndex({ created_at: -1 });
```

### Running Migrations

```bash
# Apply SQL migrations
go run . migration sql up

# Rollback SQL migrations
go run . migration sql down

# Apply MongoDB migrations
go run . migration mongo up
```

### Working with Databases

**PostgreSQL (via sqlx):**
```go
import "github.com/jmoiron/sqlx"

type Repository struct {
    db *sqlx.DB
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Entity, error) {
    var e Entity
    err := r.db.GetContext(ctx, &e, 
        "SELECT id, name, email FROM entities WHERE id = $1", id)
    return &e, err
}
```

**MongoDB:**
```go
import "go.mongodb.org/mongo-driver/mongo"

type Repository struct {
    client *mongo.Client
    db     *mongo.Database
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Entity, error) {
    var e Entity
    err := r.db.Collection("entities").FindOne(ctx, 
        bson.M{"_id": id}).Decode(&e)
    return &e, err
}
```

---

## Cross-Module Communication

### Anti-Corruption Layer (ACL) Pattern

Use ACL when module A needs to use module B's repository synchronously:

**Step 1: Module A defines its interface**
```go
// internal/modules/auth/domain/interfaces.go
package domain

type UserCreator interface {
    CreateUser(ctx context.Context, name, email string) (string, error)
}
```

**Step 2: Module A uses the interface**
```go
// internal/modules/auth/service/v1/service_v1.auth.go
type Service struct {
    userCreator domain.UserCreator  // Injected interface
}
```

**Step 3: ACL adapter implements the interface**
```go
// internal/modules/auth/acl/user_creator.go
package acl

import userdomain "go-pste-boilerplate/internal/modules/user/domain"

type UserCreatorAdapter struct {
    repo userdomain.Repository
}

func (a *UserCreatorAdapter) CreateUser(ctx context.Context, name, email string) (string, error) {
    user := &userdomain.User{
        ID:    uuid.New().String(),
        Name:  name,
        Email: email,
    }
    return user.ID, a.repo.Create(ctx, user)
}
```

**Step 4: Container wires it**
```go
// internal/app/core/container.go
userCreator := acl.NewUserCreatorAdapter(userRepository)
authService = serviceV1.NewService(userCreator, ...)
```

### Event Bus Pattern

Use events when you need asynchronous, decoupled communication:

**Step 1: Define domain event**
```go
// internal/modules/user/domain/events.go
type UserCreated struct {
    UserID    string
    Email     string
    Timestamp time.Time
}

func (e UserCreated) EventName() string     { return "user.created" }
func (e UserCreated) OccurredAt() time.Time { return e.Timestamp }
```

**Step 2: Publish event in service**
```go
// internal/modules/user/service/v1/service_v1.user.go
func (s *Service) CreateUser(ctx context.Context, user *domain.User) error {
    if err := s.repo.Create(ctx, user); err != nil {
        return err
    }
    
    // Publish event
    return s.eventBus.Publish(ctx, &domain.UserCreated{
        UserID:    user.ID,
        Email:     user.Email,
        Timestamp: time.Now(),
    })
}
```

**Step 3: Subscribe in another module**
```go
// internal/modules/auth/service/v1/init.go
func SubscribeToUserEvents(eventBus events.EventBus) error {
    return eventBus.Subscribe("user.created", func(ctx context.Context, e events.Event) error {
        userCreated := e.(*userdomain.UserCreated)
        // Handle user creation
        return nil
    })
}
```

### ACL vs Events Decision Tree

```
Need synchronous response?
├─ YES → Use ACL (interface adapter)
└─ NO → Use Events (event bus)

Need immediate consistency?
├─ YES → Use ACL
└─ NO → Use Events

Multiple modules reacting?
├─ YES → Use Events
└─ NO → Use ACL

Transaction critical?
├─ YES → Use ACL
└─ NO → Use Events
```

---

## Testing

### Unit Testing

**Test structure:**
```go
// internal/modules/product/service/v1/service_v1.product_test.go
package v1

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "go-pste-boilerplate/internal/modules/product/domain"
    "go-pste-boilerplate/internal/modules/product/repository/noop"
)

func TestCreateProduct(t *testing.T) {
    // Arrange
    repo := noop.NewRepository()
    service := NewService(repo)
    
    // Act
    product, err := service.CreateProduct(context.Background(), "Test", 99.99)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Test", product.Name)
    assert.Equal(t, 99.99, product.Price)
}
```

### Mock Generation

```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks for a package
mockery --name=Repository --dir=internal/modules/product/domain --output=internal/modules/product/repository/mocks

# Generate all mocks
make mocks
```

### Testing Handler Layer

```go
import (
    "github.com/stretchr/testify/assert"
    "go-pste-boilerplate/internal/modules/product/handler/v1"
    "go-pste-boilerplate/internal/shared/context"
)

func TestCreateHandler(t *testing.T) {
    // Use mock service
    mockService := new(domain.MockService)
    handler := v1.NewHandler(mockService)
    
    // Use mock context
    mockCtx := new(context.MockContext)
    
    // Test
    err := handler.Create(mockCtx)
    assert.NoError(t, err)
}
```

---

## Dependency Linter

### Running the Linter

```bash
# Check for violations
go run cmd/lint-deps/main.go

# Verbose output
go run cmd/lint-deps/main.go -v

# Specify custom root
go run cmd/lint-deps/main.go -root ./internal/modules
```

### Rules Enforced

| Rule | Status | Description |
|------|--------|-------------|
| No cross-module imports | ❌ Error | Modules cannot import from other modules |
| ACL exception | ✅ Allowed | `acl/` folders can import other modules |
| Shared kernel access | ✅ Allowed | All modules can import from `internal/shared/` |
| No cyclic dependencies | ❌ Error | Module A → B → A is forbidden |

### Understanding Violations

**Example violation:**
```
❌ internal/modules/auth/service/v1/service_v1.auth.go:5
   imports "go-pste-boilerplate/internal/modules/user/repository/sql"
   Module "auth" should not import from module "user"
```

**Fix:**
```go
// WRONG: Direct import
import userepo "go-pste-boilerplate/internal/modules/user/repository/sql"

// RIGHT: Use ACL interface
import userdomain "go-pste-boilerplate/internal/modules/user/domain"

// Implement ACL adapter
type UserCreatorAdapter struct {
    repo userdomain.Repository
}
```

---

## Debugging & Troubleshooting

### Common Issues

#### Configuration Not Loading

```bash
# Check config path
go run . -config=./config/config.yaml

# Verify YAML syntax
yamllint config/config.yaml

# Check file permissions
ls -la config/config.yaml
```

#### Database Connection Failed

```bash
# Test PostgreSQL connection
psql postgresql://user:password@localhost:5432/appdb

# Test MongoDB connection
mongosh mongodb://localhost:27017

# Check connection string in config
cat config/config.yaml
```

#### Feature Flag Not Working

```bash
# Verify feature flag file exists
cat config/featureflags.yaml

# Check flag name (case-sensitive)
grep "http_handler:" config/featureflags.yaml

# Verbose container logs
go run . server -v
```

#### Linter Violations

```bash
# Run linter with verbose output
go run cmd/lint-deps/main.go -v

# Check import statements
grep -r "internal/modules/" internal/modules/ | grep -v "/domain/" | grep -v "/acl/"
```

### Debugging Tips

**Enable structured logging:**
```go
import "github.com/rs/zerolog/log"

log.Debug().Str("module", "product").Msg("creating product")
```

**Print dependency status:**
```go
go run . -print-container  // (if implemented)
```

**Trace HTTP requests:**
```yaml
# config/config.yaml
app:
  http:
    debug: true
    log_requests: true
```

### Useful Commands

```bash
# Check Go version
go version

# Verify module dependencies
go mod tidy && go mod verify

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Profile memory
go run . -memprofile=mem.prof

# Profile CPU
go run . -cpuprofile=cpu.prof
```

---

## Best Practices

### 1. Module Independence

✅ **DO:**
- Each module has its own domain types
- Modules communicate via ACL or events
- No shared database tables across modules

❌ **DON'T:**
- Import another module's handler/service
- Share domain types across modules
- Use one module's repository in another module

### 2. Error Handling

✅ **DO:**
```go
// Use shared error types
if err != nil {
    return sharederrors.NotFound("product", id)
}

// Handle specific errors
if err == sql.ErrNoRows {
    return sharederrors.NotFound("entity", id)
}
```

❌ **DON'T:**
```go
// Avoid generic errors
return fmt.Errorf("something went wrong")

// Avoid panics
panic(err)
```

### 3. Context Usage

✅ **DO:**
```go
func (s *Service) Process(ctx context.Context, id string) error {
    // Pass context to repository
    return s.repo.GetByID(ctx, id)
}
```

❌ **DON'T:**
```go
func (s *Service) Process(req *Request) error {
    // Don't use background context
    return s.repo.GetByID(context.Background(), req.ID)
}
```

### 4. Validation

✅ **DO:**
```go
// Validate at handler boundary
if err := validator.Validate(req); err != nil {
    return err
}
```

❌ **DON'T:**
```go
// Don't validate in service
type Service struct {
    validator validator.Validator
}
```

### 5. Caching

✅ **DO:**
```go
// Cache in service layer
cached, err := s.cache.Get(ctx, id)
if err == nil {
    return cached
}
```

❌ **DON'T:**
- Cache at handler level (tight coupling)
- Cache sensitive data without TTL

### 6. Logging

✅ **DO:**
```go
log.Info().
    Str("module", "product").
    Str("action", "create").
    Str("id", product.ID).
    Msg("product created")
```

❌ **DON'T:**
```go
// Avoid fmt.Println in production
fmt.Println("Creating product:", id)
```

### 7. Testing

✅ **DO:**
- Test business logic in service layer
- Mock repositories and external services
- Use table-driven tests for multiple scenarios
- Test error cases

❌ **DON'T:**
- Test implementation details
- Write integration tests as unit tests
- Skip error cases

### 8. Feature Flags

✅ **DO:**
- Create new versions (v2) instead of modifying v1
- Keep old implementations for backwards compatibility
- Use noop for disabled features

❌ **DON'T:**
- Modify existing implementations with flags
- Use flags for feature branching (use git branches)

### 9. Database Queries

✅ **DO:**
```go
// Use parameterized queries
query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRowContext(ctx, query, id)

// Handle context cancellation
if ctx.Err() != nil {
    return ctx.Err()
}
```

❌ **DON'T:**
```go
// String concatenation (SQL injection risk)
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)

// Ignore context timeouts
db.QueryRow(query, id)
```

### 10. Configuration

✅ **DO:**
- Load configuration at startup
- Use environment variables for sensitive data
- Validate configuration before use

❌ **DON'T:**
- Hardcode configuration values
- Change configuration at runtime
- Mix configuration sources

---

## Additional Resources

- [Technical Documentation](./TECHNICAL_DOCUMENTATION.md)
- [API Reference](./TECHNICAL_DOCUMENTATION.md#api-reference)
- [Worker Implementation](./WORKER_IMPLEMENTATION.md)
- [Architecture Patterns](./TECHNICAL_DOCUMENTATION.md#architecture)
- [Go Modules](https://go.dev/blog/using-go-modules)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/)

---

## Getting Help

- **Issues:** Create an issue on [GitHub](https://github.com/kamil5b/go-pste-boilerplate/issues)
- **Documentation:** See [docs/](./TECHNICAL_DOCUMENTATION.md)
- **Examples:** Check existing modules in `internal/modules/`
- **Scripts:** Use `scripts/new-module.sh` to generate module template

---

**Happy coding! 🚀**
