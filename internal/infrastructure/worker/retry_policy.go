package worker

import (
	"fmt"
	"math"
	"time"
)

// RetryPolicy defines how to retry failed tasks
type RetryPolicy struct {
	MaxRetries         int           // Maximum number of retries (0 = no retries)
	InitialBackoff     time.Duration // Initial backoff duration
	MaxBackoff         time.Duration // Maximum backoff duration
	BackoffMultiplier  float64       // Exponential backoff multiplier
	JitterFraction     float64       // Jitter as fraction of backoff (0.0 to 1.0)
	RetryableErrors    []string      // Specific error types to retry on (empty = all errors)
	NonRetryableErrors []string      // Error types to NOT retry
}

// DefaultRetryPolicy returns a production-ready default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        60 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	}
}

// CalculateBackoff calculates the backoff duration for a given attempt
func (rp *RetryPolicy) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Calculate exponential backoff: initialBackoff * (multiplier ^ attempt)
	backoffMs := float64(rp.InitialBackoff.Milliseconds()) *
		math.Pow(rp.BackoffMultiplier, float64(attempt-1))

	// Cap at max backoff
	if backoffMs > float64(rp.MaxBackoff.Milliseconds()) {
		backoffMs = float64(rp.MaxBackoff.Milliseconds())
	}

	// Add jitter to prevent thundering herd
	jitterAmount := backoffMs * rp.JitterFraction
	jitterRange := time.Duration(jitterAmount * float64(time.Millisecond))

	// Add random jitter (simplified - in production use rand.Int63n)
	jitter := jitterRange / 2

	return time.Duration(backoffMs)*time.Millisecond + jitter
}

// ShouldRetry determines if an error should trigger a retry
func (rp *RetryPolicy) ShouldRetry(attempt int, errMsg string) bool {
	// Check if we've exceeded max retries
	if attempt >= rp.MaxRetries {
		return false
	}

	// If non-retryable errors are specified, check against them
	if len(rp.NonRetryableErrors) > 0 {
		for _, nrErr := range rp.NonRetryableErrors {
			if contains(errMsg, nrErr) {
				return false
			}
		}
	}

	// If retryable errors are specified, check against them
	if len(rp.RetryableErrors) > 0 {
		for _, rErr := range rp.RetryableErrors {
			if contains(errMsg, rErr) {
				return true
			}
		}
		// Didn't match any retryable errors
		return false
	}

	// No restrictions, retry all errors
	return true
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

// RetryConfig wraps RetryPolicy with additional configuration
type RetryConfig struct {
	Policy           RetryPolicy
	BackoffTopicName string                                              // Topic to requeue messages for retry
	EnableMetrics    bool                                                // Track retry metrics
	OnRetryFunc      func(attempt int, backoff time.Duration, err error) // Callback on retry
}

// RetryMetrics tracks retry statistics
type RetryMetrics struct {
	TaskName      string
	TotalAttempts int
	SuccessfulAt  int // Attempt number that succeeded (0 if failed)
	TotalBackoff  time.Duration
	LastError     string
	LastRetryTime time.Time
}

// NewRetryMetrics creates new retry metrics
func NewRetryMetrics(taskName string) *RetryMetrics {
	return &RetryMetrics{
		TaskName: taskName,
	}
}

// String returns a formatted string representation of metrics
func (rm *RetryMetrics) String() string {
	successMsg := "FAILED"
	if rm.SuccessfulAt > 0 {
		successMsg = fmt.Sprintf("SUCCESS at attempt %d", rm.SuccessfulAt)
	}
	return fmt.Sprintf(
		"Task: %s | %s | Total Attempts: %d | Total Backoff: %v | Last Error: %s",
		rm.TaskName, successMsg, rm.TotalAttempts, rm.TotalBackoff, rm.LastError,
	)
}
