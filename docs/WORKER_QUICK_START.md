# Worker Bootstrap - Quick Start Guide

## What Was Implemented

✅ **Worker Manager** (`internal/app/worker/manager.go`)
- Orchestrates worker lifecycle (start, stop)
- Coordinates task registration
- Provides clean API for bootstrap

✅ **Module Registry Pattern** (`internal/app/worker/registrar.go`)
- Allows modular task registration
- Modules implement `ModuleTaskRegistrar` interface
- Feature flag-based task activation

✅ **User Module Registrar** (`internal/modules/user/worker/registrar.go`)
- Registers user module tasks
- Respects feature flags for selective registration
- Manages handler lifecycle

✅ **Bootstrap Integration** (updated `cmd/bootstrap/bootstrap.worker.go`)
- Uses new worker manager
- Coordinates module registrations
- Handles graceful shutdown

✅ **CLI Integration** (updated `main.go`)
- New `worker` command for starting the worker

## Architecture Overview

```
┌─────────────────────────────────────┐
│  cmd/bootstrap/bootstrap.worker.go  │
└────────────┬────────────────────────┘
             │
             ├─ Create WorkerManager
             ├─ Create ModuleRegistry
             ├─ Register modules:
             │  └─ UserModuleRegistrar
             │     └─ RegisterTasks(...)
             │        └─ Check feature flags
             │        └─ Add to TaskRegistry
             │           (uses sharedworker.TaskHandler)
             │
             ├─ WorkerManager.RegisterTasks()
             │  └─ Register all tasks with server
             │
             └─ WorkerManager.Start()
                └─ Start worker server

## Package Architecture

```
internal/shared/worker/           ← Types & Interfaces (TaskPayload, TaskHandler,
│                                   Client, Server, Scheduler, CronExpression)
│
internal/infrastructure/worker/   ← Implementations (Asynq, RabbitMQ, Redpanda, NoOp)
│
internal/app/worker/              ← Manager & Registry
│
internal/modules/*/worker/        ← Module-specific handlers & registrars
```

## Quick Start

### 1. Enable Workers

Edit `config/featureflags.yaml`:
```yaml
worker:
  enabled: true
  backend: "asynq"
  tasks:
    email_notifications: true
    data_export: true
    report_generation: true
    image_processing: false
```

### 2. Configure Backend

For **Asynq** (Redis-based), edit `config/config.yaml`:
```yaml
app:
  worker:
    enabled: true
    backend: "asynq"
    asynq:
      redis_url: "redis://localhost:6379"
      concurrency: 10
```

For **RabbitMQ**:
```yaml
app:
  worker:
    backend: "rabbitmq"
    rabbitmq:
      url: "amqp://guest:guest@localhost:5672/"
      exchange: "tasks"
      queue: "tasks_queue"
```

For **Redpanda**:
```yaml
app:
  worker:
    backend: "redpanda"
    redpanda:
      brokers: ["localhost:9092"]
      topic: "tasks"
      consumer_group: "workers"
```

### 3. Start Worker

```bash
go run . worker
```

### 4. Enqueue Tasks

From your service:
```go
payload := map[string]interface{}{
    "user_id": userID,
    "email":   email,
    "name":    name,
}

container.WorkerClient.Enqueue(
    ctx,
    userworker.TaskSendWelcomeEmail,
    payload,
)
```

## Available Tasks

| Task Name | Feature Flag | Purpose |
|-----------|-------------|---------|
| `user:send_welcome_email` | `email_notifications` | Send welcome email |
| `user:send_password_reset_email` | `email_notifications` | Send password reset email |
| `user:export_user_data` | `data_export` | Export user data |
| `user:generate_user_report` | `report_generation` | Generate user report |

## Adding New Module Tasks

### Step 1: Create Task Definitions
File: `internal/modules/mymodule/worker/tasks.go`
```go
package worker

const TaskMyAction = "mymodule:my_action"

type MyActionPayload struct {
    ID string `json:"id"`
}
```

### Step 2: Create Handlers
File: `internal/modules/mymodule/worker/handlers.go`
```go
package worker

import (
    "context"
    sharedworker "go-modular-monolith/internal/shared/worker"
)

type MyModuleHandler struct {
    // dependencies
}

func (h *MyModuleHandler) HandleMyAction(
    ctx context.Context,
    payload sharedworker.TaskPayload,
) error {
    // Implement logic
    return nil
}
```

### Step 3: Create Module Registrar
File: `internal/modules/mymodule/worker/registrar.go`
```go
package worker

import (
    "go-modular-monolith/internal/app/core"
    appworker "go-modular-monolith/internal/app/worker"
)

type MyModuleRegistrar struct{}

func (r *MyModuleRegistrar) RegisterTasks(
    registry *appworker.TaskRegistry,
    container *core.Container,
    featureFlags *core.FeatureFlag,
) error {
    handler := NewMyModuleHandler(...)
    
    if featureFlags.Worker.Tasks.MyTask {
        registry.Register(TaskMyAction, handler.HandleMyAction)
    }
    
    return nil
}
```

### Step 4: Register in Bootstrap
File: `cmd/bootstrap/bootstrap.worker.go`
```go
moduleRegistry := worker.NewModuleRegistry()
moduleRegistry.Register(userworker.NewUserModuleRegistrar())
moduleRegistry.Register(mymoduleworker.NewMyModuleRegistrar())  // Add this
```

## Module Task Registrar Pattern

The `ModuleTaskRegistrar` interface allows modules to:
- Own their task definitions
- Implement their own handlers
- Control task registration via feature flags
- Maintain clean separation of concerns

```go
type ModuleTaskRegistrar interface {
    RegisterTasks(
        registry *TaskRegistry,
        container *core.Container,
        featureFlags *core.FeatureFlag,
    ) error
}
```

## Execution Flow

```
Worker Start
    ↓
Load Config & Feature Flags
    ↓
Initialize Databases
    ↓
Create Container with Dependencies
    ↓
Create WorkerManager
    ↓
Create ModuleRegistry
    ↓
Loop: For Each Module Registrar
    ├─ Call RegisterTasks()
    ├─ Module checks feature flags
    ├─ Module adds tasks to registry
    ↓
RegisterAll Tasks with Worker Server
    ↓
Start Worker Server
    ↓
Wait for Signals (SIGINT/SIGTERM)
    ↓
Graceful Shutdown (30 second timeout)
```

## Log Output

```
[INFO] Setting up task registrations...
[INFO] Registering tasks from module 1...
[INFO] Registering user email notification tasks...
[INFO] Registered handler: user:send_welcome_email
[INFO] Registered handler: user:send_password_reset_email
[INFO] Registering user data export task...
[INFO] Registered handler: user:export_user_data
[INFO] Registering user report generation task...
[INFO] Registered handler: user:generate_user_report
[INFO] Registering all tasks with worker server...
[INFO] Worker server running (backend: asynq)
```

## Key Features

✅ **Modular**: Add new modules without changing bootstrap code
✅ **Feature Flag Controlled**: Enable/disable tasks per deployment
✅ **Clean Architecture**: Each module owns its tasks
✅ **Extensible**: Easy to add new backends
✅ **Production Ready**: Graceful shutdown, error handling, logging

## Deployment Strategies

### Single Worker Instance
```bash
docker run myapp:latest ./app worker
```

### Multiple Worker Instances
```bash
docker run -d myapp:latest ./app worker  # Instance 1
docker run -d myapp:latest ./app worker  # Instance 2
docker run -d myapp:latest ./app worker  # Instance 3
```

### Separate Worker Service
```yaml
# docker-compose.yml
services:
  api:
    image: myapp:latest
    command: ./app server
    ports: ["8080:8080"]
  
  worker:
    image: myapp:latest
    command: ./app worker
    depends_on: [redis, postgres]
```

## Troubleshooting

**Worker won't start:**
- ✓ Check `config/featureflags.yaml` has `enabled: true`
- ✓ Check `config/config.yaml` has correct backend config
- ✓ Ensure backend service is running (Redis/RabbitMQ/Redpanda)

**Tasks not processing:**
- ✓ Check feature flag for task is enabled
- ✓ Check logs for handler errors
- ✓ Verify task payload structure

**Connection issues:**
- ✓ Verify backend service connectivity
- ✓ Check connection strings in config
- ✓ Check firewall/network access

## Files Changed

| File | Purpose |
|------|---------|
| `internal/app/worker/manager.go` | WorkerManager & TaskRegistry |
| `internal/app/worker/registrar.go` | ModuleRegistry & ModuleTaskRegistrar |
| `internal/modules/user/worker/registrar.go` | UserModuleRegistrar implementation |
| `cmd/bootstrap/bootstrap.worker.go` | Bootstrap using new manager |
| `main.go` | CLI worker command |
