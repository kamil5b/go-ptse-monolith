package worker

import (
	"context"
	"time"
)

// TaskPayload represents the data passed to a task
type TaskPayload map[string]interface{}

// TaskHandler processes a task and returns an error if processing fails
type TaskHandler func(ctx context.Context, payload TaskPayload) error

// TaskDefinition defines a task that a module provides
type TaskDefinition struct {
	TaskName string
	Handler  TaskHandler
}

// CronJobDefinition defines a cron job that a module provides
type CronJobDefinition struct {
	JobID          string
	TaskName       string
	CronExpression CronExpression
	Payload        map[string]interface{}
}

// CronExpression represents a simplified cron expression
// Format: "minute hour day month weekday" (subset of standard cron)
type CronExpression struct {
	Minute  int // 0-59 or -1 for any
	Hour    int // 0-23 or -1 for any
	Day     int // 1-31 or -1 for any
	Month   int // 1-12 or -1 for any
	Weekday int // 0-6 (Sun-Sat) or -1 for any
}

// EveryMinute creates a cron expression that runs every minute
func EveryMinute() CronExpression {
	return CronExpression{Minute: -1, Hour: -1, Day: -1, Month: -1, Weekday: -1}
}

// EveryHour creates a cron expression that runs every hour at minute 0
func EveryHour() CronExpression {
	return CronExpression{Minute: 0, Hour: -1, Day: -1, Month: -1, Weekday: -1}
}

// Daily creates a cron expression that runs daily at specified time
func Daily(hour, minute int) CronExpression {
	return CronExpression{Minute: minute, Hour: hour, Day: -1, Month: -1, Weekday: -1}
}

// Weekly creates a cron expression that runs weekly on specified day and time
func Weekly(weekday, hour, minute int) CronExpression {
	return CronExpression{Minute: minute, Hour: hour, Day: -1, Month: -1, Weekday: weekday}
}

// Monthly creates a cron expression that runs monthly on specified day and time
func Monthly(day, hour, minute int) CronExpression {
	return CronExpression{Minute: minute, Hour: hour, Day: day, Month: -1, Weekday: -1}
}

// Option is used to specify task options like priority, retry count, timeout, etc.
type Option interface{}

// Client is responsible for enqueueing tasks
// Can be used by modules to enqueue tasks without importing infrastructure
type Client interface {
	// Enqueue adds a task to the queue
	Enqueue(ctx context.Context, taskName string, payload TaskPayload, options ...Option) error

	// EnqueueDelayed adds a task to the queue with a delay before processing
	EnqueueDelayed(ctx context.Context, taskName string, payload TaskPayload, delay time.Duration, options ...Option) error

	// Close closes the client connection
	Close() error
}

// Server is responsible for processing tasks from the queue
// Used by app layer to register handlers and start/stop the worker
type Server interface {
	// RegisterHandler registers a handler for a specific task type
	RegisterHandler(taskName string, handler TaskHandler) error

	// Start starts the worker server and begins processing tasks
	Start(ctx context.Context) error

	// Stop gracefully stops the worker server
	Stop(ctx context.Context) error
}

// Scheduler is responsible for scheduling recurring tasks
// Used by app layer to manage cron jobs
type Scheduler interface {
	// AddJob adds a new scheduled job
	AddJob(id, taskName string, schedule interface{}, payload TaskPayload) error

	// RemoveJob removes a scheduled job
	RemoveJob(id string) error

	// EnableJob enables a job
	EnableJob(id string) error

	// DisableJob disables a job
	DisableJob(id string) error

	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop stops the scheduler
	Stop() error
}

// PriorityOption sets the priority of a task (higher = more important)
type PriorityOption struct {
	Priority int
}

// MaxRetriesOption sets the maximum number of retries for a task
type MaxRetriesOption struct {
	MaxRetries int
}

// TimeoutOption sets the maximum time a task can run
type TimeoutOption struct {
	Timeout time.Duration
}

// QueueOption specifies which queue to use
type QueueOption struct {
	Queue string
}

// NewPriorityOption creates a new priority option
func NewPriorityOption(priority int) *PriorityOption {
	return &PriorityOption{Priority: priority}
}

// NewMaxRetriesOption creates a new max retries option
func NewMaxRetriesOption(maxRetries int) *MaxRetriesOption {
	return &MaxRetriesOption{MaxRetries: maxRetries}
}

// NewTimeoutOption creates a new timeout option
func NewTimeoutOption(timeout time.Duration) *TimeoutOption {
	return &TimeoutOption{Timeout: timeout}
}

// NewQueueOption creates a new queue option
func NewQueueOption(queue string) *QueueOption {
	return &QueueOption{Queue: queue}
}
