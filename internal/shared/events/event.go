package events

import "context"

// Event represents a domain event that can be published and consumed
type Event interface {
	// EventName returns the unique name of the event (e.g., "product.created")
	EventName() string
	// Payload returns the event data
	Payload() any
}

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, event Event) error

// EventBus defines the interface for publishing and subscribing to domain events
// In monolith: uses in-memory channels
// In microservices: can be swapped with Kafka, NATS, RabbitMQ, etc.
type EventBus interface {
	// Publish sends an event to all subscribers
	Publish(ctx context.Context, event Event) error

	// Subscribe registers a handler for a specific event type
	Subscribe(eventName string, handler EventHandler)

	// Unsubscribe removes a handler for a specific event type
	Unsubscribe(eventName string, handler EventHandler)

	// Close gracefully shuts down the event bus
	Close() error
}
