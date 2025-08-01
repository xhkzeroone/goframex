package delayqueue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.io/xhkzeroone/goframex/pkg/cache/redisx"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Queue struct {
	name       string
	prefix     string
	redis      *redisx.Redis
	handler    JobHandler
	maxRetry   int
	retryDelay time.Duration
	dlqKey     string
	logger     *logrus.Logger

	// Metrics
	metrics struct {
		jobsProcessed    int64
		jobsFailed       int64
		jobsRetried      int64
		jobsMovedToDLQ   int64
		totalProcessTime int64 // nanoseconds
		lastProcessedAt  time.Time
	}

	// Concurrency control
	mu sync.RWMutex
}

// generateUUID generates a random UUID
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// newQueue creates a new queue instance
func newQueue(cfg QueueConfig, redis *redisx.Redis, handler JobHandler, logger *logrus.Logger) (*Queue, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid queue config: %w", err)
	}

	if redis == nil {
		return nil, errors.New("redis backend cannot be nil")
	}

	if handler == nil {
		return nil, errors.New("job handler cannot be nil")
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &Queue{
		name:       cfg.Name,
		prefix:     cfg.KeyPrefix,
		redis:      redis,
		handler:    handler,
		maxRetry:   cfg.MaxRetry,
		retryDelay: cfg.RetryDelay,
		dlqKey:     cfg.DLQKey,
		logger:     logger,
	}, nil
}

// Push adds a new job to the queue with the specified delay
func (q *Queue) Push(ctx context.Context, payload string, delay time.Duration) (string, error) {
	// Validation
	if payload == "" {
		return "", errors.New("payload cannot be empty")
	}
	if delay <= 0 {
		return "", errors.New("delay must be positive")
	}

	// Generate UUID for the job
	uuid := generateUUID()
	triggerKey := q.prefix + ":trigger:" + uuid
	dataKey := q.prefix + ":data:" + uuid

	data := JobData{
		Payload:    payload,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Use pipeline to set both keys atomically
	pipe := q.redis.Pipeline()
	pipe.Set(ctx, dataKey, jsonData, 0)   // No expiration for data
	pipe.Set(ctx, triggerKey, "1", delay) // Only trigger key has expiration

	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to set job in Redis: %w", err)
	}

	q.logger.WithFields(logrus.Fields{
		"queue":      q.name,
		"uuid":       uuid,
		"triggerKey": triggerKey,
		"dataKey":    dataKey,
		"delay":      delay,
		"payload":    payload,
	}).Info("Job pushed to queue")

	return uuid, nil
}

// handleExpiredKey processes an expired key from Redis
func (q *Queue) handleExpiredKey(ctx context.Context, fullKey string) {
	if !strings.HasPrefix(fullKey, q.prefix+":trigger:") {
		return
	}

	// Extract UUID from key format: {prefix}:trigger:{uuid}
	parts := strings.SplitN(fullKey, ":", 3)
	if len(parts) != 3 || parts[1] != "trigger" {
		q.logger.WithField("key", fullKey).Warn("Invalid trigger key format")
		return
	}

	uuid := parts[2]
	dataKey := q.prefix + ":data:" + uuid
	startTime := time.Now()

	q.logger.WithFields(logrus.Fields{
		"queue":      q.name,
		"uuid":       uuid,
		"triggerKey": fullKey,
		"dataKey":    dataKey,
	}).Debug("Processing expired job")

	// Get job data from Redis - try multiple times in case of race conditions
	var val string
	var err error
	var job JobData

	// Try to get the job data, with retries for race conditions
	for attempt := 0; attempt < 3; attempt++ {
		val, err = q.redis.Get(ctx, dataKey).Result()
		if err == nil {
			// Successfully got the data
			break
		}

		if errors.Is(err, redis.Nil) {
			// Job not found - this can happen if the job was deleted
			// or if there's a race condition between expiration and our retrieval
			if attempt < 2 {
				// Wait a bit and retry
				time.Sleep(100 * time.Millisecond)
				continue
			}
			q.logger.WithFields(logrus.Fields{
				"queue":    q.name,
				"uuid":     uuid,
				"attempts": attempt + 1,
			}).Warn("Job data not found in Redis after retries (may have been deleted)")
			return
		}

		// Other Redis error
		q.logger.WithFields(logrus.Fields{
			"queue": q.name,
			"uuid":  uuid,
			"error": err,
		}).Error("Failed to get job from Redis")
		return
	}

	// Parse job data
	if err := json.Unmarshal([]byte(val), &job); err != nil {
		q.logger.WithFields(logrus.Fields{
			"queue": q.name,
			"uuid":  uuid,
			"error": err,
			"data":  val,
		}).Error("Failed to unmarshal job data")
		return
	}

	// Process the job - use UUID as jobID for handler
	err = q.handler(ctx, uuid, job.Payload)
	processTime := time.Since(startTime)

	if err != nil {
		atomic.AddInt64(&q.metrics.jobsFailed, 1)
		q.logger.WithFields(logrus.Fields{
			"queue":        q.name,
			"uuid":         uuid,
			"error":        err,
			"retry_count":  job.RetryCount,
			"max_retry":    q.maxRetry,
			"process_time": processTime,
		}).Error("Job processing failed")

		if job.RetryCount < q.maxRetry {
			// Retry the job
			atomic.AddInt64(&q.metrics.jobsRetried, 1)
			if err := q.retryJob(ctx, uuid, job); err != nil {
				q.logger.WithFields(logrus.Fields{
					"queue": q.name,
					"uuid":  uuid,
					"error": err,
				}).Error("Failed to retry job")
			}
		} else {
			// Move to DLQ or log permanent failure
			if err := q.moveToDLQ(ctx, uuid, job, err); err != nil {
				q.logger.WithFields(logrus.Fields{
					"queue": q.name,
					"uuid":  uuid,
					"error": err,
				}).Error("Failed to move job to DLQ")
			}
		}
	} else {
		atomic.AddInt64(&q.metrics.jobsProcessed, 1)
		q.logger.WithFields(logrus.Fields{
			"queue":        q.name,
			"uuid":         uuid,
			"process_time": processTime,
		}).Info("Job completed successfully")

		// Clean up data key after successful completion
		if err := q.redis.Del(ctx, dataKey).Err(); err != nil {
			q.logger.WithFields(logrus.Fields{
				"queue": q.name,
				"uuid":  uuid,
				"error": err,
			}).Warn("Failed to cleanup job data after successful completion")
		}
	}

	// Update metrics
	atomic.AddInt64(&q.metrics.totalProcessTime, int64(processTime))
	q.mu.Lock()
	q.metrics.lastProcessedAt = time.Now()
	q.mu.Unlock()
}

// retryJob retries a failed job
func (q *Queue) retryJob(ctx context.Context, uuid string, job JobData) error {
	job.RetryCount++
	job.UpdatedAt = time.Now()

	newVal, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal retry job data: %w", err)
	}

	retryTriggerKey := q.prefix + ":trigger:" + uuid
	retryDataKey := q.prefix + ":data:" + uuid

	// Use pipeline to update both keys atomically
	pipe := q.redis.Pipeline()
	pipe.Set(ctx, retryDataKey, newVal, 0)            // Update data
	pipe.Set(ctx, retryTriggerKey, "1", q.retryDelay) // Set new trigger with retry delay

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set retry job in Redis: %w", err)
	}

	q.logger.WithFields(logrus.Fields{
		"queue":       q.name,
		"uuid":        uuid,
		"retry_count": job.RetryCount,
		"max_retry":   q.maxRetry,
		"delay":       q.retryDelay,
	}).Info("Job scheduled for retry")

	return nil
}

// moveToDLQ moves a permanently failed job to the dead letter queue
func (q *Queue) moveToDLQ(ctx context.Context, uuid string, job JobData, processError error) error {
	atomic.AddInt64(&q.metrics.jobsMovedToDLQ, 1)

	if q.dlqKey == "" {
		q.logger.WithFields(logrus.Fields{
			"queue": q.name,
			"uuid":  uuid,
		}).Warn("Job failed permanently (no DLQ configured)")
		return nil
	}

	dlqEntry := map[string]interface{}{
		"uuid":        uuid,
		"payload":     job.Payload,
		"retry_count": job.RetryCount,
		"error":       processError.Error(),
		"failed_at":   time.Now().Format(time.RFC3339),
		"queue_name":  q.name,
		"created_at":  job.CreatedAt.Format(time.RFC3339),
		"updated_at":  job.UpdatedAt.Format(time.RFC3339),
	}

	entryJSON, err := json.Marshal(dlqEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ entry: %w", err)
	}

	if err := q.redis.RPush(ctx, q.dlqKey, entryJSON).Err(); err != nil {
		return fmt.Errorf("failed to push to DLQ: %w", err)
	}

	// Clean up data key after moving to DLQ
	dataKey := q.prefix + ":data:" + uuid
	if err := q.redis.Del(ctx, dataKey).Err(); err != nil {
		q.logger.WithFields(logrus.Fields{
			"queue": q.name,
			"uuid":  uuid,
			"error": err,
		}).Warn("Failed to cleanup job data after moving to DLQ")
	}

	q.logger.WithFields(logrus.Fields{
		"queue": q.name,
		"uuid":  uuid,
		"dlq":   q.dlqKey,
	}).Info("Job moved to dead letter queue")

	return nil
}

// GetStats returns current queue statistics
func (q *Queue) GetStats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var avgProcessTime time.Duration
	if q.metrics.jobsProcessed > 0 {
		avgProcessTime = time.Duration(q.metrics.totalProcessTime / q.metrics.jobsProcessed)
	}

	metrics := Metrics{
		JobsProcessed:      q.metrics.jobsProcessed,
		JobsFailed:         q.metrics.jobsFailed,
		JobsRetried:        q.metrics.jobsRetried,
		JobsMovedToDLQ:     q.metrics.jobsMovedToDLQ,
		AverageProcessTime: avgProcessTime,
		LastProcessedAt:    q.metrics.lastProcessedAt,
	}

	return QueueStats{
		QueueName:    q.name,
		LastActivity: q.metrics.lastProcessedAt,
		Metrics:      metrics,
	}
}

// processJobFromDataKey processes a job directly from its data key
func (q *Queue) processJobFromDataKey(ctx context.Context, dataKey string, uuid string) error {
	startTime := time.Now()

	q.logger.WithFields(logrus.Fields{
		"queue":   q.name,
		"uuid":    uuid,
		"dataKey": dataKey,
	}).Debug("Processing job from data key")

	// Get job data from Redis
	val, err := q.redis.Get(ctx, dataKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			q.logger.WithFields(logrus.Fields{
				"queue":   q.name,
				"uuid":    uuid,
				"dataKey": dataKey,
			}).Warn("Job data not found in Redis")
			return nil
		}
		return fmt.Errorf("failed to get job data from Redis: %w", err)
	}

	// Parse job data
	var job JobData
	if err := json.Unmarshal([]byte(val), &job); err != nil {
		q.logger.WithFields(logrus.Fields{
			"queue": q.name,
			"uuid":  uuid,
			"error": err,
			"data":  val,
		}).Error("Failed to unmarshal job data")
		return fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	// Process the job
	err = q.handler(ctx, uuid, job.Payload)
	processTime := time.Since(startTime)

	if err != nil {
		atomic.AddInt64(&q.metrics.jobsFailed, 1)
		q.logger.WithFields(logrus.Fields{
			"queue":        q.name,
			"uuid":         uuid,
			"error":        err,
			"retry_count":  job.RetryCount,
			"max_retry":    q.maxRetry,
			"process_time": processTime,
		}).Error("Job processing failed")

		if job.RetryCount < q.maxRetry {
			// Retry the job
			atomic.AddInt64(&q.metrics.jobsRetried, 1)
			if err := q.retryJob(ctx, uuid, job); err != nil {
				q.logger.WithFields(logrus.Fields{
					"queue": q.name,
					"uuid":  uuid,
					"error": err,
				}).Error("Failed to retry job")
			}
		} else {
			// Move to DLQ or log permanent failure
			if err := q.moveToDLQ(ctx, uuid, job, err); err != nil {
				q.logger.WithFields(logrus.Fields{
					"queue": q.name,
					"uuid":  uuid,
					"error": err,
				}).Error("Failed to move job to DLQ")
			}
		}
	} else {
		atomic.AddInt64(&q.metrics.jobsProcessed, 1)
		q.logger.WithFields(logrus.Fields{
			"queue":        q.name,
			"uuid":         uuid,
			"process_time": processTime,
		}).Info("Job completed successfully")

		// Clean up data key after successful completion
		if err := q.redis.Del(ctx, dataKey).Err(); err != nil {
			q.logger.WithFields(logrus.Fields{
				"queue": q.name,
				"uuid":  uuid,
				"error": err,
			}).Warn("Failed to cleanup job data after successful completion")
		}
	}

	// Update metrics
	atomic.AddInt64(&q.metrics.totalProcessTime, int64(processTime))
	q.mu.Lock()
	q.metrics.lastProcessedAt = time.Now()
	q.mu.Unlock()

	return nil
}

// GetName returns the queue name
func (q *Queue) GetName() string {
	return q.name
}
