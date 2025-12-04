package asynq

import (
	"context"
	"encoding/json"
	"fmt"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"

	"github.com/hibiken/asynq"
)

// AsynqServer is an Asynq-based implementation of the sharedworker.Server interface
type AsynqServer struct {
	srv      *asynq.Server
	mux      *asynq.ServeMux
	handlers map[string]sharedworker.TaskHandler
}

// NewAsynqServer creates a new Asynq server
func NewAsynqServer(redisURL string, concurrency int) *AsynqServer {
	return &AsynqServer{
		srv: asynq.NewServer(
			asynq.RedisClientOpt{Addr: redisURL},
			asynq.Config{
				Concurrency: concurrency,
				Queues: map[string]int{
					"critical": 6,
					"default":  3,
					"low":      1,
				},
			},
		),
		mux:      asynq.NewServeMux(),
		handlers: make(map[string]sharedworker.TaskHandler),
	}
}

// RegisterHandler registers a handler for a task type
func (s *AsynqServer) RegisterHandler(taskName string, handler sharedworker.TaskHandler) error {
	s.handlers[taskName] = handler
	s.mux.HandleFunc(taskName, func(ctx context.Context, t *asynq.Task) error {
		// Convert Asynq task payload to sharedworker.TaskPayload
		var payload sharedworker.TaskPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		return handler(ctx, payload)
	})
	return nil
}

// Start starts the Asynq worker server
func (s *AsynqServer) Start(ctx context.Context) error {
	return s.srv.Start(s.mux)
}

// Stop gracefully stops the Asynq worker server
func (s *AsynqServer) Stop(ctx context.Context) error {
	s.srv.Stop()
	return nil
}
