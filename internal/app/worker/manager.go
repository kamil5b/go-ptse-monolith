package worker

import (
	"context"
	"fmt"

	"go-modular-monolith/internal/app/core"
	infraworker "go-modular-monolith/internal/infrastructure/worker"
	sharedworker "go-modular-monolith/internal/shared/worker"
)

// TaskRegistry holds a collection of task registrations
type TaskRegistry struct {
	registrations []TaskRegistration
}

// TaskRegistration defines how a task should be registered
type TaskRegistration struct {
	TaskName string
	Handler  sharedworker.TaskHandler
}

// NewTaskRegistry creates a new task registry
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{
		registrations: make([]TaskRegistration, 0),
	}
}

// Register adds a task registration to the registry
func (r *TaskRegistry) Register(taskName string, handler sharedworker.TaskHandler) *TaskRegistry {
	r.registrations = append(r.registrations, TaskRegistration{
		TaskName: taskName,
		Handler:  handler,
	})
	return r
}

// RegisterAll registers all tasks in the registry with the worker server
func (r *TaskRegistry) RegisterAll(server sharedworker.Server) error {
	for _, reg := range r.registrations {
		if err := server.RegisterHandler(reg.TaskName, reg.Handler); err != nil {
			return fmt.Errorf("failed to register handler for task %s: %w", reg.TaskName, err)
		}
		fmt.Printf("[INFO] Registered handler: %s\n", reg.TaskName)
	}
	return nil
}

// WorkerManager handles worker initialization and task registration
type WorkerManager struct {
	container     *core.Container
	registry      *TaskRegistry
	cronScheduler *infraworker.CronScheduler
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(container *core.Container) *WorkerManager {
	return &WorkerManager{
		container:     container,
		registry:      NewTaskRegistry(),
		cronScheduler: infraworker.NewCronScheduler(container.WorkerClient),
	}
}

// GetRegistry returns the task registry for adding registrations
func (m *WorkerManager) GetRegistry() *TaskRegistry {
	return m.registry
}

// GetCronScheduler returns the cron scheduler for adding scheduled jobs
func (m *WorkerManager) GetCronScheduler() *infraworker.CronScheduler {
	return m.cronScheduler
}

// RegisterTasks registers all tasks from the registry with the worker server
func (m *WorkerManager) RegisterTasks() error {
	return m.registry.RegisterAll(m.container.WorkerServer)
}

// Start initializes and starts the worker server
func (m *WorkerManager) Start(ctx context.Context) error {
	fmt.Println("[INFO] Starting worker server...")
	fmt.Printf("[INFO] Worker server running (backend: %s)\n", "configured")

	// Start cron scheduler in background
	go func() {
		if err := m.cronScheduler.Start(context.Background()); err != nil {
			fmt.Printf("[ERROR] Cron scheduler error: %v\n", err)
		}
	}()

	return m.container.WorkerServer.Start(ctx)
}

// Stop gracefully stops the worker server
func (m *WorkerManager) Stop(ctx context.Context) error {
	fmt.Println("[INFO] Stopping worker server...")
	// Stop cron scheduler
	if err := m.cronScheduler.Stop(); err != nil {
		fmt.Printf("[WARN] Error stopping cron scheduler: %v\n", err)
	}
	return m.container.WorkerServer.Stop(ctx)
}
