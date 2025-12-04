package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharedworker "go-modular-monolith/internal/shared/worker"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient is a RabbitMQ-based implementation of the worker.Client interface
type RabbitMQClient struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url, exchange, queue string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQClient{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
	}, nil
}

// Enqueue enqueues a task immediately
func (c *RabbitMQClient) Enqueue(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	options ...sharedworker.Option,
) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	return c.channel.PublishWithContext(
		ctx,
		c.exchange,
		taskName, // routing key
		true,     // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         data,
		},
	)
}

// EnqueueDelayed enqueues a task with a delay (not natively supported by RabbitMQ)
// For delayed delivery, clients should use the RabbitMQ Delayed Message Plugin
// or implement their own scheduling mechanism
func (c *RabbitMQClient) EnqueueDelayed(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	delay time.Duration,
	options ...sharedworker.Option,
) error {
	// Use RabbitMQ Delayed Message Plugin for delayed delivery
	// Plugin: https://github.com/rabbitmq/rabbitmq-delayed-message-exchange
	// Install with: rabbitmq-plugins enable rabbitmq_delayed_message_exchange
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Set x-delay header in milliseconds
	headers := amqp.Table{
		"x-delay": int64(delay.Milliseconds()),
	}

	return c.channel.PublishWithContext(
		ctx,
		c.exchange,
		taskName,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         data,
			Headers:      headers,
		},
	)
}

// Close closes the RabbitMQ client
func (c *RabbitMQClient) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
