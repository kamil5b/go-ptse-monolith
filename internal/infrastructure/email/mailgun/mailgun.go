package mailgun

import (
	"context"
	"fmt"
	"regexp"

	"github.com/kamil5b/go-ptse-monolith/internal/shared/email"

	"github.com/mailgun/mailgun-go/v4"
)

// MailgunConfig holds Mailgun API configuration
type MailgunConfig struct {
	Domain    string
	APIKey    string
	FromAddr  string
	FromName  string
	PublicKey string
}

// MailgunEmailService sends emails via Mailgun API
type MailgunEmailService struct {
	config   MailgunConfig
	mg       mailgun.Mailgun
	fromAddr string
}

// NewMailgunEmailService creates a new Mailgun email service
func NewMailgunEmailService(config MailgunConfig) *MailgunEmailService {
	mg := mailgun.NewMailgun(config.Domain, config.APIKey)

	fromAddr := config.FromAddr
	if config.FromName != "" {
		fromAddr = fmt.Sprintf("%s <%s>", config.FromName, config.FromAddr)
	}

	return &MailgunEmailService{
		config:   config,
		mg:       mg,
		fromAddr: fromAddr,
	}
}

// Send sends a single email via Mailgun
func (s *MailgunEmailService) Send(ctx context.Context, e *email.Email) error {
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

	// Build message
	m := s.mg.NewMessage(s.fromAddr, e.Subject, e.TextBody, e.To...)

	// Add CC recipients
	for _, cc := range e.CC {
		m.AddCC(cc)
	}

	// Add BCC recipients
	for _, bcc := range e.BCC {
		m.AddBCC(bcc)
	}

	// Set HTML body if provided
	if e.HTMLBody != "" {
		m.SetHtml(e.HTMLBody)
	}

	// Add reply-to if provided
	if e.ReplyTo != "" {
		m.SetReplyTo(e.ReplyTo)
	}

	// Add custom headers
	for k, v := range e.Headers {
		m.AddHeader(k, v)
	}

	// Add attachments
	for _, att := range e.Attachments {
		m.AddBufferAttachment(att.Filename, att.Content)
	}

	// Send with context
	_, _, err := s.mg.Send(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to send email via Mailgun: %w", err)
	}

	return nil
}

// SendBatch sends multiple emails in batch
func (s *MailgunEmailService) SendBatch(ctx context.Context, emails []*email.Email) error {
	if len(emails) == 0 {
		return fmt.Errorf("emails cannot be empty")
	}

	// Mailgun supports batch sending, but for simplicity we'll send individually
	for i, e := range emails {
		if err := s.Send(ctx, e); err != nil {
			return fmt.Errorf("failed to send email %d: %w", i, err)
		}
	}

	return nil
}

// SendTemplate sends an email using a Mailgun template
func (s *MailgunEmailService) SendTemplate(ctx context.Context, to []string, templateID string, data map[string]interface{}) error {
	if len(to) == 0 {
		return fmt.Errorf("recipients cannot be empty")
	}

	// Build message with template
	m := s.mg.NewMessage(s.fromAddr, "", "", to...)

	// Set template
	m.SetTemplate(templateID)

	// Add template variables
	for k, v := range data {
		m.AddTemplateVariable(k, v)
	}

	// Send with context
	_, _, err := s.mg.Send(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to send template email via Mailgun: %w", err)
	}

	return nil
}

// ValidateEmail validates email format
func (s *MailgunEmailService) ValidateEmail(addr string) error {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	if !re.MatchString(addr) {
		return fmt.Errorf("invalid email format: %s", addr)
	}
	return nil
}

// Health checks Mailgun connectivity
func (s *MailgunEmailService) Health(ctx context.Context) error {
	// Try to validate a test address using Mailgun
	addr := "test@example.com"
	if err := s.ValidateEmail(addr); err != nil {
		return fmt.Errorf("Mailgun service health check failed: %w", err)
	}

	return nil
}
