package events

import (
	"context"
	"sync"
)

// InMemoryEventBus implements EventBus using in-memory channels
// Suitable for monolith deployments; can be replaced with Kafka/NATS for microservices
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	closed   bool
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish sends an event to all registered handlers synchronously
// For async processing, handlers can spawn goroutines internally
func (b *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrEventBusClosed
	}

	handlers, ok := b.handlers[event.EventName()]
	if !ok {
		return nil // No handlers registered, not an error
	}

	// Execute all handlers
	var lastErr error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			lastErr = err
			// Continue processing other handlers even if one fails
		}
	}

	return lastErr
}

// Subscribe registers a handler for a specific event type
func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Unsubscribe removes a handler for a specific event type
// Note: This compares function pointers, so the same function reference must be used
func (b *InMemoryEventBus) Unsubscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers, ok := b.handlers[eventName]
	if !ok {
		return
	}

	// Find and remove the handler
	for i := range handlers {
		// Compare function pointers (this has limitations in Go)
		// In production, consider using handler IDs instead
		if &handlers[i] == &handler {
			b.handlers[eventName] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

// Close gracefully shuts down the event bus
func (b *InMemoryEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.handlers = make(map[string][]EventHandler)
	return nil
}
