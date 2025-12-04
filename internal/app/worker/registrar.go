package worker

import (
	infraworker "go-modular-monolith/internal/infrastructure/worker"
	logger "go-modular-monolith/internal/logger"
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
		logger.WithFields(map[string]interface{}{
			"module_index": i + 1,
			"total":        len(r.modules),
		}).Info("Getting tasks from module")
		taskDefs := provider.GetTaskDefinitions(
			userRepository,
			emailService,
			emailNotificationsEnabled,
			dataExportEnabled,
			reportGenerationEnabled,
		)

		for _, taskDef := range taskDefs {
			logger.WithField("task_name", taskDef.TaskName).Info("Registering task")
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
		logger.WithFields(map[string]interface{}{
			"module_index": i + 1,
			"total":        len(r.modules),
		}).Info("Getting cron jobs from module")
		cronDefs := provider.GetCronJobDefinitions(emailNotificationsEnabled)

		for _, cronDef := range cronDefs {
			logger.WithField("cron_job_id", cronDef.JobID).Info("Scheduling cron job")
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
				return err
			}
		}
	}
	return nil
}
