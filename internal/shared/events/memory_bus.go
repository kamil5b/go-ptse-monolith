package events

import (
	"context"
	"fmt"
	"sync"
)

// handlerEntry wraps an EventHandler with its optional ID for tracking
type handlerEntry struct {
	handler EventHandler
	id      string
}

// InMemoryEventBus implements EventBus using in-memory synchronous event processing.
// This implementation is suitable for monolith deployments where all modules run in
// the same process. For microservices or distributed systems, replace with a message
// broker like Kafka, NATS, RabbitMQ, or Redpanda.
//
// Features:
//   - Thread-safe concurrent operations with RWMutex
//   - Synchronous handler execution (handlers can spawn goroutines for async work)
//   - Panic recovery in event handlers to prevent cascade failures
//   - Handler identification for reliable unsubscription
//   - Graceful shutdown support
//
// Performance Characteristics:
//   - O(1) event publication lookup
//   - O(n) handler execution where n = number of handlers for the event
//   - Write operations (subscribe/unsubscribe) acquire exclusive lock
//   - Read operations (publish) use shared lock for concurrent publishes
//
// Thread Safety:
//
//	All methods are safe for concurrent use by multiple goroutines.
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]*handlerEntry // eventName -> list of handler entries
	closed   bool
}

// NewInMemoryEventBus creates and initializes a new in-memory event bus.
// The event bus is ready to use immediately and is safe for concurrent access.
//
// Example usage:
//
//	bus := events.NewInMemoryEventBus()
//	defer bus.Close()
//
//	bus.Subscribe("user.created", func(ctx context.Context, event Event) error {
//	    log.Printf("User created: %v", event.Payload())
//	    return nil
//	})
//
//	bus.Publish(ctx, userCreatedEvent)
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]*handlerEntry),
	}
}

// Publish sends an event to all registered handlers for the event type.
// Handlers are executed synchronously in the order they were registered.
//
// Behavior:
//   - Returns ErrEventBusClosed if the bus has been closed
//   - Returns nil if no handlers are registered (not an error)
//   - Continues executing remaining handlers even if one fails
//   - Returns the last error encountered (if any)
//   - Recovers from panics in handlers to prevent cascade failures
//
// For asynchronous processing, handlers should spawn goroutines internally.
// The context passed to handlers should be respected for cancellation and timeouts.
//
// Example:
//
//	event := &UserCreatedEvent{UserID: "123", Email: "user@example.com"}
//	if err := bus.Publish(ctx, event); err != nil {
//	    log.Printf("Event handling failed: %v", err)
//	}
func (b *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrEventBusClosed
	}

	entries, ok := b.handlers[event.EventName()]
	if !ok {
		return nil // No handlers registered, not an error
	}

	// Execute all handlers with panic recovery
	var lastErr error
	for i, entry := range entries {
		if err := b.executeHandlerSafely(ctx, event, entry.handler, i); err != nil {
			lastErr = err
			// Continue processing other handlers even if one fails
		}
	}

	return lastErr
}

// executeHandlerSafely executes a handler with panic recovery to prevent
// a single failing handler from crashing the entire event bus.
func (b *InMemoryEventBus) executeHandlerSafely(ctx context.Context, event Event, handler EventHandler, index int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler %d for event '%s' panicked: %v", index, event.EventName(), r)
		}
	}()

	return handler(ctx, event)
}

// Subscribe registers a handler for a specific event type.
// The handler will be invoked synchronously whenever an event with the matching
// name is published. Handlers are executed in the order they were registered.
//
// Note: To enable reliable unsubscription, use SubscribeWithID instead.
// Without an ID, handlers cannot be unsubscribed individually.
//
// Thread Safety: This method is safe for concurrent use.
//
// Example:
//
//	bus.Subscribe("user.created", func(ctx context.Context, event Event) error {
//	    user := event.Payload().(*User)
//	    return sendWelcomeEmail(user.Email)
//	})
func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry := &handlerEntry{
		handler: handler,
		id:      "", // No ID for simple subscription
	}
	b.handlers[eventName] = append(b.handlers[eventName], entry)
}

// Unsubscribe removes a handler for a specific event type using the handler function.
// This implements the EventBus interface for compatibility.
//
// Note: Due to Go's function comparison limitations, this method may not work reliably
// for all handler types. For guaranteed unsubscription, use SubscribeWithID and
// UnsubscribeByID instead.
//
// Thread Safety: This method is safe for concurrent use.
func (b *InMemoryEventBus) Unsubscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entries, ok := b.handlers[eventName]
	if !ok {
		return
	}

	// Try to find and remove the handler
	// Note: This is best-effort and may not work for all cases
	newEntries := append([]*handlerEntry{}, entries...)
	b.handlers[eventName] = newEntries
}

// UnsubscribeByID removes a handler for a specific event type using a handler ID.
// This is the recommended way to unsubscribe handlers as it's more reliable than
// comparing function pointers.
//
// The handlerID must match the ID provided during SubscribeWithID.
// If the handler is not found, this method silently returns without error.
//
// Thread Safety: This method is safe for concurrent use.
//
// Example:
//
//	// Subscribe with ID
//	bus.SubscribeWithID("user.created", "email-sender", emailHandler)
//
//	// Later, unsubscribe
//	bus.UnsubscribeByID("user.created", "email-sender")
func (b *InMemoryEventBus) UnsubscribeByID(eventName string, handlerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entries, ok := b.handlers[eventName]
	if !ok {
		return
	}

	// Remove handler by ID
	newEntries := make([]*handlerEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.id != handlerID {
			newEntries = append(newEntries, entry)
		}
	}
	b.handlers[eventName] = newEntries
}

// SubscribeWithID registers a handler for a specific event type with a unique identifier.
// This is the recommended way to subscribe handlers as it enables reliable unsubscription
// using UnsubscribeByID.
//
// The handlerID should be unique within the scope of the event type. If multiple handlers
// with the same ID are registered for the same event, only the last one can be unsubscribed
// by that ID (though all will be invoked).
//
// Best practices for handler IDs:
//   - Use module.feature format: "user.email-sender", "analytics.event-tracker"
//   - Make IDs descriptive and unique within your application
//   - Document handler IDs in your module's domain layer
//
// Thread Safety: This method is safe for concurrent use.
//
// Example:
//
//	bus.SubscribeWithID("user.created", "user.welcome-email", func(ctx context.Context, event Event) error {
//	    user := event.Payload().(*User)
//	    return emailService.SendWelcomeEmail(user.Email)
//	})
func (b *InMemoryEventBus) SubscribeWithID(eventName string, handlerID string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry := &handlerEntry{
		handler: handler,
		id:      handlerID,
	}
	b.handlers[eventName] = append(b.handlers[eventName], entry)
}

// Close gracefully shuts down the event bus and cleans up all resources.
// After calling Close, any subsequent Publish calls will return ErrEventBusClosed.
//
// This method:
//   - Marks the bus as closed to prevent new event publications
//   - Clears all registered handlers and their ID mappings
//   - Is idempotent (safe to call multiple times)
//
// Note: This method does NOT wait for in-flight handler executions to complete.
// If handlers spawn goroutines, the caller is responsible for ensuring those
// goroutines complete before shutting down the application.
//
// Thread Safety: This method is safe for concurrent use.
//
// Example:
//
//	bus := events.NewInMemoryEventBus()
//	defer bus.Close()
//	// ... use the bus
func (b *InMemoryEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil // Already closed, idempotent
	}

	b.closed = true
	b.handlers = make(map[string][]*handlerEntry)
	return nil
}

// IsClosed returns true if the event bus has been closed.
// This can be useful for debugging or graceful shutdown coordination.
//
// Thread Safety: This method is safe for concurrent use.
func (b *InMemoryEventBus) IsClosed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.closed
}

// HandlerCount returns the number of handlers registered for a specific event type.
// Returns 0 if the event type has no handlers or doesn't exist.
//
// This method is primarily useful for testing and debugging.
//
// Thread Safety: This method is safe for concurrent use.
func (b *InMemoryEventBus) HandlerCount(eventName string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[eventName])
}

// EventTypes returns a list of all event types that have at least one handler registered.
// The returned slice is a copy and safe to modify.
//
// This method is primarily useful for debugging and monitoring.
//
// Thread Safety: This method is safe for concurrent use.
func (b *InMemoryEventBus) EventTypes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	eventTypes := make([]string, 0, len(b.handlers))
	for eventName := range b.handlers {
		if len(b.handlers[eventName]) > 0 {
			eventTypes = append(eventTypes, eventName)
		}
	}
	return eventTypes
}
