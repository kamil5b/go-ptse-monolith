# Worker Configuration Across All Backends

This guide shows how to configure cron scheduling and retry policies for all supported worker backends.

## Common Setup Pattern

All three backends follow the same pattern:

```
1. Create Client    → Used by Scheduler to enqueue tasks
2. Create Scheduler → Manages cron jobs
3. Create Server    → Processes tasks from queue
4. Register Handler → Business logic for each task
5. Set Retry Policy → Configure exponential backoff
```

## 1. Redpanda/Kafka Setup

### Dependencies
```bash
go get github.com/segmentio/kafka-go
```

### Configuration

```go
package bootstrap

import (
	"context"
	"time"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/redpanda"
)

func setupRedpandaWorker(cfg *Config) error {
	brokers := []string{"localhost:9092"}
	
	// Create client for scheduler to enqueue tasks
	client := redpanda.NewRedpandaClient(brokers, "tasks")
	
	// Create server to process tasks
	server := redpanda.NewRedpandaServer(
		brokers,
		"tasks",          // Topic
		"worker-group",   // Consumer group
		4,                // Worker count
	)
	
	// Configure retry policy
	server.SetRetryPolicy(infraworker.RetryPolicy{
		MaxRetries:        5,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        2 * time.Minute,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	})
	
	// Register handlers
	server.RegisterHandler("send_monthly_email", handleMonthlyEmail)
	
	// Start server
	go server.Start(context.Background())
	
	// Create and start scheduler
	scheduler := infraworker.NewCronScheduler(client)
	scheduler.AddJob(
		"monthly_user_email",
		"send_monthly_email",
		sharedworker.Monthly(15, 9, 0),
		sharedworker.TaskPayload{"message": "Today is the day"},
	)
	go scheduler.Start(context.Background())
	
	return nil
}
```

### Topics Created
- `tasks`: Main task topic
- `tasks-retry`: Tasks awaiting retry
- `tasks-dlq`: Permanently failed tasks

---

## 2. RabbitMQ Setup

### Dependencies
```bash
go get github.com/rabbitmq/amqp091-go
```

### Configuration

```go
package bootstrap

import (
	"context"
	"time"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/rabbitmq"
)

func setupRabbitMQWorker(cfg *Config) error {
	amqpURL := "amqp://user:password@localhost:5672/"
	
	// Create client for scheduler to enqueue tasks
	client, err := rabbitmq.NewRabbitMQClient(amqpURL, "tasks-exchange", "tasks-queue")
	if err != nil {
		return err
	}
	
	// Create server to process tasks
	server, err := rabbitmq.NewRabbitMQServer(
		amqpURL,
		"tasks-exchange",    // Exchange
		"tasks-queue",       // Queue
		10,                  // Prefetch count
	)
	if err != nil {
		return err
	}
	
	// Configure retry policy
	server.SetRetryPolicy(infraworker.RetryPolicy{
		MaxRetries:        5,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        2 * time.Minute,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	})
	
	// Register handlers
	server.RegisterHandler("send_monthly_email", handleMonthlyEmail)
	
	// Start server
	go server.Start(context.Background())
	
	// Create and start scheduler
	scheduler := infraworker.NewCronScheduler(client)
	scheduler.AddJob(
		"monthly_user_email",
		"send_monthly_email",
		sharedworker.Monthly(15, 9, 0),
		sharedworker.TaskPayload{"message": "Today is the day"},
	)
	go scheduler.Start(context.Background())
	
	return nil
}
```

### Queues Created
- `tasks-queue`: Main task queue
- `tasks-retry-queue`: Tasks awaiting retry (with TTL)
- `tasks-dlq-queue`: Permanently failed tasks

---

## 3. Asynq (Redis-based) Setup

### Dependencies
```bash
go get github.com/hibiken/asynq
```

### Configuration

```go
package bootstrap

import (
	"context"
	"time"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/asynq"
)

func setupAsynqWorker(cfg *Config) error {
	redisAddr := "localhost:6379"
	redisPassword := ""
	
	// Create client for scheduler to enqueue tasks
	client := asynq.NewAsynqClient(redisAddr, redisPassword)
	
	// Create server to process tasks
	server := asynq.NewAsynqServer(
		redisAddr,
		redisPassword,
		10, // Concurrency
	)
	
	// Configure retry policy
	server.SetRetryPolicy(infraworker.RetryPolicy{
		MaxRetries:        5,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        2 * time.Minute,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	})
	
	// Register handlers
	server.RegisterHandler("send_monthly_email", handleMonthlyEmail)
	
	// Start server
	go server.Start(context.Background())
	
	// Create and start scheduler
	scheduler := infraworker.NewCronScheduler(client)
	scheduler.AddJob(
		"monthly_user_email",
		"send_monthly_email",
		sharedworker.Monthly(15, 9, 0),
		sharedworker.TaskPayload{"message": "Today is the day"},
	)
	go scheduler.Start(context.Background())
	
	return nil
}
```

### Redis Keys Created
- `asynq:queues:default`: Main task queue
- `asynq:queues:retry`: Retry queue
- `asynq:queues:deadletter`: Dead-letter queue

---

## Shared Handler Code

All backends use the same handler code:

```go
package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/email/smtp"
	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	userdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
)

func handleMonthlyEmail(
	userRepo userdomain.Repository,
	emailService *smtp.SMTPEmailService,
) sharedworker.TaskHandler {
	return func(ctx context.Context, payload sharedworker.TaskPayload) error {
		// Extract message from payload
		message, ok := payload["message"].(string)
		if !ok {
			return fmt.Errorf("validation error: message required")
		}
		
		// Get users
		users, err := userRepo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch users: %w", err)
		}
		
		// Send emails
		successCount := 0
		for _, user := range users {
			email := &email.Email{
				To:       []string{user.Email},
				From:     "noreply@example.com",
				Subject:  "Monthly Notification",
				TextBody: fmt.Sprintf("Hello %s,\n\n%s", user.Name, message),
				HTMLBody: fmt.Sprintf("<h2>Hello %s</h2><p>%s</p>", user.Name, message),
			}
			
			if err := emailService.Send(ctx, email); err != nil {
				log.Printf("Email failed for %s: %v", user.Email, err)
				continue
			}
			successCount++
		}
		
		if successCount == 0 && len(users) > 0 {
			return fmt.Errorf("no emails sent (attempted %d)", len(users))
		}
		
		log.Printf("Sent monthly email to %d users", successCount)
		return nil
	}
}
```

---

## Switching Backends

To change from one backend to another:

```go
// Before: Redpanda
// client := redpanda.NewRedpandaClient(brokers, "tasks")
// server := redpanda.NewRedpandaServer(brokers, "tasks", "group", 4)

// After: RabbitMQ (just swap these 2 lines)
client, _ := rabbitmq.NewRabbitMQClient(amqpURL, "exchange", "queue")
server, _ := rabbitmq.NewRabbitMQServer(amqpURL, "exchange", "queue", 10)

// Everything else (scheduler, handler, retry policy) stays the same!
```

---

## Configuration Comparison

| Feature | Redpanda | RabbitMQ | Asynq |
|---------|----------|----------|-------|
| Setup Complexity | Low | Medium | Low |
| Performance | Very High | High | High |
| Persistence | Default | Default | Configurable |
| Monitoring | Good | Excellent | Good |
| Scaling | Horizontal | Horizontal | Horizontal |
| Message TTL | Yes | Yes | Yes |
| Dead-Letter | Native | Native | Custom |

---

## Cron Schedules for 15th

All backends support the same cron expression:

```go
import sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

// 15th at 9:00 AM
sharedworker.Monthly(15, 9, 0)

// 15th at 3:30 PM
sharedworker.Monthly(15, 15, 30)

// 15th at midnight
sharedworker.Monthly(15, 0, 0)
```

---

## Monitoring Retry Flow

### Redpanda/Kafka
```
tasks → Process → Fail → tasks-retry → tasks-dlq
          ↑___________|  (backoff delay)
```

### RabbitMQ
```
tasks-queue → Process → Fail → tasks-retry-queue → tasks-dlq-queue
                         ↑_____________| (TTL delay)
```

### Asynq
```
default queue → Process → Fail → retry queue → deadletter queue
                          ↑________|  (backoff delay)
```

---

## Environment Configuration

```yaml
# config/config.yaml

worker:
  backend: "redpanda"  # or "rabbitmq", "asynq"
  enabled: true
  
redpanda:
  brokers:
    - "localhost:9092"
  topic: "tasks"
  consumer_group: "worker-group"
  worker_count: 4

rabbitmq:
  url: "amqp://user:password@localhost:5672/"
  exchange: "tasks-exchange"
  queue: "tasks-queue"
  prefetch: 10

asynq:
  redis_addr: "localhost:6379"
  redis_password: ""
  concurrency: 10

retry:
  max_retries: 5
  initial_backoff: "1s"
  max_backoff: "2m"
  multiplier: 2.0
  jitter_fraction: 0.1
```

---

## Production Deployment

1. **Staging**: Test with Redpanda (simple)
2. **Production**: Choose based on requirements
3. **Monitoring**: Set up DLQ alerts regardless of backend
4. **Scaling**: Run multiple worker instances

All implementations handle horizontal scaling automatically!
