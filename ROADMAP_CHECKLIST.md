# Roadmap – Project Checklist

## Completed ✅

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

## In Progress 🚧

- [ ] Unit Tests (Priority: High)

## Planned 📋

- [ ] Worker support: Asynq, RabbitMQ, Redpanda
- [ ] Storage support: S3-Compatible, GCS, MinIO, Local, etc
- [ ] gRPC & Protocol Buffers support
- [ ] WebSocket integration
- [ ] OpenTelemetry integration for distributed tracing
- [ ] Database-per-module schema separation
- [ ] API Gateway setup (Kong/Traefik)
- [ ] Kubernetes deployment manifests