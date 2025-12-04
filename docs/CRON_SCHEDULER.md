# Cron Scheduler & Retry Policy

This package provides production-ready cron scheduling and retry mechanisms for task queues.

## Features

- **Cron Scheduler**: Schedule recurring tasks (hourly, daily, weekly, monthly)
- **Retry Policy**: Configurable exponential backoff with jitter
- **Support for**: Redpanda/Kafka, RabbitMQ, Asynq
- **Metrics**: Track retry attempts and task execution
- **Dead-Letter Queue**: Failed tasks with metadata for debugging

## Quick Example: Daily Email on the 15th

Here's how to set up a task that sends an email to users every 15th of the month:

### Step 1: Initialize the Scheduler

```go
package main

import (
	"context"
	"log"

	sharedworker "go-modular-monolith/internal/shared/worker"
	infraworker "go-modular-monolith/internal/infrastructure/worker"
	"go-modular-monolith/internal/infrastructure/worker/redpanda"
)

func main() {
	// Create Redpanda client
	brokers := []string{"localhost:9092"}
	client := redpanda.NewRedpandaClient(brokers, "tasks")
	
	// Create scheduler
	scheduler := infraworker.NewCronScheduler(client)
	
	// Schedule task: Send email on 15th of every month at 9 AM
	scheduler.AddJob(
		"monthly_user_email",                           // Job ID
		"send_user_notification",                       // Task name
		sharedworker.Monthly(15, 9, 0),                 // 15th at 9:00 AM
		sharedworker.TaskPayload{"message": "Today is the day"}, // Payload
	)
	
	// Start scheduler in background
	go scheduler.Start(context.Background())
	
	// Keep running
	select {}
}
```

### Step 2: Register Task Handler

```go
package main

import (
	"context"
	"fmt"
	"log"

	sharedworker "go-modular-monolith/internal/shared/worker"
	"go-modular-monolith/internal/infrastructure/worker/redpanda"
)

func main() {
	brokers := []string{"localhost:9092"}
	
	// Create server to process tasks
	server := redpanda.NewRedpandaServer(brokers, "tasks", "worker-group", 1)
	
	// Register handler for the task
	server.RegisterHandler("send_user_notification", func(ctx context.Context, payload sharedworker.TaskPayload) error {
		message, ok := payload["message"].(string)
		if !ok {
			return fmt.Errorf("invalid message in payload")
		}
		
		// Send email to users
		users := getAllUsers(ctx) // Your function to get users
		
		for _, user := range users {
			// Send email using email service
			log.Printf("Sending email to %s: %s\n", user.Email, message)
		}
		
		return nil
	})
	
	// Start server
	if err := server.Start(context.Background()); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
```

## Cron Expression Examples

```go
import sharedworker "go-modular-monolith/internal/shared/worker"

// Every minute
sharedworker.EveryMinute()

// Every hour at minute 0
sharedworker.EveryHour()

// Daily at 8:00 AM
sharedworker.Daily(8, 0)

// Weekly on Monday at 9:00 AM
sharedworker.Weekly(1, 9, 0)  // 0=Sunday, 1=Monday, ..., 6=Saturday

// Monthly on 15th at 9:00 AM
sharedworker.Monthly(15, 9, 0)
```

## Retry Policy Configuration

### Default Retry Policy

```go
import infraworker "go-modular-monolith/internal/infrastructure/worker"

policy := infraworker.DefaultRetryPolicy()
// MaxRetries: 3
// InitialBackoff: 1 second
// MaxBackoff: 60 seconds
// BackoffMultiplier: 2.0 (exponential)
// JitterFraction: 0.1 (10% jitter)
```

### Custom Retry Policy

```go
import infraworker "go-modular-monolith/internal/infrastructure/worker"

server.SetRetryPolicy(infraworker.RetryPolicy{
	MaxRetries:         5,
	InitialBackoff:     500 * time.Millisecond,
	MaxBackoff:         5 * time.Minute,
	BackoffMultiplier:  2.0,
	JitterFraction:     0.1,
	NonRetryableErrors: []string{
		"validation error",
		"unauthorized",
		"not found",
	},
	RetryableErrors: []string{
		"timeout",
		"connection refused",
		"unavailable",
	},
})
```

## Backoff Schedule Example

With default policy (MaxRetries=3, InitialBackoff=1s, Multiplier=2.0):

- Attempt 1: Fails → Wait 1s (±10% jitter)
- Attempt 2: Fails → Wait 2s (±10% jitter)
- Attempt 3: Fails → Wait 4s (±10% jitter)
- Attempt 4: Fails → Send to DLQ

## Worker Implementations

All examples work with any worker backend:

### Redpanda/Kafka

```go
client := redpanda.NewRedpandaClient(brokers, "tasks")
server := redpanda.NewRedpandaServer(brokers, "tasks", "worker-group", 1)
```

### RabbitMQ

```go
client := rabbitmq.NewRabbitMQClient(url, exchange, queue)
server := rabbitmq.NewRabbitMQServer(url, exchange, queue, prefetch)
```

### Asynq

```go
client := asynq.NewAsynqClient(redisAddr, redisPassword)
server := asynq.NewAsynqServer(redisAddr, redisPassword, concurrency)
```

## Dead-Letter Queue

Failed tasks after max retries are sent to a dead-letter queue with full metadata:

```
Topic: {topic}-dlq
Headers:
  - error: error message
  - retry_count: number of attempts
  - correlation_id: for tracing
  - original_offset: original message offset
  - metadata: JSON with processing history
```

## Monitoring & Metrics

Track task execution:

```go
metrics := worker.NewRetryMetrics("send_user_notification")
// Access:
// - metrics.RetryCount
// - metrics.SuccessfulAt
// - metrics.LastError
// - metrics.ProcessingSteps
```

## Production Checklist

- [ ] Configure retry policy based on task type
- [ ] Set up DLQ monitoring and alerting
- [ ] Configure max concurrent jobs per scheduler
- [ ] Test cron expressions before deploying
- [ ] Monitor task latency and failure rates
- [ ] Set up correlation IDs for tracing
- [ ] Configure logging for debugging
