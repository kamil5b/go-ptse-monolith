package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQServer is a RabbitMQ-based implementation of the sharedworker.Server interface
type RabbitMQServer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
	handlers map[string]sharedworker.TaskHandler
	done     chan struct{}
}

// NewRabbitMQServer creates a new RabbitMQ server
func NewRabbitMQServer(url, exchange, queue string, prefetchCount int) (*RabbitMQServer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
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

	return &RabbitMQServer{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
		handlers: make(map[string]sharedworker.TaskHandler),
		done:     make(chan struct{}),
	}, nil
}

// RegisterHandler registers a handler for a task type
func (s *RabbitMQServer) RegisterHandler(taskName string, handler sharedworker.TaskHandler) error {
	s.handlers[taskName] = handler
	// Bind queue to exchange with routing key
	if err := s.channel.QueueBind(s.queue, taskName, s.exchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	return nil
}

// Start starts the RabbitMQ worker server
func (s *RabbitMQServer) Start(ctx context.Context) error {
	msgs, err := s.channel.Consume(s.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			// Get handler for this task
			handler, ok := s.handlers[msg.RoutingKey]
			if !ok {
				// No handler registered for this task type, nack and requeue
				msg.Nack(false, true)
				continue
			}

			// Parse payload
			var payload sharedworker.TaskPayload
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				// Payload is invalid, nack and don't requeue
				msg.Nack(false, false)
				continue
			}

			// Process the task
			if err := handler(ctx, payload); err != nil {
				// Task failed, nack and requeue
				msg.Nack(false, true)
				continue
			}

			// Task succeeded, ack
			msg.Ack(false)
		}
	}
}

// Stop gracefully stops the RabbitMQ worker server
func (s *RabbitMQServer) Stop(ctx context.Context) error {
	close(s.done)
	if err := s.channel.Close(); err != nil {
		return err
	}
	return s.conn.Close()
}
