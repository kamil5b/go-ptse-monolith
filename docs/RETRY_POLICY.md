# Retry Policy Configuration

The retry mechanism provides production-ready exponential backoff with jitter to prevent thundering herd problems.

## Retry Policy Fields

```go
type RetryPolicy struct {
	MaxRetries         int           // Maximum number of retry attempts (0 = no retries)
	InitialBackoff     time.Duration // Initial backoff duration
	MaxBackoff         time.Duration // Maximum backoff duration
	BackoffMultiplier  float64       // Exponential backoff multiplier (e.g., 2.0 for doubling)
	JitterFraction     float64       // Jitter as fraction of backoff (0.0 to 1.0)
	RetryableErrors    []string      // Specific error types to retry on (empty = all errors)
	NonRetryableErrors []string      // Error types to NOT retry
}
```

## Predefined Policies

### Default Policy

```go
import infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"

policy := infraworker.DefaultRetryPolicy()
// MaxRetries: 3
// InitialBackoff: 1 second
// MaxBackoff: 60 seconds
// BackoffMultiplier: 2.0
// JitterFraction: 0.1 (10%)
```

Backoff schedule: 1s → 2s → 4s → DLQ

## Custom Policies by Task Type

### 1. Quick Retries (Network/Temporary Errors)

```go
import infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"

// For tasks that might have temporary network issues
infraworker.RetryPolicy{
	MaxRetries:        5,
	InitialBackoff:    100 * time.Millisecond,
	MaxBackoff:        10 * time.Second,
	BackoffMultiplier: 2.0,
	JitterFraction:    0.1,
	RetryableErrors:   []string{"timeout", "unavailable", "connection"},
}
```

### 2. Long-Running Tasks (External API Calls)

```go
// For tasks calling external APIs
infraworker.RetryPolicy{
	MaxRetries:        7,
	InitialBackoff:    2 * time.Second,
	MaxBackoff:        2 * time.Minute,
	BackoffMultiplier: 1.5,
	JitterFraction:    0.2,
	NonRetryableErrors: []string{
		"404",
		"authentication failed",
		"invalid request",
	},
}
```

### 3. Database Operations

```go
// For database queries that might have deadlocks
infraworker.RetryPolicy{
	MaxRetries:        3,
	InitialBackoff:    50 * time.Millisecond,
	MaxBackoff:        5 * time.Second,
	BackoffMultiplier: 2.0,
	JitterFraction:    0.05,
	RetryableErrors:   []string{"deadlock", "connection closed"},
}
```

### 4. Email Delivery

```go
// For email tasks - be patient with mail servers
infraworker.RetryPolicy{
	MaxRetries:        10,
	InitialBackoff:    5 * time.Second,
	MaxBackoff:        5 * time.Minute,
	BackoffMultiplier: 1.3,
	JitterFraction:    0.1,
	NonRetryableErrors: []string{"invalid email", "unsubscribed"},
}
```

## How Backoff Works

### Exponential Backoff Formula

```
backoff = InitialBackoff * (BackoffMultiplier ^ (attempt - 1))
final_backoff = min(backoff, MaxBackoff)
jitter = final_backoff * JitterFraction
randomized = final_backoff ± (random * jitter)
```

### Example: Default Policy

| Attempt | Base Backoff | With Max Cap | ±Jitter (10%) | Total Range |
|---------|--------------|--------------|---------------|-------------|
| 1       | 1s           | 1s           | ±100ms        | 0.9s-1.1s   |
| 2       | 2s           | 2s           | ±200ms        | 1.8s-2.2s   |
| 3       | 4s           | 4s           | ±400ms        | 3.6s-4.4s   |
| 4       | 8s           | 8s           | ±800ms        | 7.2s-8.8s   |
| 5       | 16s          | 16s          | ±1.6s         | 14.4s-17.6s |

**Total Time**: ~34.3 seconds (before hitting DLQ)

## Error Classification

### Non-Retryable Errors

These errors indicate the task itself is invalid and won't succeed:

```go
NonRetryableErrors: []string{
	"validation error",      // Input is invalid
	"unauthorized",          // Authentication issue
	"forbidden",             // Authorization issue
	"not found",             // Resource doesn't exist
	"invalid request",       // Request format wrong
	"bad credentials",       // Permanent auth failure
	"rate limit exceeded",   // Permanent quota issue
	"request too large",     // Payload too big
	"duplicate key error",   // Database constraint
}
```

### Retryable Errors

These errors suggest a transient problem:

```go
RetryableErrors: []string{
	"timeout",               // Network timeout
	"connection refused",    // Temporary network issue
	"unavailable",          // Service temporarily down
	"deadlock",             // Database deadlock
	"connection closed",    // Network reconnect
	"too many open files",  // Resource exhaustion
	"temporary failure",    // Transient error
}
```

## Implementation in Worker Server

```go
import (
	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
	infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"
	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker/redpanda"
)

// Set up retry policy
server := redpanda.NewRedpandaServer(brokers, "tasks", "group", 1)

server.SetRetryPolicy(infraworker.RetryPolicy{
	MaxRetries:        5,
	InitialBackoff:    1 * time.Second,
	MaxBackoff:        1 * time.Minute,
	BackoffMultiplier: 2.0,
	JitterFraction:    0.1,
	NonRetryableErrors: []string{"validation error"},
})

// Register handler
server.RegisterHandler("send_email", func(ctx context.Context, payload sharedworker.TaskPayload) error {
	// If this returns an error matching RetryableErrors,
	// it will be requeued with exponential backoff
	// If it matches NonRetryableErrors, it goes to DLQ immediately
	// If MaxRetries exceeded, it goes to DLQ
	return emailService.Send(ctx, email)
})
```

## Monitoring Retries

Track retry metrics:

```go
import infraworker "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/worker"

metrics := infraworker.NewRetryMetrics("send_email")
metrics.TotalAttempts = 3
metrics.SuccessfulAt = 2
metrics.LastError = "timeout"

log.Println(metrics.String())
// Output: Task: send_email | SUCCESS at attempt 2 | Total Attempts: 3 | ...
```

## Best Practices

1. **Be Generous with Retries**: Most transient failures clear within seconds
2. **Exponential Backoff**: Prevents overwhelming struggling services
3. **Jitter**: Distributes retry attempts (prevents thundering herd)
4. **Classify Errors**: Know which errors are retryable
5. **Monitor DLQ**: Failed tasks in DLQ need investigation
6. **Log Correlation ID**: Track task through retries
7. **Set Reasonable MaxBackoff**: Prevent waiting too long

## Example Task Implementation

```go
import sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

server.RegisterHandler("process_payment", func(ctx context.Context, payload sharedworker.TaskPayload) error {
	userID, ok := payload["user_id"].(string)
	if !ok {
		return fmt.Errorf("validation error: missing user_id")
	}
	
	// This error will be retried (temporary)
	if err := validatePaymentMethod(userID); err != nil {
		return fmt.Errorf("payment service unavailable: %w", err)
	}
	
	// This error will NOT be retried (permanent)
	if !userHasPermission(userID) {
		return fmt.Errorf("validation error: user not allowed")
	}
	
	return processPayment(userID)
})
```
