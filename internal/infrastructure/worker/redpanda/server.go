package redpanda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	infraworker "go-modular-monolith/internal/infrastructure/worker"
	sharedworker "go-modular-monolith/internal/shared/worker"

	"github.com/segmentio/kafka-go"
)

// TaskMetadata holds retry and tracking information for tasks
type TaskMetadata struct {
	RetryCount      int                       `json:"retry_count"`
	OriginalOffset  int64                     `json:"original_offset"`
	OriginalTime    time.Time                 `json:"original_time"`
	LastError       string                    `json:"last_error"`
	CorrelationID   string                    `json:"correlation_id"`
	ProcessingSteps []string                  `json:"processing_steps"`
	RetryMetrics    *infraworker.RetryMetrics `json:"retry_metrics"`
}

// RedpandaServer is a Redpanda/Kafka-based implementation of the sharedworker.Server interface
type RedpandaServer struct {
	reader        *kafka.Reader
	retryWriter   *kafka.Writer // Writer for retry queue
	handlers      map[string]sharedworker.TaskHandler
	done          chan struct{}
	taskMetadata  map[string]*TaskMetadata // Track metadata for failed tasks
	metadataMutex sync.RWMutex
	retryPolicy   infraworker.RetryPolicy
	dlqWriter     *kafka.Writer
	topic         string
}

// NewRedpandaServer creates a new Redpanda server with retry policy
func NewRedpandaServer(brokers []string, topic, consumerGroup string, workerCount int) *RedpandaServer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        consumerGroup,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
		MaxBytes:       10e6, // 10MB
	})

	retryWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic + "-retry",
		Balancer: &kafka.LeastBytes{},
	}

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic + "-dlq",
		Balancer: &kafka.LeastBytes{},
	}

	return &RedpandaServer{
		reader:       reader,
		retryWriter:  retryWriter,
		handlers:     make(map[string]sharedworker.TaskHandler),
		done:         make(chan struct{}),
		taskMetadata: make(map[string]*TaskMetadata),
		retryPolicy:  infraworker.DefaultRetryPolicy(),
		dlqWriter:    dlqWriter,
		topic:        topic,
	}
}

// SetRetryPolicy sets the retry policy for the server
func (s *RedpandaServer) SetRetryPolicy(policy infraworker.RetryPolicy) {
	s.retryPolicy = policy
}

// RegisterHandler registers a handler for a task type
func (s *RedpandaServer) RegisterHandler(taskName string, handler sharedworker.TaskHandler) error {
	s.handlers[taskName] = handler
	return nil
}

// Start starts the Redpanda worker server with retry mechanism
func (s *RedpandaServer) Start(ctx context.Context) error {
	log.Println("Starting Redpanda worker server...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
		}

		msg, err := s.reader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return fmt.Errorf("failed to read message: %w", err)
		}

		// Get handler for this task
		taskName := string(msg.Key)
		handler, ok := s.handlers[taskName]
		if !ok {
			log.Printf("No handler registered for task: %s\n", taskName)
			continue
		}

		// Parse payload
		var payload sharedworker.TaskPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			// Payload is invalid, send to DLQ
			s.sendToDeadLetterTopic(ctx, taskName, msg, fmt.Errorf("invalid payload: %w", err), nil)
			continue
		}

		// Get or create metadata for tracking
		taskID := s.getTaskID(msg)
		metadata := s.getTaskMetadata(taskID)
		metadata.ProcessingSteps = append(metadata.ProcessingSteps,
			fmt.Sprintf("attempt_%d_at_%s", metadata.RetryCount+1, time.Now().Format(time.RFC3339)))

		// Process the task
		if err := handler(ctx, payload); err != nil {
			metadata.LastError = err.Error()
			metadata.RetryCount++

			// Check if we should retry
			if s.retryPolicy.ShouldRetry(metadata.RetryCount-1, err.Error()) {
				// Calculate backoff
				backoff := s.retryPolicy.CalculateBackoff(metadata.RetryCount)

				log.Printf("Task %s failed (attempt %d), retrying in %v: %v\n",
					taskName, metadata.RetryCount, backoff, err)

				// Enqueue for retry with delay
				s.requeueForRetry(ctx, taskName, msg, backoff, metadata)
				s.removeTaskMetadata(taskID)
			} else {
				// Send to DLQ
				log.Printf("Task %s failed after %d attempts, moving to DLQ: %v\n",
					taskName, metadata.RetryCount, err)
				s.sendToDeadLetterTopic(ctx, taskName, msg, err, metadata)
				s.removeTaskMetadata(taskID)
			}
			continue
		}

		// Task succeeded, clean up metadata
		log.Printf("Task %s completed successfully after %d attempts\n", taskName, metadata.RetryCount)
		s.removeTaskMetadata(taskID)
	}
}

// Stop gracefully stops the Redpanda worker server
func (s *RedpandaServer) Stop(ctx context.Context) error {
	close(s.done)
	if s.retryWriter != nil {
		s.retryWriter.Close()
	}
	if s.dlqWriter != nil {
		s.dlqWriter.Close()
	}
	return s.reader.Close()
}

// requeueForRetry sends a task to the retry queue with backoff metadata
func (s *RedpandaServer) requeueForRetry(ctx context.Context, taskName string, msg kafka.Message, backoff time.Duration, metadata *TaskMetadata) error {
	retryMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers,
			kafka.Header{Key: "retry_attempt", Value: []byte(fmt.Sprintf("%d", metadata.RetryCount))},
			kafka.Header{Key: "scheduled_for", Value: []byte(time.Now().Add(backoff).Format(time.RFC3339))},
			kafka.Header{Key: "backoff_ms", Value: []byte(fmt.Sprintf("%d", backoff.Milliseconds()))},
			kafka.Header{Key: "correlation_id", Value: []byte(metadata.CorrelationID)},
			kafka.Header{Key: "last_error", Value: []byte(metadata.LastError)},
		),
	}

	return s.retryWriter.WriteMessages(ctx, retryMsg)
}

// sendToDeadLetterTopic sends failed tasks to a dead-letter topic with full metadata
func (s *RedpandaServer) sendToDeadLetterTopic(ctx context.Context, taskName string, msg kafka.Message, err error, metadata *TaskMetadata) error {
	// Build comprehensive metadata for DLQ
	dlqMetadata := TaskMetadata{
		RetryCount:      metadata.RetryCount,
		OriginalOffset:  msg.Offset,
		OriginalTime:    time.Now(),
		LastError:       err.Error(),
		CorrelationID:   s.getCorrelationID(msg),
		ProcessingSteps: metadata.ProcessingSteps,
	}

	// Marshal metadata as JSON
	metadataJSON, _ := json.Marshal(dlqMetadata)

	// Add comprehensive error information to headers
	errorMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers,
			kafka.Header{Key: "error", Value: []byte(err.Error())},
			kafka.Header{Key: "retry_count", Value: []byte(fmt.Sprintf("%d", dlqMetadata.RetryCount))},
			kafka.Header{Key: "original_offset", Value: []byte(fmt.Sprintf("%d", msg.Offset))},
			kafka.Header{Key: "original_topic", Value: []byte(msg.Topic)},
			kafka.Header{Key: "correlation_id", Value: []byte(dlqMetadata.CorrelationID)},
			kafka.Header{Key: "metadata", Value: metadataJSON},
			kafka.Header{Key: "dlq_timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		),
	}

	return s.dlqWriter.WriteMessages(ctx, errorMsg)
}

// getTaskID generates a unique identifier for task tracking
func (s *RedpandaServer) getTaskID(msg kafka.Message) string {
	return fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)
}

// getCorrelationID extracts or generates a correlation ID from message headers
func (s *RedpandaServer) getCorrelationID(msg kafka.Message) string {
	for _, h := range msg.Headers {
		if h.Key == "correlation_id" {
			return string(h.Value)
		}
	}
	// Generate new correlation ID if not present
	return fmt.Sprintf("%s-%d", time.Now().Format(time.RFC3339Nano), msg.Offset)
}

// getTaskMetadata retrieves or creates metadata for a task
func (s *RedpandaServer) getTaskMetadata(taskID string) *TaskMetadata {
	s.metadataMutex.Lock()
	defer s.metadataMutex.Unlock()

	if metadata, ok := s.taskMetadata[taskID]; ok {
		return metadata
	}

	metadata := &TaskMetadata{
		RetryCount:      0,
		OriginalTime:    time.Now(),
		ProcessingSteps: []string{},
	}
	s.taskMetadata[taskID] = metadata
	return metadata
}

// removeTaskMetadata cleans up metadata for a task
func (s *RedpandaServer) removeTaskMetadata(taskID string) {
	s.metadataMutex.Lock()
	defer s.metadataMutex.Unlock()
	delete(s.taskMetadata, taskID)
}
