# Example: Monthly Email on the 15th

This guide shows how to implement a real-world task: sending an email to all users on the 15th of every month with a specific message.

## Setup Overview

1. **Scheduler**: Enqueues task on 15th each month
2. **Worker Server**: Receives and processes the task
3. **Retry Policy**: Handles failures with exponential backoff
4. **DLQ**: Captures permanently failed tasks for investigation

## Implementation

### Step 1: Configure in Bootstrap

In `cmd/bootstrap/bootstrap.worker.go`:

```go
package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/email/smtp"
	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/redpanda"
	userdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
)

// RunWorkerWithScheduler initializes worker and cron scheduler
func RunWorkerWithScheduler(cfg *Config, userRepo userdomain.Repository, emailService *smtp.SMTPEmailService) error {
	brokers := []string{cfg.Kafka.Brokers}
	
	// Create worker server
	server := redpanda.NewRedpandaServer(brokers, "tasks", "worker-group", 4)
	
	// Configure retry policy
	server.SetRetryPolicy(infraworker.RetryPolicy{
		MaxRetries:        5,
		InitialBackoff:    2 * time.Second,
		MaxBackoff:        2 * time.Minute,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
		NonRetryableErrors: []string{
			"validation error",
			"user not found",
		},
	})
	
	// Register handler
	server.RegisterHandler("send_monthly_email", 
		newMonthlyEmailHandler(userRepo, emailService))
	
	// Start server in background
	go func() {
		if err := server.Start(context.Background()); err != nil {
			log.Printf("Worker error: %v\n", err)
		}
	}()
	
	// Create and start scheduler
	client := redpanda.NewRedpandaClient(brokers, "tasks")
	scheduler := infraworker.NewCronScheduler(client)
	
	// Schedule task for 15th at 9 AM every month
	scheduler.AddJob(
		"monthly_user_email",
		"send_monthly_email",
		sharedworker.Monthly(15, 9, 0),
		sharedworker.TaskPayload{"message": "Today is the day"},
	)
	
	// Start scheduler in background
	go func() {
		if err := scheduler.Start(context.Background()); err != nil {
			log.Printf("Scheduler error: %v\n", err)
		}
	}()
	
	return nil
}

// Handler function
func newMonthlyEmailHandler(
	userRepo userdomain.Repository, 
	emailService *smtp.SMTPEmailService,
) sharedworker.TaskHandler {
	return func(ctx context.Context, payload sharedworker.TaskPayload) error {
		message, ok := payload["message"].(string)
		if !ok {
			return fmt.Errorf("validation error: message not found in payload")
		}
		
		// Get all users
		users, err := userRepo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch users: %w", err)
		}
		
		// Send email to each user
		successCount := 0
		for _, user := range users {
			// Build email
			email := &email.Email{
				To:       []string{user.Email},
				From:     "noreply@example.com",
				Subject:  "Monthly Notification",
				TextBody: fmt.Sprintf("Hello %s,\n\n%s\n\nBest regards,\nOur Team", user.Name, message),
				HTMLBody: fmt.Sprintf(
					`<h2>Hello %s,</h2><p>%s</p><p>Best regards,<br/>Our Team</p>`,
					user.Name, message,
				),
			}
			
			// Send
			if err := emailService.Send(ctx, email); err != nil {
				log.Printf("Failed to send email to %s: %v\n", user.Email, err)
				// Continue with other users, but track failure
				continue
			}
			
			successCount++
		}
		
		// If no emails sent successfully, return error for retry
		if successCount == 0 && len(users) > 0 {
			return fmt.Errorf("failed to send any emails (attempted %d)", len(users))
		}
		
		log.Printf("Sent monthly email to %d users\n", successCount)
		return nil
	}
}
```

### Step 2: Cron Expression Details

```go
import sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

// Monthly on the 15th at 9:00 AM
sharedworker.Monthly(15, 9, 0)

// Parameters:
// - 15: Day of month
// - 9: Hour (0-23, so 9 = 9 AM)
// - 0: Minute
```

Other useful expressions:

```go
// Daily at 8 AM
sharedworker.Daily(8, 0)

// Every Monday at 10 AM
sharedworker.Weekly(1, 10, 0)  // 0=Sun, 1=Mon, ..., 6=Sat

// Every hour
sharedworker.EveryHour()

// Every minute
sharedworker.EveryMinute()
```

### Step 3: Error Handling & Retries

The handler will:

1. **First Attempt** (9:00 AM on 15th)
   - If fails, waits 2 seconds, tries again
   
2. **Retry 1** (9:00:02 AM)
   - If fails, waits 4 seconds, tries again
   
3. **Retry 2** (9:00:06 AM)
   - If fails, waits 8 seconds, tries again
   
4. **Retry 3** (9:00:14 AM)
   - If fails, waits 16 seconds, tries again
   
5. **Retry 4** (9:00:30 AM)
   - If fails, waits 32 seconds, tries again
   
6. **Retry 5** (9:01:02 AM)
   - If fails, moves to DLQ for manual review

### Step 4: Monitoring

Track what happens:

```go
// In your monitoring/logging system
// Subscribe to DLQ topic: "tasks-dlq"

// Each failed message includes:
- error: error message
- retry_count: number of attempts (should be 5)
- original_offset: original message ID
- correlation_id: for tracing through logs
- metadata: JSON with timing and history
- dlq_timestamp: when it failed finally
```

## Database Query Example

If getting users from database:

```go
// In user repository
func (r *Repository) List(ctx context.Context) ([]*userdomain.User, error) {
	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Your database query
	rows, err := r.db.QueryContext(ctx, "SELECT id, email, name FROM users WHERE active = true")
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()
	
	var users []*userdomain.User
	for rows.Next() {
		user := &userdomain.User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.Name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		users = append(users, user)
	}
	
	return users, rows.Err()
}
```

## Testing the Scheduler

Create a test that verifies the job is scheduled:

```go
import (
	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/redpanda"
)

func TestMonthlyEmailScheduling(t *testing.T) {
	brokers := []string{"localhost:9092"}
	client := redpanda.NewRedpandaClient(brokers, "tasks")
	scheduler := infraworker.NewCronScheduler(client)
	
	// Add job
	err := scheduler.AddJob(
		"test_monthly_email",
		"send_monthly_email",
		sharedworker.Monthly(15, 9, 0),
		sharedworker.TaskPayload{"message": "Test message"},
	)
	assert.NoError(t, err)
	
	// Verify job exists
	jobs := scheduler.ListJobs()
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test_monthly_email", jobs[0].ID)
	assert.True(t, jobs[0].Enabled)
	
	// Next run should be in the future
	assert.True(t, jobs[0].NextRun.After(time.Now()))
}
```

## Troubleshooting

### Task Not Running

1. Check logs: `docker logs worker-container`
2. Verify Kafka/Redpanda is running
3. Ensure cron expression matches current time
4. Check if job is enabled: `scheduler.EnableJob("monthly_user_email")`

### Too Many Retries

1. Reduce `MaxRetries` in RetryPolicy
2. Check if error is in `NonRetryableErrors` list
3. Verify service dependency is available
4. Check DLQ for root cause

### Wrong Time

1. Verify server timezone matches cron calculation
2. Check `worker.Monthly(15, 9, 0)` parameters
3. Remember hour is in 24-hour format (9 = 9 AM, 21 = 9 PM)

## Alternative Worker Backends

The same setup works with RabbitMQ or Asynq:

```go
// RabbitMQ
amqpURL := "amqp://user:pass@localhost:5672/"
server := rabbitmq.NewRabbitMQServer(amqpURL, "tasks", "tasks", 10)
client := rabbitmq.NewRabbitMQClient(amqpURL, "tasks", "tasks")

// Asynq (Redis-based)
redisAddr := "localhost:6379"
server := asynq.NewAsynqServer(redisAddr, "", 10)
client := asynq.NewAsynqClient(redisAddr, "")
```

The cron scheduler and retry policy work identically with all backends!
