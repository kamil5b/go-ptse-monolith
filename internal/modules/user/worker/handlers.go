package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	userdomain "go-modular-monolith/internal/modules/user/domain"
	"go-modular-monolith/internal/shared/email"
	sharedworker "go-modular-monolith/internal/shared/worker"
)

// UserWorkerHandler processes user-related tasks
type UserWorkerHandler struct {
	userRepository userdomain.Repository
	emailService   email.EmailService
}

// NewUserWorkerHandler creates a new user worker handler
func NewUserWorkerHandler(userRepository userdomain.Repository, emailService email.EmailService) *UserWorkerHandler {
	return &UserWorkerHandler{
		userRepository: userRepository,
		emailService:   emailService,
	}
}

// HandleSendWelcomeEmail handles the welcome email task
func (h *UserWorkerHandler) HandleSendWelcomeEmail(ctx context.Context, payload sharedworker.TaskPayload) error {
	var p SendWelcomeEmailPayload

	// Unmarshal payload
	data, _ := json.Marshal(payload)
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Validate payload
	if p.UserID == "" || p.Email == "" {
		return fmt.Errorf("missing required fields in payload")
	}

	// Get user details
	user, err := h.userRepository.GetByID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", p.UserID)
	}

	// Send welcome email
	emailMsg := &email.Email{
		To:       []string{user.Email},
		Subject:  "Welcome to Our Platform!",
		HTMLBody: fmt.Sprintf("<h1>Welcome %s!</h1><p>Thank you for joining us. We're excited to have you on board.</p>", user.Name),
		TextBody: fmt.Sprintf("Welcome %s!\n\nThank you for joining us. We're excited to have you on board.", user.Name),
	}

	if err := h.emailService.Send(ctx, emailMsg); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	fmt.Printf("Sent welcome email to %s (%s)\n", user.Name, user.Email)

	return nil
}

// HandleSendPasswordResetEmail handles the password reset email task
func (h *UserWorkerHandler) HandleSendPasswordResetEmail(ctx context.Context, payload sharedworker.TaskPayload) error {
	var p SendPasswordResetEmailPayload

	data, _ := json.Marshal(payload)
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if p.UserID == "" || p.Email == "" || p.ResetLink == "" {
		return fmt.Errorf("missing required fields in payload")
	}

	user, err := h.userRepository.GetByID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", p.UserID)
	}

	// Send password reset email
	emailMsg := &email.Email{
		To:       []string{user.Email},
		Subject:  "Password Reset Request",
		HTMLBody: fmt.Sprintf("<h1>Password Reset</h1><p>Click <a href=\"%s\">here</a> to reset your password.</p>", p.ResetLink),
		TextBody: fmt.Sprintf("Password Reset\n\nClick the following link to reset your password:\n%s", p.ResetLink),
	}

	if err := h.emailService.Send(ctx, emailMsg); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	fmt.Printf("Sent password reset email to %s (%s)\n", user.Name, user.Email)

	return nil
}

// HandleExportUserData handles the user data export task
func (h *UserWorkerHandler) HandleExportUserData(ctx context.Context, payload sharedworker.TaskPayload) error {
	var p ExportUserDataPayload

	data, _ := json.Marshal(payload)
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if p.UserID == "" || p.Format == "" {
		return fmt.Errorf("missing required fields in payload")
	}

	// Get user
	user, err := h.userRepository.GetByID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", p.UserID)
	}

	// Export user data based on format
	var exportData interface{}
	switch p.Format {
	case "json":
		exportData = user
	case "csv":
		// Convert user to CSV format
		exportData = fmt.Sprintf("ID,Name,Email\n%s,%s,%s", user.ID, user.Name, user.Email)
	default:
		return fmt.Errorf("unsupported export format: %s", p.Format)
	}

	// Here you would store the export data in S3, local storage, or other destination
	// Example: h.storageService.UploadBytes(ctx, path, data, opts)
	// Then send notification to user with download link
	_ = exportData // Placeholder - implementation depends on storage backend

	fmt.Printf("Successfully exported user data for %s in format %s\n", p.UserID, p.Format)

	return nil
}

// HandleGenerateUserReport handles the user report generation task
func (h *UserWorkerHandler) HandleGenerateUserReport(ctx context.Context, payload sharedworker.TaskPayload) error {
	var p GenerateUserReportPayload

	data, _ := json.Marshal(payload)
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if p.UserID == "" || p.ReportType == "" {
		return fmt.Errorf("missing required fields in payload")
	}

	// Get user
	user, err := h.userRepository.GetByID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", p.UserID)
	}

	// Generate report based on type
	var reportContent string
	switch p.ReportType {
	case "activity":
		reportContent = fmt.Sprintf("Activity Report for %s\nUser: %s\nEmail: %s\nReport generated at: %s",
			p.UserID, user.Name, user.Email, time.Now().Format(time.RFC3339))
	case "summary":
		reportContent = fmt.Sprintf("Summary Report for %s\nUser: %s\nEmail: %s",
			p.UserID, user.Name, user.Email)
	default:
		return fmt.Errorf("unsupported report type: %s", p.ReportType)
	}

	// Store report in destination (S3, database, or file system)
	// Example: h.storageService.UploadBytes(ctx, reportPath, []byte(reportContent), opts)
	// Then send notification email to user with report details
	_ = reportContent // Placeholder - implementation depends on storage backend

	fmt.Printf("Successfully generated %s report for user %s\n", p.ReportType, p.UserID)

	return nil
}

// HandleSendMonthlyEmail handles the monthly email task sent to all users on the 15th
func (h *UserWorkerHandler) HandleSendMonthlyEmail(ctx context.Context, payload sharedworker.TaskPayload) error {
	var p SendMonthlyEmailPayload

	// Unmarshal payload
	data, _ := json.Marshal(payload)
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Default message if not provided
	if p.Message == "" {
		p.Message = "Today is the day"
	}

	fmt.Printf("[MONTHLY EMAIL] Starting monthly email task. Message: %s\n", p.Message)

	// Get all users
	users, err := h.userRepository.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("[MONTHLY EMAIL] No users found, skipping email send")
		return nil
	}

	fmt.Printf("[MONTHLY EMAIL] Found %d users to send emails to\n", len(users))

	// Send email to each user
	successCount := 0
	failureCount := 0

	for _, user := range users {
		emailMsg := &email.Email{
			To:      []string{user.Email},
			Subject: "Monthly Notification - Today is Special",
			TextBody: fmt.Sprintf("Hello %s,\n\n%s\n\nBest regards,\nThe Team",
				user.Name, p.Message),
			HTMLBody: fmt.Sprintf(`
				<html>
					<body>
						<h2>Hello %s,</h2>
						<p>%s</p>
						<p>Best regards,<br>The Team</p>
					</body>
				</html>
			`, user.Name, p.Message),
		}

		if err := h.emailService.Send(ctx, emailMsg); err != nil {
			fmt.Printf("[MONTHLY EMAIL] Failed to send email to %s (%s): %v\n", user.Name, user.Email, err)
			failureCount++
			continue
		}

		fmt.Printf("[MONTHLY EMAIL] Sent to %s (%s)\n", user.Name, user.Email)
		successCount++
	}

	if successCount == 0 && len(users) > 0 {
		return fmt.Errorf("failed to send email to any users: attempted %d, failed %d", len(users), failureCount)
	}

	fmt.Printf("[MONTHLY EMAIL] Task completed: %d sent, %d failed\n", successCount, failureCount)
	return nil
}
