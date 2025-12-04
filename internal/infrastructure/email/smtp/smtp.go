package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/kamil5b/go-ptse-monolith/internal/infrastructure/email/template"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/email"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	FromAddr string
	FromName string
}

// SMTPEmailService sends emails via SMTP
type SMTPEmailService struct {
	config         SMTPConfig
	addr           string
	templateLoader *template.TemplateLoader
}

// NewSMTPEmailService creates a new SMTP email service with template support
func NewSMTPEmailService(config SMTPConfig) *SMTPEmailService {
	loader := template.NewTemplateLoader()

	// Register default templates
	_ = loader.RegisterTemplate(
		"welcome",
		"Welcome {{.name}}!",
		`<h1>Welcome {{.name}}!</h1><p>Thank you for joining us.</p>`,
		`Welcome {{.name}}!\n\nThank you for joining us.`,
		[]string{"name"},
	)

	_ = loader.RegisterTemplate(
		"password_reset",
		"Password Reset Request",
		`<h1>Reset Your Password</h1><p><a href="{{.reset_link}}">Click here</a> to reset your password.</p>`,
		`Reset Your Password\n\nClick the link: {{.reset_link}}`,
		[]string{"reset_link"},
	)

	return &SMTPEmailService{
		config:         config,
		addr:           fmt.Sprintf("%s:%d", config.Host, config.Port),
		templateLoader: loader,
	}
}

// Send sends a single email via SMTP
func (s *SMTPEmailService) Send(ctx context.Context, e *email.Email) error {
	if e == nil {
		return fmt.Errorf("email cannot be nil")
	}

	if len(e.To) == 0 {
		return fmt.Errorf("email must have at least one recipient")
	}

	// Validate recipients
	for _, addr := range e.To {
		if err := s.ValidateEmail(addr); err != nil {
			return fmt.Errorf("invalid recipient: %w", err)
		}
	}

	// Create message
	msg := s.buildMessage(e)

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Send email
	allRecipients := append(append(e.To, e.CC...), e.BCC...)
	if err := smtp.SendMail(s.addr, auth, e.From, allRecipients, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendBatch sends multiple emails in batch
func (s *SMTPEmailService) SendBatch(ctx context.Context, emails []*email.Email) error {
	if len(emails) == 0 {
		return fmt.Errorf("emails cannot be empty")
	}

	for i, e := range emails {
		if err := s.Send(ctx, e); err != nil {
			return fmt.Errorf("failed to send email %d: %w", i, err)
		}
	}

	return nil
}

// SendTemplate sends an email using a registered template
func (s *SMTPEmailService) SendTemplate(ctx context.Context, to []string, templateID string, data map[string]interface{}) error {
	if len(to) == 0 {
		return fmt.Errorf("recipients cannot be empty")
	}

	// Render template using the template loader
	subject, htmlBody, textBody, err := s.templateLoader.RenderTemplate(templateID, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	e := &email.Email{
		To:       to,
		From:     s.config.FromAddr,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return s.Send(ctx, e)
}

// RegisterTemplate allows registering custom email templates at runtime
func (s *SMTPEmailService) RegisterTemplate(name, subject, htmlBody, textBody string, requiredKeys []string) error {
	return s.templateLoader.RegisterTemplate(name, subject, htmlBody, textBody, requiredKeys)
}

// ValidateEmail validates email format
func (s *SMTPEmailService) ValidateEmail(addr string) error {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	if !re.MatchString(addr) {
		return fmt.Errorf("invalid email format: %s", addr)
	}
	return nil
}

// Health checks SMTP connectivity
func (s *SMTPEmailService) Health(ctx context.Context) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Try to connect
	client, err := smtp.Dial(s.addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Try to authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate with SMTP server: %w", err)
	}

	return nil
}

// buildMessage constructs the email message with headers and body
func (s *SMTPEmailService) buildMessage(e *email.Email) string {
	headers := []string{
		fmt.Sprintf("To: %s", strings.Join(e.To, ", ")),
		fmt.Sprintf("From: %s", e.From),
		fmt.Sprintf("Subject: %s", e.Subject),
		"MIME-Version: 1.0",
	}

	if len(e.CC) > 0 {
		headers = append(headers, fmt.Sprintf("CC: %s", strings.Join(e.CC, ", ")))
	}

	if e.ReplyTo != "" {
		headers = append(headers, fmt.Sprintf("Reply-To: %s", e.ReplyTo))
	}

	// Add custom headers
	for k, v := range e.Headers {
		headers = append(headers, fmt.Sprintf("%s: %s", k, v))
	}

	// Set content type based on HTML or text
	contentType := "text/plain; charset=UTF-8"
	body := e.TextBody
	if e.HTMLBody != "" {
		contentType = "text/html; charset=UTF-8"
		body = e.HTMLBody
	}
	headers = append(headers, fmt.Sprintf("Content-Type: %s", contentType))

	// Combine headers and body
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}
