package redpanda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// DelayedTaskScheduler monitors delayed tasks and schedules them for processing
type DelayedTaskScheduler struct {
	delayedReader *kafka.Reader
	taskWriter    *kafka.Writer
	done          chan struct{}
	checkInterval time.Duration
}

// NewDelayedTaskScheduler creates a new delayed task scheduler
func NewDelayedTaskScheduler(brokers []string, delayedTopic, mainTopic string) *DelayedTaskScheduler {
	delayedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          delayedTopic,
		GroupID:        "task-scheduler",
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
		MaxBytes:       10e6,
	})

	taskWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    mainTopic,
		Balancer: &kafka.LeastBytes{},
	}

	return &DelayedTaskScheduler{
		delayedReader: delayedReader,
		taskWriter:    taskWriter,
		done:          make(chan struct{}),
		checkInterval: 5 * time.Second, // Check every 5 seconds
	}
}

// Start starts the delayed task scheduler
func (s *DelayedTaskScheduler) Start(ctx context.Context) error {
	log.Println("Starting delayed task scheduler...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
		}

		msg, err := s.delayedReader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			log.Printf("failed to read delayed message: %v\n", err)
			continue
		}

		// Extract scheduled time from headers
		scheduledTime := s.getScheduledTime(msg)
		now := time.Now().Unix()

		if scheduledTime <= now {
			// Task is ready to be processed
			if err := s.promoteToMainTopic(ctx, msg); err != nil {
				log.Printf("failed to promote task to main topic: %v\n", err)
				continue
			}
		} else {
			// Task is not ready yet, requeue it with metadata
			waitTime := time.Duration(scheduledTime-now) * time.Second
			if err := s.requeueDelayedTask(ctx, msg, waitTime); err != nil {
				log.Printf("failed to requeue delayed task: %v\n", err)
			}
		}
	}
}

// Stop stops the scheduler
func (s *DelayedTaskScheduler) Stop(ctx context.Context) error {
	close(s.done)
	if err := s.taskWriter.Close(); err != nil {
		return err
	}
	return s.delayedReader.Close()
}

// getScheduledTime extracts the scheduled time from message headers
func (s *DelayedTaskScheduler) getScheduledTime(msg kafka.Message) int64 {
	for _, h := range msg.Headers {
		if h.Key == "scheduled_at" {
			var timestamp int64
			if err := json.Unmarshal(h.Value, &timestamp); err == nil {
				return timestamp
			}
			// Try parsing as string
			fmt.Sscanf(string(h.Value), "%d", &timestamp)
			return timestamp
		}
	}
	return time.Now().Unix() // Default to now if not found
}

// promoteToMainTopic moves a task from delayed queue to main queue
func (s *DelayedTaskScheduler) promoteToMainTopic(ctx context.Context, msg kafka.Message) error {
	// Add metadata about when it was promoted
	promotedMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers,
			kafka.Header{Key: "promoted_at", Value: []byte(time.Now().Format(time.RFC3339))},
			kafka.Header{Key: "was_delayed", Value: []byte("true")},
		),
	}

	return s.taskWriter.WriteMessages(ctx, promotedMsg)
}

// requeueDelayedTask requeues a task that's not yet ready
func (s *DelayedTaskScheduler) requeueDelayedTask(ctx context.Context, msg kafka.Message, waitTime time.Duration) error {
	// Add exponential backoff header
	var backoff int64 = waitTime.Milliseconds()
	newHeaders := append(msg.Headers, kafka.Header{Key: "backoff_ms", Value: []byte(fmt.Sprintf("%d", backoff))})

	// Optionally, update scheduled_at to new time
	newScheduledAt := time.Now().Add(waitTime).Unix()
	for i, h := range newHeaders {
		if h.Key == "scheduled_at" {
			newHeaders[i].Value, _ = json.Marshal(newScheduledAt)
			goto writeMsg
		}
	}
	newHeaders = append(newHeaders, kafka.Header{Key: "scheduled_at", Value: []byte(fmt.Sprintf("%d", newScheduledAt))})

writeMsg:
	requeuedMsg := kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: newHeaders,
	}

	// Write back to delayed topic
	return s.taskWriter.WriteMessages(ctx, requeuedMsg)
}
