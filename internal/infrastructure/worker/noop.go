package worker

import (
	"context"
	"time"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

// NoOpClient is a no-op implementation of the Client interface
// Used when workers are disabled via feature flags
type NoOpClient struct{}

// NewNoOpClient creates a new no-op client
func NewNoOpClient() *NoOpClient {
	return &NoOpClient{}
}

// Enqueue is a no-op
func (c *NoOpClient) Enqueue(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	options ...sharedworker.Option,
) error {
	return nil
}

// EnqueueDelayed is a no-op
func (c *NoOpClient) EnqueueDelayed(
	ctx context.Context,
	taskName string,
	payload sharedworker.TaskPayload,
	delay time.Duration,
	options ...sharedworker.Option,
) error {
	return nil
}

// Close is a no-op
func (c *NoOpClient) Close() error {
	return nil
}

// NoOpServer is a no-op implementation of the Server interface
// Used when workers are disabled via feature flags
type NoOpServer struct{}

// NewNoOpServer creates a new no-op server
func NewNoOpServer() *NoOpServer {
	return &NoOpServer{}
}

// RegisterHandler is a no-op
func (s *NoOpServer) RegisterHandler(taskName string, handler sharedworker.TaskHandler) error {
	return nil
}

// Start is a no-op
func (s *NoOpServer) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Stop is a no-op
func (s *NoOpServer) Stop(ctx context.Context) error {
	return nil
}
