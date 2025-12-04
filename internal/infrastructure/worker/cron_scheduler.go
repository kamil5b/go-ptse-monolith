package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

// CronExpression is an alias to the shared type
type CronExpression = sharedworker.CronExpression

// Re-export cron expression helpers for convenience
var (
	EveryMinute = sharedworker.EveryMinute
	EveryHour   = sharedworker.EveryHour
	Daily       = sharedworker.Daily
	Weekly      = sharedworker.Weekly
	Monthly     = sharedworker.Monthly
)

// CronScheduler schedules recurring tasks based on cron expressions
type CronScheduler struct {
	jobs      map[string]*CronJob
	jobsMutex sync.RWMutex
	done      chan struct{}
	client    sharedworker.Client
	ticker    *time.Ticker
}

// CronJob represents a scheduled recurring task
type CronJob struct {
	ID            string
	TaskName      string
	Schedule      CronExpression
	Payload       sharedworker.TaskPayload
	LastRun       time.Time
	NextRun       time.Time
	Enabled       bool
	MaxConcurrent int // Maximum concurrent executions
	running       int // Current running count
	runMutex      sync.Mutex
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(client sharedworker.Client) *CronScheduler {
	return &CronScheduler{
		jobs:   make(map[string]*CronJob),
		done:   make(chan struct{}),
		client: client,
		ticker: time.NewTicker(1 * time.Minute), // Check every minute
	}
}

// AddJob adds a new cron job
func (cs *CronScheduler) AddJob(id, taskName string, schedule CronExpression, payload sharedworker.TaskPayload) error {
	cs.jobsMutex.Lock()
	defer cs.jobsMutex.Unlock()

	if _, exists := cs.jobs[id]; exists {
		return fmt.Errorf("job with ID %s already exists", id)
	}

	job := &CronJob{
		ID:            id,
		TaskName:      taskName,
		Schedule:      schedule,
		Payload:       payload,
		Enabled:       true,
		MaxConcurrent: 1,
		NextRun:       cs.calculateNextRun(schedule, time.Now()),
	}

	cs.jobs[id] = job
	log.Printf("Added cron job: %s | Next run: %v\n", id, job.NextRun)
	return nil
}

// RemoveJob removes a cron job
func (cs *CronScheduler) RemoveJob(id string) error {
	cs.jobsMutex.Lock()
	defer cs.jobsMutex.Unlock()

	if _, exists := cs.jobs[id]; !exists {
		return fmt.Errorf("job with ID %s not found", id)
	}

	delete(cs.jobs, id)
	log.Printf("Removed cron job: %s\n", id)
	return nil
}

// EnableJob enables a job
func (cs *CronScheduler) EnableJob(id string) error {
	cs.jobsMutex.Lock()
	defer cs.jobsMutex.Unlock()

	job, exists := cs.jobs[id]
	if !exists {
		return fmt.Errorf("job with ID %s not found", id)
	}

	job.Enabled = true
	return nil
}

// DisableJob disables a job
func (cs *CronScheduler) DisableJob(id string) error {
	cs.jobsMutex.Lock()
	defer cs.jobsMutex.Unlock()

	job, exists := cs.jobs[id]
	if !exists {
		return fmt.Errorf("job with ID %s not found", id)
	}

	job.Enabled = false
	return nil
}

// Start starts the cron scheduler
func (cs *CronScheduler) Start(ctx context.Context) error {
	log.Println("Starting cron scheduler...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cs.done:
			return nil
		case <-cs.ticker.C:
			cs.checkAndRunJobs(ctx)
		}
	}
}

// Stop stops the scheduler
func (cs *CronScheduler) Stop() error {
	close(cs.done)
	cs.ticker.Stop()
	return nil
}

// checkAndRunJobs checks if any jobs should run
func (cs *CronScheduler) checkAndRunJobs(ctx context.Context) {
	cs.jobsMutex.RLock()
	jobsCopy := make([]*CronJob, 0, len(cs.jobs))
	for _, job := range cs.jobs {
		if job.Enabled && time.Now().After(job.NextRun) {
			jobsCopy = append(jobsCopy, job)
		}
	}
	cs.jobsMutex.RUnlock()

	// Execute jobs concurrently
	for _, job := range jobsCopy {
		go cs.executeJob(ctx, job)
	}
}

// executeJob executes a cron job
func (cs *CronScheduler) executeJob(ctx context.Context, job *CronJob) {
	job.runMutex.Lock()
	if job.running >= job.MaxConcurrent {
		job.runMutex.Unlock()
		return // Skip if already running max concurrent
	}
	job.running++
	job.runMutex.Unlock()

	defer func() {
		job.runMutex.Lock()
		job.running--
		job.runMutex.Unlock()
	}()

	// Enqueue the task
	if err := cs.client.Enqueue(ctx, job.TaskName, job.Payload); err != nil {
		log.Printf("Failed to enqueue cron job %s: %v\n", job.ID, err)
		return
	}

	job.LastRun = time.Now()
	job.NextRun = cs.calculateNextRun(job.Schedule, time.Now())

	log.Printf("Executed cron job: %s | Last run: %v | Next run: %v\n",
		job.ID, job.LastRun, job.NextRun)
}

// calculateNextRun calculates the next execution time for a job
func (cs *CronScheduler) calculateNextRun(expr CronExpression, from time.Time) time.Time {
	next := from

	// If all fields are -1 (every minute), just add a minute
	if expr.Minute == -1 && expr.Hour == -1 && expr.Day == -1 &&
		expr.Month == -1 && expr.Weekday == -1 {
		return next.Add(1 * time.Minute).Truncate(time.Minute)
	}

	// Start from next minute boundary
	next = next.Add(1 * time.Minute).Truncate(time.Minute)

	for attempts := 0; attempts < 366*24*60; attempts++ {
		if cs.matchesCronExpression(expr, next) {
			return next
		}
		next = next.Add(1 * time.Minute)
	}

	// Fallback - schedule in 1 hour if no match found
	return from.Add(1 * time.Hour)
}

// matchesCronExpression checks if a time matches the cron expression
func (cs *CronScheduler) matchesCronExpression(expr CronExpression, t time.Time) bool {
	// Check minute
	if expr.Minute >= 0 && expr.Minute != t.Minute() {
		return false
	}

	// Check hour
	if expr.Hour >= 0 && expr.Hour != t.Hour() {
		return false
	}

	// Check day of month
	if expr.Day > 0 && expr.Day != t.Day() {
		return false
	}

	// Check month
	if expr.Month > 0 && expr.Month != int(t.Month()) {
		return false
	}

	// Check weekday (0 = Sunday)
	if expr.Weekday >= 0 && expr.Weekday != int(t.Weekday()) {
		return false
	}

	return true
}

// ListJobs returns all scheduled jobs as pointers to avoid copying structs with sync.Mutex
func (cs *CronScheduler) ListJobs() []*CronJob {
	cs.jobsMutex.RLock()
	defer cs.jobsMutex.RUnlock()

	jobs := make([]*CronJob, 0, len(cs.jobs))
	for _, job := range cs.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}
