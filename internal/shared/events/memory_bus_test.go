package events

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockEvent implements the Event interface for testing
type mockEvent struct {
	name    string
	payload interface{}
}

func (e *mockEvent) EventName() string {
	return e.name
}

func (e *mockEvent) Payload() interface{} {
	return e.payload
}

func TestNewInMemoryEventBus(t *testing.T) {
	bus := NewInMemoryEventBus()

	if bus == nil {
		t.Fatal("NewInMemoryEventBus returned nil")
	}

	if bus.handlers == nil {
		t.Error("handlers map not initialized")
	}

	if bus.closed {
		t.Error("bus should not be closed on creation")
	}
}

func TestPublish_NoHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	event := &mockEvent{name: "test.event", payload: "data"}
	err := bus.Publish(context.Background(), event)

	if err != nil {
		t.Errorf("Publish with no handlers should not return error, got: %v", err)
	}
}

func TestPublish_WithHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	var callCount int32
	var receivedPayload string

	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		receivedPayload = event.Payload().(string)
		return nil
	}

	bus.Subscribe("test.event", handler)

	event := &mockEvent{name: "test.event", payload: "test-data"}
	err := bus.Publish(context.Background(), event)

	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected handler to be called once, got %d", callCount)
	}

	if receivedPayload != "test-data" {
		t.Errorf("Expected payload 'test-data', got '%s'", receivedPayload)
	}
}

func TestPublish_MultipleHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	var callCount int32

	handler1 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	handler2 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	handler3 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)
	bus.Subscribe("test.event", handler3)

	event := &mockEvent{name: "test.event", payload: "data"}
	err := bus.Publish(context.Background(), event)

	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("Expected 3 handlers to be called, got %d", callCount)
	}
}

func TestPublish_HandlerError(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	expectedErr := errors.New("handler error")
	var callCount int32

	handler1 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return expectedErr
	}

	handler2 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)

	event := &mockEvent{name: "test.event", payload: "data"}
	err := bus.Publish(context.Background(), event)

	// Should return the last error but continue processing
	if err == nil {
		t.Error("Expected error from handler, got nil")
	}

	// Both handlers should have been called
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected both handlers to be called, got %d", callCount)
	}
}

func TestPublish_HandlerPanic(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	var callCount int32

	handler1 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		panic("test panic")
	}

	handler2 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)

	event := &mockEvent{name: "test.event", payload: "data"}
	err := bus.Publish(context.Background(), event)

	// Should recover from panic and return error
	if err == nil {
		t.Error("Expected error from panicking handler, got nil")
	}

	// Second handler should still be called
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected both handlers to be called, got %d", callCount)
	}
}

func TestPublish_ClosedBus(t *testing.T) {
	bus := NewInMemoryEventBus()
	bus.Close()

	event := &mockEvent{name: "test.event", payload: "data"}
	err := bus.Publish(context.Background(), event)

	if err != ErrEventBusClosed {
		t.Errorf("Expected ErrEventBusClosed, got: %v", err)
	}
}

func TestSubscribeWithID_UnsubscribeByID(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	var callCount int32

	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	bus.SubscribeWithID("test.event", "handler-1", handler)

	// Handler should be called
	event := &mockEvent{name: "test.event", payload: "data"}
	bus.Publish(context.Background(), event)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected handler to be called once, got %d", callCount)
	}

	// Unsubscribe by ID
	bus.UnsubscribeByID("test.event", "handler-1")

	// Handler should not be called after unsubscribe
	bus.Publish(context.Background(), event)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected handler not to be called after unsubscribe, got %d calls", callCount)
	}
}

func TestSubscribeWithID_MultipleHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	var handler1Calls, handler2Calls int32

	handler1 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&handler1Calls, 1)
		return nil
	}

	handler2 := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&handler2Calls, 1)
		return nil
	}

	bus.SubscribeWithID("test.event", "handler-1", handler1)
	bus.SubscribeWithID("test.event", "handler-2", handler2)

	event := &mockEvent{name: "test.event", payload: "data"}
	bus.Publish(context.Background(), event)

	if atomic.LoadInt32(&handler1Calls) != 1 {
		t.Errorf("Expected handler1 to be called once, got %d", handler1Calls)
	}

	if atomic.LoadInt32(&handler2Calls) != 1 {
		t.Errorf("Expected handler2 to be called once, got %d", handler2Calls)
	}

	// Unsubscribe only handler1
	bus.UnsubscribeByID("test.event", "handler-1")

	bus.Publish(context.Background(), event)

	// handler1 should not be called again
	if atomic.LoadInt32(&handler1Calls) != 1 {
		t.Errorf("Expected handler1 not to be called after unsubscribe, got %d calls", handler1Calls)
	}

	// handler2 should still be called
	if atomic.LoadInt32(&handler2Calls) != 2 {
		t.Errorf("Expected handler2 to be called twice, got %d", handler2Calls)
	}
}

func TestUnsubscribeByID_NonExistent(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	// Should not panic or error when unsubscribing non-existent handler
	bus.UnsubscribeByID("test.event", "non-existent")
	bus.UnsubscribeByID("non-existent-event", "handler-1")
}

func TestClose_Idempotent(t *testing.T) {
	bus := NewInMemoryEventBus()

	err1 := bus.Close()
	if err1 != nil {
		t.Errorf("First Close returned error: %v", err1)
	}

	err2 := bus.Close()
	if err2 != nil {
		t.Errorf("Second Close returned error: %v", err2)
	}

	if !bus.IsClosed() {
		t.Error("Bus should be closed")
	}
}

func TestIsClosed(t *testing.T) {
	bus := NewInMemoryEventBus()

	if bus.IsClosed() {
		t.Error("New bus should not be closed")
	}

	bus.Close()

	if !bus.IsClosed() {
		t.Error("Bus should be closed after Close()")
	}
}

func TestHandlerCount(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	if bus.HandlerCount("test.event") != 0 {
		t.Error("New bus should have 0 handlers")
	}

	handler := func(ctx context.Context, event Event) error { return nil }

	bus.Subscribe("test.event", handler)
	if bus.HandlerCount("test.event") != 1 {
		t.Errorf("Expected 1 handler, got %d", bus.HandlerCount("test.event"))
	}

	bus.Subscribe("test.event", handler)
	if bus.HandlerCount("test.event") != 2 {
		t.Errorf("Expected 2 handlers, got %d", bus.HandlerCount("test.event"))
	}

	bus.Subscribe("other.event", handler)
	if bus.HandlerCount("test.event") != 2 {
		t.Errorf("Expected 2 handlers for test.event, got %d", bus.HandlerCount("test.event"))
	}
}

func TestEventTypes(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	types := bus.EventTypes()
	if len(types) != 0 {
		t.Errorf("Expected 0 event types, got %d", len(types))
	}

	handler := func(ctx context.Context, event Event) error { return nil }

	bus.Subscribe("event1", handler)
	bus.Subscribe("event2", handler)
	bus.Subscribe("event3", handler)

	types = bus.EventTypes()
	if len(types) != 3 {
		t.Errorf("Expected 3 event types, got %d", len(types))
	}

	// Check that all event types are present
	typeMap := make(map[string]bool)
	for _, eventType := range types {
		typeMap[eventType] = true
	}

	if !typeMap["event1"] || !typeMap["event2"] || !typeMap["event3"] {
		t.Error("Not all event types were returned")
	}
}

func TestConcurrentPublish(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	var callCount int32

	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	bus.Subscribe("test.event", handler)

	// Publish from multiple goroutines concurrently
	const numPublishers = 10
	var wg sync.WaitGroup
	wg.Add(numPublishers)

	for i := 0; i < numPublishers; i++ {
		go func() {
			defer wg.Done()
			event := &mockEvent{name: "test.event", payload: "data"}
			bus.Publish(context.Background(), event)
		}()
	}

	wg.Wait()

	if atomic.LoadInt32(&callCount) != numPublishers {
		t.Errorf("Expected %d handler calls, got %d", numPublishers, callCount)
	}
}

func TestConcurrentSubscribeUnsubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	const numOperations = 100
	var wg sync.WaitGroup
	wg.Add(numOperations * 2)

	// Concurrent subscribes
	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer wg.Done()
			bus.SubscribeWithID("test.event", "handler-"+string(rune(id)), handler)
		}(i)
	}

	// Concurrent unsubscribes
	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer wg.Done()
			bus.UnsubscribeByID("test.event", "handler-"+string(rune(id)))
		}(i)
	}

	wg.Wait()

	// Should not crash or deadlock
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	var wg sync.WaitGroup
	const numOperations = 50

	// Concurrent publishes
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()
			event := &mockEvent{name: "test.event", payload: "data"}
			bus.Publish(context.Background(), event)
		}()
	}

	// Concurrent subscribes
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer wg.Done()
			bus.SubscribeWithID("test.event", "handler-"+string(rune(id)), handler)
		}(i)
	}

	wg.Wait()

	// Should not crash or deadlock
}

func TestContextCancellation(t *testing.T) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handlerCalled := false
	ctx, cancel := context.WithCancel(context.Background())

	handler := func(ctx context.Context, event Event) error {
		handlerCalled = true
		// Handler should respect context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	bus.Subscribe("test.event", handler)

	cancel() // Cancel before publishing

	event := &mockEvent{name: "test.event", payload: "data"}
	err := bus.Publish(ctx, event)

	if !handlerCalled {
		t.Error("Handler should have been called even with cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

// Benchmark tests
func BenchmarkPublish_SingleHandler(b *testing.B) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	bus.Subscribe("test.event", handler)
	event := &mockEvent{name: "test.event", payload: "data"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
	}
}

func BenchmarkPublish_MultipleHandlers(b *testing.B) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	for i := 0; i < 10; i++ {
		bus.Subscribe("test.event", handler)
	}

	event := &mockEvent{name: "test.event", payload: "data"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
	}
}

func BenchmarkSubscribe(b *testing.B) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Subscribe("test.event", handler)
	}
}

func BenchmarkSubscribeWithID(b *testing.B) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.SubscribeWithID("test.event", "handler", handler)
	}
}
