# Go Modular Monolith

A practical guide and starter layout for building a **Modular Monolith** in Go using **Echo**. Includes: authentication (JWT), middleware (logger, recovery), REST API, gRPC (push/pull), CRUD with PostgreSQL, optional MongoDB support, migrations, Docker, and project conventions.

---

## Goals

* Keep code modular by domain (feature) while running as a single process.
* Clear separation: `api` (http), `grpc`, `service` (business logic), `repo` (persistence), `pkg` (shared utilities).
* Production-ready pieces: logging, panic recovery, configuration, graceful shutdown, and pluggable persistence.

---

## Recommended tech choices

* Framework: `github.com/labstack/echo/v4` (HTTP)
* gRPC: `google.golang.org/grpc`
* DB: PostgreSQL with `github.com/jmoiron/sqlx` (or `gorm`), optional MongoDB with `go.mongodb.org/mongo-driver`
* Migrations: `pressly/goose` (simple SQL up/down) or `github.com/golang-migrate/migrate/v4` — both supported in examples
* JWT: `github.com/golang-jwt/jwt/v5`
* Logger: `github.com/rs/zerolog` (structured, fast)
* Dependency injection: simple constructor injection (no heavy DI framework)

---

## Project layout (modular monolith)

```
cmd/
    server/
        main.go
internal/
    auth/
        service.go
        middleware.go
    product/             # domain module example
        handler.go         # HTTP handlers
        repository.go      # SQL repository (sqlx)
        repository_mongo.go# Mongo repository (mongo-driver)
        service.go         # business logic (uses Repository interface)
        model.go
    article/             # another domain
pkg/
    config/
    logger/
    db/
    grpcclient/
api/
    proto/
migrations/
    goose/               # goose-compatible SQL files (up/down)
    postgres/            # optional raw .up.sql files
    mongo/               # mongo shell scripts
Dockerfile
docker-compose.yml
Makefile
```

`internal/` keeps domain code private to the project.

---

## Configuration

Use environment variables (examples below). The project supports selecting which persistence to use at runtime via `DB_TYPE`.

Key env vars (examples):

- `PORT` (default `8080`)
- `DATABASE_URL` — Postgres DSN (used for SQL mode and/or auth storage)
- `DB_TYPE` — `postgres` (default) or `mongo`
- `MONGO_URL` — Mongo connection string (default `mongodb://localhost:27017`)
- `MONGO_DB` — Mongo database name (default `app`)
- `JWT_SECRET` — secret for signing JWTs

Example `pkg/config/config.go` (high level):

```go
type Config struct {
        Port      string
        GRPCPort  string
        DBUrl     string
        DBType    string // "postgres" or "mongo"
        MongoURL  string
        MongoDB   string
        JWTSecret string
}

// Load reads env vars with sensible defaults.
```

Behavior notes:

- If `DB_TYPE=mongo`, the product repo uses MongoDB; the server will still try to open Postgres (if `DATABASE_URL` provided) for the `auth` module. If Postgres isn't available, auth routes are disabled (logged) — you can change this to fail-fast if desired.
- If `DB_TYPE=postgres` (default), the app uses Postgres for products and auth.

---

## Logger & panic recovery middleware

`pkg/logger/logger.go` (zerolog):

```go
package logger

import (
        "os"
        "time"

        "github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
        zerolog.TimeFieldFormat = time.RFC3339
        Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
```

Use Echo's `Recover` middleware and adapt request logging to zerolog for structured HTTP logs.

---

## Authentication (JWT)

`internal/auth` implements JWT auth backed by Postgres (sqlx) in the starter. It exposes:

- `service.go`: create/validate tokens, register/activate, forgot/reset password logic stored in Postgres.
- `middleware.go`: `JWTMiddleware(s *Service) echo.MiddlewareFunc` — extracts `Authorization: Bearer <token>`, validates, and sets `user_id` in the context.
- `handler.go`: HTTP handlers for register/activate/login/forgot/reset.

Notes:

- The current auth implementation stores users in Postgres. If you want full Mongo-only deployment, add a Mongo-backed `auth` implementation (similar pattern to `product/repository_mongo.go`) and wire it when `DB_TYPE=mongo`.

---

## Database layer and pluggable repositories

This starter uses interfaces to allow swapping storage implementations.

- Define a `Repository` interface in each domain (e.g., `internal/product/service.go`) with CRUD methods.
- Provide an SQL implementation (`internal/product/repository.go`) using `sqlx` and a Mongo implementation (`internal/product/repository_mongo.go`) using the official `mongo-driver`.
- `cmd/server/main.go` decides which implementation to wire based on `DB_TYPE` and available connections.

SQL helper (pkg/db/db.go):

```go
func Open(dsn string) (*sqlx.DB, error) { ... }
```

Mongo helper (pkg/db/mongo.go):

```go
func OpenMongo(uri string) (*mongo.Client, error) { ... }
func CloseMongo(client *mongo.Client) error { ... }
```

---

## Migrations

Two workflows are available:

- SQL migrations: using `pressly/goose` (goose-style Up/Down SQL files are provided under `migrations/goose`). A `Makefile` target is included to install and run goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
DATABASE_URL="postgres://postgres:pass@localhost:5432/app?sslmode=disable" make migrate-up
```

- Mongo migrations: simple `mongosh` scripts in `migrations/mongo/` that create collections, JSON schema validators and indexes. Run with:

```bash
mongosh "mongodb://localhost:27017/app" migrations/mongo/0002_create_products_collection.js
```

Notes:

- `migrations/postgres/` still contains plain `.up.sql` files (can be used with `golang-migrate` if preferred).
- For production, integrate migrations into your CI/CD pipeline and consider idempotency and transactional safety.

---

## REST API with Echo

Use Echo to expose domain handlers. Example wiring in `cmd/server/main.go`:

```go
e := echo.New()
e.Use(middleware.Recover())
e.Use(middleware.Logger())

// auth routes (registered only if auth service is available)
// product routes use a Repository implementation chosen at startup
```

Product endpoints (examples):

- `POST /v1/products` — create (sets `created_by` from token `sub` when available)
- `GET /v1/products` — list (excludes soft-deleted)
- `GET /v1/products/:id` — detail
- `PUT /v1/products/:id` — update (soft metadata)
- `DELETE /v1/products/:id` — soft delete

---

## Graceful shutdown

`cmd/server/main.go` includes signal handling and a timeout context for graceful shutdown. On SIGINT/SIGTERM it:

- calls `e.Shutdown(ctx)` to stop the HTTP server
- closes SQL connections (`db.Close()`) and/or Mongo client (`CloseMongo`)
- logs shutdown steps

This ensures in-flight requests are drained before exit (within the configured timeout).

---

## gRPC (push/pull) — overview

Keep gRPC services under `internal/*/grpc.go`. Run gRPC server in a goroutine alongside HTTP and include it in graceful shutdown.

---

## Testing strategy

* Unit tests for services and repositories (use `sqlmock` for SQL, in-memory fixtures or a test Mongo instance for Mongo).
* Integration tests against test Postgres/Mongo (Docker) and real gRPC server.
* End-to-end tests via `docker-compose` with migrations run at startup.

---

## Docker and docker-compose

The `docker-compose.yml` provided runs Postgres + app by default. To test Mongo mode, add a `mongo` service and set `DB_TYPE=mongo`, `MONGO_URL` and `MONGO_DB` in the `app` environment.

Example: add a `mongo` service to `docker-compose` and wire `MONGO_URL` to `mongodb://mongo:27017`.

---

## Useful commands / dev flow

```bash
# fetch deps
go mod tidy

# run SQL migrations (goose)
DATABASE_URL="$DATABASE_URL" make migrate-up

# run Mongo migration
mongosh "$MONGO_URL/$MONGO_DB" migrations/mongo/0002_create_products_collection.js

# run locally (choose DB_TYPE and envs)
DB_TYPE=postgres DATABASE_URL="..." JWT_SECRET=devsecret go run ./cmd/server

DB_TYPE=mongo MONGO_URL="mongodb://localhost:27017" MONGO_DB=app JWT_SECRET=devsecret go run ./cmd/server
```

---

## Where to go next / extensions

* Add Mongo-backed auth implementation to support Mongo-only deployments.
* Add RBAC/ACL, rate limiting, request size limits.
* Observability: Prometheus metrics, distributed tracing (OpenTelemetry).
* CI/CD: run migrations and health checks in deploy pipelines.

---

