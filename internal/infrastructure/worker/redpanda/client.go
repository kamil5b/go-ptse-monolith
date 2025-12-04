package redpanda

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

	"github.com/segmentio/kafka-go"
)

// RedpandaClient is a Redpanda/Kafka-based implementation of the worker.Client interface
type RedpandaClient struct {
	writer *kafka.Writer
	topic  string
}

// NewRedpandaClient creates a new Redpanda client
func NewRedpandaClient(brokers []string, topic string) *RedpandaClient {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &RedpandaClient{
		writer: writer,
		topic:  topic,
	}
}

// Enqueue enqueues a task immediately
func (c *RedpandaClient) Enqueue(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	options ...sharedworker.Option,
) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	return c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(taskName),
		Value: data,
	})
}

// EnqueueDelayed enqueues a task with a delay
// Uses a separate delayed topic + scheduler mechanism to handle delayed delivery
// The scheduler monitors the delayed topic and promotes tasks to the main topic when ready
func (c *RedpandaClient) EnqueueDelayed(
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

	// Enqueue to delayed task topic with scheduling metadata
	delayedTopic := c.topic + ".delayed"
	delayedWriter := &kafka.Writer{
		Addr:     c.writer.Addr,
		Topic:    delayedTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer delayedWriter.Close()

	scheduledTime := time.Now().Add(delay).Unix()
	return delayedWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(taskName),
		Value: data,
		Headers: []kafka.Header{
			{
				Key:   "scheduled_at",
				Value: []byte(fmt.Sprintf("%d", scheduledTime)),
			},
			{
				Key:   "original_task",
				Value: []byte(taskName),
			},
			{
				Key:   "enqueued_at",
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
		},
	})
}

// Close closes the Redpanda client
func (c *RedpandaClient) Close() error {
	return c.writer.Close()
}
