# Worker Bootstrap Implementation

## Overview

The worker bootstrap system enables running background task processing using different backend implementations (Asynq, RabbitMQ, Redpanda, or no-op).

## Architecture

The worker system is organized with a clean separation of concerns:

```
┌──────────────────────────────────────────────────┐
│        cmd/bootstrap/bootstrap.worker.go         │
│           (Worker initialization)                │
└────────────────────┬─────────────────────────────┘
                     │
        ┌────────────▼──────────────┐
        │  internal/app/worker/     │
        │  ├── manager.go           │
        │  │   (WorkerManager)      │
        │  ├── registrar.go         │
        │  │   (ModuleRegistry)     │
        │  └── (TaskRegistry)       │
        └────────────┬──────────────┘
                     │
        ┌────────────▼──────────────────┐
        │  Module Task Registrars       │
        │ ├── user/worker/registrar.go  │
        │ │   (UserModuleRegistrar)    │
        │ └── [product/worker/...]      │
        └────────────┬──────────────────┘
                     │
        ┌────────────▼────────────────────┐
        │  Module Task Handlers           │
        │ ├── user/worker/handlers.go     │
        │ └── [product/worker/handlers.go]│
        └────────────┬────────────────────┘
                     │
        ┌────────────▼──────────────────────┐
        │  Worker Server (Infrastructure)   │
        │  (Asynq/RabbitMQ/Redpanda/NoOp)   │
        └───────────────────────────────────┘
```

## Files

### Shared Worker Types (`internal/shared/worker/`)

#### `worker.go` - Shared Types & Interfaces
Defines all types and interfaces used across the application:
- **TaskPayload** - Map type for task data
- **TaskHandler** - Function type for processing tasks
- **TaskDefinition** - Task with name and handler
- **CronJobDefinition** - Cron job with schedule
- **CronExpression** - Simplified cron expression
- **Client** - Interface for enqueueing tasks
- **Server** - Interface for processing tasks
- **Scheduler** - Interface for scheduling recurring tasks
- Option types: Priority, MaxRetries, Timeout, Queue

### Core Worker System (`internal/app/worker/`)

#### 1. `manager.go` - WorkerManager
Handles:
- Worker server lifecycle (start, stop)
- Task registration coordination
- Logging and status reporting

Key classes:
- **WorkerManager** - Main coordinator for worker operations
- **TaskRegistry** - Collects and registers tasks with the worker server

#### 2. `registrar.go` - ModuleRegistry  
Handles:
- Module task registration coordination
- Defines ModuleTaskRegistrar interface for modules to implement

Key classes:
- **ModuleTaskRegistrar** - Interface that modules implement
- **ModuleRegistry** - Manages multiple module registrars

### Module Task Registrars

#### `internal/modules/user/worker/registrar.go` - UserModuleRegistrar
Implements `ModuleTaskRegistrar` and:
- Registers user module tasks based on feature flags
- Creates user worker handlers
- Registers: welcome email, password reset email, data export, report generation

## Implementation Flow

```
1. RunWorker() - cmd/bootstrap/bootstrap.worker.go
   │
   ├─ Load config & feature flags
   ├─ Initialize databases
   ├─ Create container with dependencies
   │
   ├─ Create WorkerManager(container)
   │
   ├─ Create ModuleRegistry
   │  └─ Register UserModuleRegistrar
   │
   ├─ Call ModuleRegistry.RegisterAll()
   │  └─ For each module:
   │     └─ module.RegisterTasks(taskRegistry, container, featureFlags)
   │        └─ UserModuleRegistrar checks feature flags
   │           └─ Registers only enabled tasks to taskRegistry
   │
   ├─ Call WorkerManager.RegisterTasks()
   │  └─ taskRegistry.RegisterAll(server)
   │     └─ For each task: server.RegisterHandler(taskName, handler)
   │
   ├─ Start WorkerManager
   │  └─ Start worker server
   │
   └─ Wait for signals & graceful shutdown
```

## How to Add a New Module with Worker Tasks

### Step 1: Create Task Definitions

Create `internal/modules/mymodule/worker/tasks.go`:

```go
package worker

const (
    TaskMyTask = "mymodule:my_task"
)

type MyTaskPayload struct {
    ID string `json:"id"`
}
```

### Step 2: Create Task Handlers

Create `internal/modules/mymodule/worker/handlers.go`:

```go
package worker

import (
    "context"
    sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

type MyModuleWorkerHandler struct {
    // dependencies...
}

func (h *MyModuleWorkerHandler) HandleMyTask(ctx context.Context, payload sharedworker.TaskPayload) error {
    // Implement task logic
    return nil
}
```

### Step 3: Create Module Registrar

Create `internal/modules/mymodule/worker/registrar.go`:

```go
package worker

import (
    "github.com/kamil5b/go-ptse-monolith/internal/app/core"
    appworker "github.com/kamil5b/go-ptse-monolith/internal/app/worker"
)

type MyModuleRegistrar struct{}

func NewMyModuleRegistrar() *MyModuleRegistrar {
    return &MyModuleRegistrar{}
}

func (r *MyModuleRegistrar) RegisterTasks(
    registry *appworker.TaskRegistry,
    container *core.Container,
    featureFlags *core.FeatureFlag,
) error {
    handler := NewMyModuleWorkerHandler(
        container.MyRepository,
        // other dependencies...
    )

    if featureFlags.Worker.Tasks.MyTask {
        registry.Register(TaskMyTask, handler.HandleMyTask)
    }

    return nil
}
```

### Step 4: Register in Bootstrap

Update `cmd/bootstrap/bootstrap.worker.go`:

```go
moduleRegistry := worker.NewModuleRegistry()
moduleRegistry.Register(userworker.NewUserModuleRegistrar())
moduleRegistry.Register(mymoduleworker.NewMyModuleRegistrar())  // Add this
```

## Configuration

### Enable Workers

Edit `config/featureflags.yaml`:

```yaml
worker:
  enabled: true
  backend: "asynq"  # asynq, rabbitmq, redpanda, disable
  tasks:
    email_notifications: true
    data_export: true
    report_generation: true
    image_processing: false
```

### Configure Backend

Edit `config/config.yaml`:

```yaml
app:
  worker:
    enabled: true
    backend: "asynq"
    
    asynq:
      redis_url: "redis://localhost:6379"
      concurrency: 10
    
    # OR for RabbitMQ:
    rabbitmq:
      url: "amqp://guest:guest@localhost:5672/"
      exchange: "tasks"
      queue: "tasks_queue"
      worker_count: 10
    
    # OR for Redpanda:
    redpanda:
      brokers:
        - "localhost:9092"
      topic: "tasks"
      consumer_group: "workers"
      worker_count: 10
```

## Usage

### Run Worker

```bash
go run . worker
```

### Enqueue Tasks

```go
payload := map[string]interface{}{
    "user_id": userID,
    "email":   email,
}

container.WorkerClient.Enqueue(
    ctx,
    userworker.TaskSendWelcomeEmail,
    payload,
)
```

## Supported Tasks

| Module | Task | Feature Flag |
|--------|------|-------------|
| User | TaskSendWelcomeEmail | email_notifications |
| User | TaskSendPasswordResetEmail | email_notifications |
| User | TaskExportUserData | data_export |
| User | TaskGenerateUserReport | report_generation |

## Key Design Patterns

### 1. Module Task Registrar Pattern
Each module implements `ModuleTaskRegistrar` interface to register its own tasks. This:
- Keeps task logic with the module
- Allows feature flag-based registration
- Makes adding new modules simple

### 2. Task Registry Pattern
A centralized registry collects all tasks before registering with the worker server. This:
- Separates task collection from registration
- Allows for pre-registration hooks
- Provides clear visibility of all registered tasks

### 3. Graceful Shutdown
The worker listens for SIGINT/SIGTERM and:
- Stops accepting new tasks
- Waits up to 30 seconds for in-flight tasks
- Closes all connections cleanly

## Logging Output

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

## Production Deployment

1. **Separate Instances**: Run workers in dedicated containers
2. **Scaling**: Deploy multiple worker instances for throughput
3. **Monitoring**: Log all task executions, errors, timings
4. **Health Checks**: Implement health check endpoints
5. **Dead Letter Queues**: Configure for backend (backend-specific)

## Troubleshooting

**Worker won't start:**
- Check `config/featureflags.yaml`: `enabled: true`
- Check `config/config.yaml`: Worker backend configuration
- Verify backend service is running (Redis, RabbitMQ, Redpanda)

**Tasks not processing:**
- Check task feature flag is enabled
- Check task handler logs
- Verify payload structure

**Connection errors:**
- Verify backend service connectivity
- Check connection strings in config
- Check firewall/network access
