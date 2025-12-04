package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoOpEmailService(t *testing.T) {
	service := NewNoOpEmailService()
	require.NotNil(t, service)
}

func TestNoOpEmailServiceSend(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	email := &Email{
		To:       []string{"test@example.com"},
		From:     "sender@example.com",
		Subject:  "Test Subject",
		HTMLBody: "Test Body",
	}

	err := service.Send(ctx, email)
	assert.NoError(t, err)
}

func TestNoOpEmailServiceSendNil(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	err := service.Send(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be nil")
}

func TestNoOpEmailServiceSendBatch(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	emails := []*Email{
		{
			To:      []string{"test1@example.com"},
			Subject: "Test 1",
		},
		{
			To:      []string{"test2@example.com"},
			Subject: "Test 2",
		},
	}

	err := service.SendBatch(ctx, emails)
	assert.NoError(t, err)
}

func TestNoOpEmailServiceSendBatchNil(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	err := service.SendBatch(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "emails cannot be nil")
}

func TestNoOpEmailServiceSendTemplate(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	err := service.SendTemplate(ctx, []string{"test@example.com"}, "welcome", map[string]interface{}{"name": "John"})
	assert.NoError(t, err)
}

func TestNoOpEmailServiceSendTemplateEmptyTo(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	err := service.SendTemplate(ctx, []string{}, "welcome", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to cannot be empty")
}

func TestNoOpEmailServiceValidateEmailValid(t *testing.T) {
	service := NewNoOpEmailService()

	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"simple", "user@example.com", true},
		{"with_dot", "user.name@example.com", true},
		{"with_plus", "user+tag@example.co.uk", true},
		{"with_underscore", "user_name@example.com", true},
		{"with_dash", "user-name@example-domain.com", true},
		{"invalid_no_at", "user.example.com", false},
		{"invalid_no_domain", "user@", false},
		{"invalid_empty", "", false},
		{"invalid_spaces", "user @example.com", false},
		{"invalid_multiple_at", "user@@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateEmail(tt.email)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNoOpEmailServiceHealth(t *testing.T) {
	service := NewNoOpEmailService()
	ctx := context.Background()

	err := service.Health(ctx)
	assert.NoError(t, err)
}
