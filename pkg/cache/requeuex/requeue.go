package delayqueue

import (
	"context"
	"errors"
	"fmt"
	"github.io/xhkzeroone/goframex/pkg/cache/redisx"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager manages multiple delay queues
type Manager struct {
	redis  *redisx.Redis
	queues []*Queue
	logger *logrus.Logger

	// Concurrency control
	mu sync.RWMutex

	// Shutdown control
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	isRunning  bool
	shutdownCh chan struct{}
}

// ManagerConfig holds configuration for the delay queue manager
type ManagerConfig struct {
	Logger *logrus.Logger
}

// NewManager creates a new delay queue manager
func NewManager(redis *redisx.Redis, cfg ManagerConfig) (*Manager, error) {
	if redis == nil {
		return nil, fmt.Errorf("redis backend cannot be nil")
	}

	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		redis:      redis,
		queues:     []*Queue{},
		logger:     cfg.Logger,
		ctx:        ctx,
		cancel:     cancel,
		shutdownCh: make(chan struct{}),
	}, nil
}

// Register adds a new queue to the manager
func (m *Manager) Register(cfg QueueConfig, handler JobHandler) (*Queue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return nil, fmt.Errorf("cannot add queue while manager is running")
	}

	queue, err := newQueue(cfg, m.redis, handler, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	m.queues = append(m.queues, queue)

	m.logger.WithFields(logrus.Fields{
		"queue_name": cfg.Name,
		"prefix":     cfg.KeyPrefix,
		"max_retry":  cfg.MaxRetry,
		"dlq":        cfg.DLQKey,
	}).Info("Queue added to manager")

	return queue, nil
}

// Start starts the delay queue manager
func (m *Manager) Start() error {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return errors.New("manager is already running")
	}
	m.isRunning = true
	m.mu.Unlock()

	// Create context for graceful shutdown
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Add worker to wait group
	m.wg.Add(1)

	// Start listener goroutine
	go func() {
		defer m.wg.Done()
		m.listen()
	}()

	// Scan and process expired jobs on startup
	go func() {
		if err := m.scanExpiredJobsOnStartup(); err != nil {
			m.logger.WithError(err).Error("Failed to scan expired jobs on startup")
		}
	}()

	m.logger.Info("Delay queue manager started")
	return nil
}

// Stop gracefully stops the manager
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if !m.isRunning {
		m.mu.Unlock()
		return nil
	}
	m.isRunning = false
	m.mu.Unlock()

	m.logger.Info("Stopping delay queue manager...")

	// Cancel context to stop the listener
	m.cancel()

	// Wait for listener to stop with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Delay queue manager stopped gracefully")
		return nil
	case <-ctx.Done():
		m.logger.Warn("Delay queue manager stop timed out")
		return ctx.Err()
	}
}

// listen listens for expired keys and processes them
func (m *Manager) listen() {
	// Enable keyspace notifications for expired events
	if err := m.enableKeyspaceNotifications(); err != nil {
		m.logger.WithError(err).Error("Failed to enable keyspace notifications")
		return
	}

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("Listener stopped due to context cancellation")
			return
		default:
			if err := m.listenWithPubSub(); err != nil {
				if m.ctx.Err() != nil {
					// Context was cancelled, exit gracefully
					return
				}
				m.logger.WithError(err).Warn("PubSub connection error, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}
}

// listenWithPubSub handles the actual PubSub listening with proper connection management
func (m *Manager) listenWithPubSub() error {
	pubsub := m.redis.PSubscribe(m.ctx, "__keyevent@0__:expired")
	defer func() {
		if err := pubsub.Close(); err != nil {
			m.logger.WithError(err).Error("Failed to close pub/sub connection")
		}
	}()

	m.logger.Info("Subscribed to Redis expired events")

	// Start a ping goroutine to keep the connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("PubSub listener stopped due to context cancellation")
			return nil
		case <-pingTicker.C:
			// Send a ping to keep the connection alive
			if err := m.redis.Ping(m.ctx).Err(); err != nil {
				m.logger.WithError(err).Warn("Redis ping failed, connection may be stale")
			}
		default:
			// Use a longer timeout for PubSub to avoid frequent timeouts
			// PubSub connections can be idle for long periods
			ctx, cancel := context.WithTimeout(m.ctx, 60*time.Second)
			msg, err := pubsub.ReceiveMessage(ctx)
			cancel()

			if err != nil {
				if m.ctx.Err() != nil {
					// Context was cancelled, exit gracefully
					return nil
				}

				// Check if it's a timeout error
				if err.Error() == "context deadline exceeded" {
					// This is expected for PubSub when no messages are received
					// Just continue the loop
					continue
				}

				return fmt.Errorf("pubsub receive error: %w", err)
			}

			key := msg.Payload
			m.logger.WithField("expired_key", key).Debug("Received expired key event")

			// Process the expired key in all queues
			m.processExpiredKey(key)
		}
	}
}

// processExpiredKey processes an expired key across all queues
func (m *Manager) processExpiredKey(key string) {
	m.mu.RLock()
	queues := make([]*Queue, len(m.queues))
	copy(queues, m.queues)
	m.mu.RUnlock()

	for _, queue := range queues {
		// Process each queue in a separate goroutine to avoid blocking
		go func(q *Queue) {
			defer func() {
				if r := recover(); r != nil {
					m.logger.WithFields(logrus.Fields{
						"queue": q.GetName(),
						"panic": r,
					}).Error("Panic in queue processing")
				}
			}()

			ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
			defer cancel()

			q.handleExpiredKey(ctx, key)
		}(queue)
	}
}

// enableKeyspaceNotifications enables Redis keyspace notifications
func (m *Manager) enableKeyspaceNotifications() error {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	// First, check current configuration
	currentConfig, err := m.redis.ConfigGet(ctx, "notify-keyspace-events").Result()
	if err != nil {
		return fmt.Errorf("failed to get current keyspace notification config: %w", err)
	}

	m.logger.WithField("current_config", currentConfig).Info("Current Redis keyspace notification config")

	// Set notify-keyspace-events to enable expired event notifications
	// 'E' for keyspace events, 'x' for expired events
	result := m.redis.ConfigSet(ctx, "notify-keyspace-events", "Ex")
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to enable keyspace notifications: %w", err)
	}

	// Verify the configuration was set
	newConfig, err := m.redis.ConfigGet(ctx, "notify-keyspace-events").Result()
	if err != nil {
		return fmt.Errorf("failed to verify keyspace notification config: %w", err)
	}

	m.logger.WithField("new_config", newConfig).Info("Redis keyspace notification config updated")

	return nil
}

// GetStats returns statistics for all queues
func (m *Manager) GetStats() map[string]QueueStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]QueueStats)
	for _, queue := range m.queues {
		stats[queue.GetName()] = queue.GetStats()
	}

	return stats
}

// GetQueue returns a queue by name
func (m *Manager) GetQueue(name string) *Queue {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, queue := range m.queues {
		if queue.GetName() == name {
			return queue
		}
	}

	return nil
}

// scanExpiredJobsOnStartup scans for expired jobs that were missed during downtime
func (m *Manager) scanExpiredJobsOnStartup() error {
	m.logger.Info("Starting scan for expired jobs on startup...")

	m.mu.RLock()
	queues := make([]*Queue, len(m.queues))
	copy(queues, m.queues)
	m.mu.RUnlock()

	for _, queue := range queues {
		if err := m.scanExpiredJobsForQueue(queue); err != nil {
			m.logger.WithFields(logrus.Fields{
				"queue": queue.GetName(),
				"error": err,
			}).Error("Failed to scan expired jobs for queue")
		}
	}

	m.logger.Info("Completed scan for expired jobs on startup")
	return nil
}

// scanExpiredJobsForQueue scans for expired jobs in a specific queue
func (m *Manager) scanExpiredJobsForQueue(queue *Queue) error {
	// Scan for data keys that might have orphaned trigger keys
	dataPattern := queue.prefix + ":data:*"

	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	// Get all data keys for this queue
	dataKeys, err := m.redis.Keys(ctx, dataPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to scan data keys for pattern %s: %w", dataPattern, err)
	}

	if len(dataKeys) == 0 {
		m.logger.WithField("queue", queue.GetName()).Debug("No data keys found for queue")
		return nil
	}

	m.logger.WithFields(logrus.Fields{
		"queue":    queue.GetName(),
		"dataKeys": len(dataKeys),
	}).Info("Found data keys during startup scan, checking for orphaned trigger keys")

	// Check each data key for orphaned trigger keys
	processedCount := 0
	for _, dataKey := range dataKeys {
		parts := strings.SplitN(dataKey, ":", 3)
		if len(parts) != 3 || parts[1] != "data" {
			m.logger.WithField("dataKey", dataKey).Warn("Invalid data key format during scan")
			continue
		}

		uuid := parts[2]
		triggerKey := queue.prefix + ":trigger:" + uuid

		// Check if the corresponding trigger key exists
		exists, err := m.redis.Exists(ctx, triggerKey).Result()
		if err != nil {
			m.logger.WithFields(logrus.Fields{
				"queue":      queue.GetName(),
				"triggerKey": triggerKey,
				"error":      err,
			}).Error("Failed to check trigger key existence")
			continue
		}

		if exists == 0 {
			// Trigger key doesn't exist, meaning the job has expired
			// Process the job using the data key
			m.logger.WithFields(logrus.Fields{
				"queue":      queue.GetName(),
				"uuid":       uuid,
				"dataKey":    dataKey,
				"triggerKey": triggerKey,
			}).Info("Found orphaned data key during startup scan, processing job...")

			// Process the job by reading from data key
			if err := queue.processJobFromDataKey(ctx, dataKey, uuid); err != nil {
				m.logger.WithFields(logrus.Fields{
					"queue":   queue.GetName(),
					"uuid":    uuid,
					"dataKey": dataKey,
					"error":   err,
				}).Error("Failed to process job from data key")
			} else {
				processedCount++
			}
		} else {
			// Trigger key exists, check if it's about to expire
			ttl, err := m.redis.TTL(ctx, triggerKey).Result()
			if err != nil {
				m.logger.WithFields(logrus.Fields{
					"queue":      queue.GetName(),
					"triggerKey": triggerKey,
					"error":      err,
				}).Error("Failed to check trigger key TTL")
				continue
			}

			if ttl <= 0 {
				// Trigger key has expired or is about to expire
				m.logger.WithFields(logrus.Fields{
					"queue":      queue.GetName(),
					"uuid":       uuid,
					"triggerKey": triggerKey,
					"ttl":        ttl,
				}).Info("Found expired trigger key during startup scan, processing job...")

				// Process the job
				queue.handleExpiredKey(ctx, triggerKey)
				processedCount++
			}
		}
	}

	if processedCount > 0 {
		m.logger.WithFields(logrus.Fields{
			"queue":          queue.GetName(),
			"processedCount": processedCount,
		}).Info("Processed expired jobs during startup scan")
	}

	return nil
}

// ScanExpiredJobs manually triggers a scan for expired jobs
func (m *Manager) ScanExpiredJobs() error {
	if !m.IsRunning() {
		return errors.New("manager is not running")
	}

	m.logger.Info("Manual scan for expired jobs triggered")
	return m.scanExpiredJobsOnStartup()
}

// IsRunning returns whether the manager is currently running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}
