package worker

import (
	userdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
	sharedemail "github.com/kamil5b/go-ptse-monolith/internal/shared/email"
	sharedworker "github.com/kamil5b/go-ptse-monolith/internal/shared/worker"
)

// TaskDefinition defines a task that the module provides
type TaskDefinition = sharedworker.TaskDefinition

// CronJobDefinition defines a cron job that the module provides
type CronJobDefinition = sharedworker.CronJobDefinition

// UserModuleWorkerTasks provides task and cron job definitions for the user module
type UserModuleWorkerTasks struct{}

// NewUserModuleWorkerTasks creates a new user module worker tasks provider
func NewUserModuleWorkerTasks() *UserModuleWorkerTasks {
	return &UserModuleWorkerTasks{}
}

// GetTaskDefinitions returns all task definitions for the user module
func (u *UserModuleWorkerTasks) GetTaskDefinitions(
	userRepository interface{},
	emailService interface{},
	emailNotificationsEnabled bool,
	dataExportEnabled bool,
	reportGenerationEnabled bool,
) []sharedworker.TaskDefinition {
	// Cast to actual types (safe because bootstrap.worker.go passes correct types)
	userRepo := userRepository.(userdomain.Repository)
	emailSvc := emailService.(sharedemail.EmailService)

	userHandler := NewUserWorkerHandler(
		userRepo,
		emailSvc,
	)

	tasks := []sharedworker.TaskDefinition{}

	// Add email notification tasks
	if emailNotificationsEnabled {
		tasks = append(tasks,
			sharedworker.TaskDefinition{
				TaskName: TaskSendWelcomeEmail,
				Handler:  userHandler.HandleSendWelcomeEmail,
			},
			sharedworker.TaskDefinition{
				TaskName: TaskSendPasswordResetEmail,
				Handler:  userHandler.HandleSendPasswordResetEmail,
			},
			sharedworker.TaskDefinition{
				TaskName: TaskSendMonthlyEmail,
				Handler:  userHandler.HandleSendMonthlyEmail,
			},
		)
	}

	// Add data export task
	if dataExportEnabled {
		tasks = append(tasks,
			sharedworker.TaskDefinition{
				TaskName: TaskExportUserData,
				Handler:  userHandler.HandleExportUserData,
			},
		)
	}

	// Add report generation task
	if reportGenerationEnabled {
		tasks = append(tasks,
			sharedworker.TaskDefinition{
				TaskName: TaskGenerateUserReport,
				Handler:  userHandler.HandleGenerateUserReport,
			},
		)
	}

	return tasks
}

// GetCronJobDefinitions returns all cron job definitions for the user module
func (u *UserModuleWorkerTasks) GetCronJobDefinitions(
	emailNotificationsEnabled bool,
) []sharedworker.CronJobDefinition {
	jobs := []sharedworker.CronJobDefinition{}

	if emailNotificationsEnabled {
		jobs = append(jobs,
			sharedworker.CronJobDefinition{
				JobID:    "monthly_email_to_all_users",
				TaskName: TaskSendMonthlyEmail,
				// CronExpression will be set by app layer: infraworker.Monthly(15, 9, 0)
				Payload: map[string]interface{}{
					"message": "Today is the day",
				},
			},
		)
	}

	return jobs
}
