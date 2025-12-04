package worker

import (
	"fmt"

	infraworker "go-modular-monolith/internal/infrastructure/worker"
	sharedworker "go-modular-monolith/internal/shared/worker"
)

// ModuleWorkerProvider defines the interface for modules to provide worker tasks
type ModuleWorkerProvider interface {
	// GetTaskDefinitions returns task definitions without the module importing app/infrastructure
	GetTaskDefinitions(
		userRepository interface{},
		emailService interface{},
		emailNotificationsEnabled bool,
		dataExportEnabled bool,
		reportGenerationEnabled bool,
	) []sharedworker.TaskDefinition

	// GetCronJobDefinitions returns cron job definitions
	GetCronJobDefinitions(
		emailNotificationsEnabled bool,
	) []sharedworker.CronJobDefinition
}

// TaskDefinition is an alias to the shared type
type TaskDefinition = sharedworker.TaskDefinition

// CronJobDefinition is an alias to the shared type
type CronJobDefinition = sharedworker.CronJobDefinition

// ModuleRegistry manages all module worker provider registrations
type ModuleRegistry struct {
	modules []ModuleWorkerProvider
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make([]ModuleWorkerProvider, 0),
	}
}

// Register adds a module worker provider
func (r *ModuleRegistry) Register(provider ModuleWorkerProvider) *ModuleRegistry {
	r.modules = append(r.modules, provider)
	return r
}

// RegisterAllTasks registers tasks from all modules
func (r *ModuleRegistry) RegisterAllTasks(
	registry *TaskRegistry,
	userRepository interface{},
	emailService interface{},
	emailNotificationsEnabled bool,
	dataExportEnabled bool,
	reportGenerationEnabled bool,
) error {
	for i, provider := range r.modules {
		fmt.Printf("[INFO] Getting tasks from module %d...\n", i+1)
		taskDefs := provider.GetTaskDefinitions(
			userRepository,
			emailService,
			emailNotificationsEnabled,
			dataExportEnabled,
			reportGenerationEnabled,
		)

		for _, taskDef := range taskDefs {
			fmt.Printf("[INFO] Registering task: %s\n", taskDef.TaskName)
			registry.Register(taskDef.TaskName, taskDef.Handler)
		}
	}
	return nil
}

// RegisterAllCronJobs registers cron jobs from all modules
func (r *ModuleRegistry) RegisterAllCronJobs(
	scheduler *infraworker.CronScheduler,
	emailNotificationsEnabled bool,
) error {
	for i, provider := range r.modules {
		fmt.Printf("[INFO] Getting cron jobs from module %d...\n", i+1)
		cronDefs := provider.GetCronJobDefinitions(emailNotificationsEnabled)

		for _, cronDef := range cronDefs {
			fmt.Printf("[INFO] Scheduling cron job: %s\n", cronDef.JobID)
			// Use the CronExpression from the definition, or default to Monthly(15, 9, 0)
			cronExpr := cronDef.CronExpression
			if cronExpr == (sharedworker.CronExpression{}) {
				cronExpr = sharedworker.Monthly(15, 9, 0) // Default
			}
			if err := scheduler.AddJob(
				cronDef.JobID,
				cronDef.TaskName,
				cronExpr,
				cronDef.Payload,
			); err != nil {
				return fmt.Errorf("failed to add cron job %s: %w", cronDef.JobID, err)
			}
		}
	}
	return nil
}
