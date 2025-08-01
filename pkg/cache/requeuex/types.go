package delayqueue

import (
	"context"
	"errors"
	"time"
)

// JobHandler defines the function signature for processing jobs
type JobHandler func(ctx context.Context, jobID string, payload string) error

// QueueConfig holds configuration for a delay queue
type QueueConfig struct {
	Name       string        `json:"name"`
	KeyPrefix  string        `json:"key_prefix"`
	MaxRetry   int           `json:"max_retry"`
	RetryDelay time.Duration `json:"retry_delay"`
	DLQKey     string        `json:"dlq_key,omitempty"` // optional
}

// Validate validates the QueueConfig
func (cfg *QueueConfig) Validate() error {
	if cfg.Name == "" {
		return errors.New("queue name cannot be empty")
	}
	if cfg.KeyPrefix == "" {
		return errors.New("key prefix cannot be empty")
	}
	if cfg.MaxRetry < 0 {
		return errors.New("max retry cannot be negative")
	}
	if cfg.RetryDelay < 0 {
		return errors.New("retry delay cannot be negative")
	}
	return nil
}

// JobData represents the data stored for each job
type JobData struct {
	Payload    string    `json:"payload"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Metrics holds queue performance metrics
type Metrics struct {
	JobsProcessed      int64         `json:"jobs_processed"`
	JobsFailed         int64         `json:"jobs_failed"`
	JobsRetried        int64         `json:"jobs_retried"`
	JobsMovedToDLQ     int64         `json:"jobs_moved_to_dlq"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	LastProcessedAt    time.Time     `json:"last_processed_at"`
}

// QueueStats holds current queue statistics
type QueueStats struct {
	QueueName    string    `json:"queue_name"`
	ActiveJobs   int64     `json:"active_jobs"`
	FailedJobs   int64     `json:"failed_jobs"`
	DLQSize      int64     `json:"dlq_size,omitempty"`
	LastActivity time.Time `json:"last_activity"`
	Metrics      Metrics   `json:"metrics"`
}
