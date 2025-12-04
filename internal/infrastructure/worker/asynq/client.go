package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharedworker "go-modular-monolith/internal/shared/worker"

	"github.com/hibiken/asynq"
)

// AsynqClient is an Asynq-based implementation of the sharedworker.Client interface
type AsynqClient struct {
	client *asynq.Client
}

// NewAsynqClient creates a new Asynq client
func NewAsynqClient(redisURL string) *AsynqClient {
	return &AsynqClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisURL}),
	}
}

// Enqueue enqueues a task immediately
func (c *AsynqClient) Enqueue(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	options ...sharedworker.Option,
) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build Asynq options
	asynqOptions := []asynq.Option{}
	for _, opt := range options {
		switch o := opt.(type) {
		case *sharedworker.PriorityOption:
			// Asynq doesn't have Priority option, use Queue instead
			asynqOptions = append(asynqOptions, asynq.Queue(fmt.Sprintf("queue_%d", o.Priority)))
		case *sharedworker.MaxRetriesOption:
			asynqOptions = append(asynqOptions, asynq.MaxRetry(o.MaxRetries))
		case *sharedworker.TimeoutOption:
			asynqOptions = append(asynqOptions, asynq.Timeout(o.Timeout))
		case *sharedworker.QueueOption:
			asynqOptions = append(asynqOptions, asynq.Queue(o.Queue))
		}
	}

	task := asynq.NewTask(taskName, data)
	_, err = c.client.EnqueueContext(ctx, task, asynqOptions...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueDelayed enqueues a task with a delay
func (c *AsynqClient) EnqueueDelayed(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	delay time.Duration,
	options ...sharedworker.Option,
) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build Asynq options with delay
	asynqOptions := []asynq.Option{asynq.ProcessIn(delay)}
	for _, opt := range options {
		switch o := opt.(type) {
		case *sharedworker.PriorityOption:
			// Asynq doesn't have Priority option, use Queue instead
			asynqOptions = append(asynqOptions, asynq.Queue(fmt.Sprintf("queue_%d", o.Priority)))
		case *sharedworker.MaxRetriesOption:
			asynqOptions = append(asynqOptions, asynq.MaxRetry(o.MaxRetries))
		case *sharedworker.TimeoutOption:
			asynqOptions = append(asynqOptions, asynq.Timeout(o.Timeout))
		case *sharedworker.QueueOption:
			asynqOptions = append(asynqOptions, asynq.Queue(o.Queue))
		}
	}

	task := asynq.NewTask(taskName, data)
	_, err = c.client.EnqueueContext(ctx, task, asynqOptions...)
	if err != nil {
		return fmt.Errorf("failed to enqueue delayed task: %w", err)
	}

	return nil
}

// Close closes the Asynq client
func (c *AsynqClient) Close() error {
	return c.client.Close()
}
