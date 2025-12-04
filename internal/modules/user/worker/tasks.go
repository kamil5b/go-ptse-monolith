package worker

const (
	// TaskSendWelcomeEmail is the task name for sending welcome emails
	TaskSendWelcomeEmail = "user:send_welcome_email"

	// TaskSendPasswordResetEmail is the task name for sending password reset emails
	TaskSendPasswordResetEmail = "user:send_password_reset_email"

	// TaskExportUserData is the task name for exporting user data
	TaskExportUserData = "user:export_user_data"

	// TaskGenerateUserReport is the task name for generating user reports
	TaskGenerateUserReport = "user:generate_user_report"

	// TaskSendMonthlyEmail is the task name for sending monthly emails to all users
	TaskSendMonthlyEmail = "user:send_monthly_email"
)

// SendWelcomeEmailPayload is the payload for the welcome email task
type SendWelcomeEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// SendPasswordResetEmailPayload is the payload for the password reset email task
type SendPasswordResetEmailPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	ResetLink string `json:"reset_link"`
}

// ExportUserDataPayload is the payload for the user data export task
type ExportUserDataPayload struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Format      string `json:"format"` // json, csv, xml
	Destination string `json:"destination"`
}

// GenerateUserReportPayload is the payload for the user report generation task
type GenerateUserReportPayload struct {
	UserID      string `json:"user_id"`
	ReportType  string `json:"report_type"` // login_history, activity, engagement
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Destination string `json:"destination"`
}

// SendMonthlyEmailPayload is the payload for the monthly email task
type SendMonthlyEmailPayload struct {
	Message string `json:"message"` // Message to send to all users
}
