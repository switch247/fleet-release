package services

import (
	"log/slog"
	"sync"
	"time"

	"fleetlease/backend/internal/config"
	"fleetlease/backend/internal/store"
)

type WorkerMetrics struct {
	mu             sync.RWMutex
	LastRunAt      time.Time `json:"lastRunAt"`
	Processed      int       `json:"processed"`
	Delivered      int       `json:"delivered"`
	DeadLettered   int       `json:"deadLettered"`
	FailedRuns     int       `json:"failedRuns"`
	LastError      string    `json:"lastError"`
	CurrentBacklog int       `json:"currentBacklog"`
}

func NewWorkerMetrics() *WorkerMetrics {
	return &WorkerMetrics{}
}

func (m *WorkerMetrics) Snapshot() WorkerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m
}

func StartNotificationRetryWorker(st store.Repository, cfg config.Config, logger *slog.Logger, metrics *WorkerMetrics) {
	if metrics == nil {
		metrics = NewWorkerMetrics()
	}
	interval := time.Duration(cfg.NotificationRetryBackoffS) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			processNotificationRetries(st, cfg, metrics, logger)
		}
	}()
}

func processNotificationRetries(st store.Repository, cfg config.Config, metrics *WorkerMetrics, logger *slog.Logger) {
	all := st.ListAllNotifications()
	now := time.Now().UTC()
	backlog := 0
	processed := 0
	delivered := 0
	deadLettered := 0

	for _, notification := range all {
		switch notification.Status {
		case "queued", "failed", "retry_pending":
		default:
			continue
		}
		backlog++
		notification.Attempts++
		processed++
		if notification.Attempts >= cfg.NotificationRetryMax {
			notification.Status = "dead_letter"
			deadLettered++
			st.SaveNotification(notification)
			continue
		}
		notification.Status = "delivered"
		notification.DeliveredAt = now
		delivered++
		st.SaveNotification(notification)
	}

	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	metrics.LastRunAt = now
	metrics.Processed += processed
	metrics.Delivered += delivered
	metrics.DeadLettered += deadLettered
	metrics.CurrentBacklog = backlog
	metrics.LastError = ""
	if backlog > 0 {
		logger.Info("notification_retry_worker_run", "processed", processed, "delivered", delivered, "deadLettered", deadLettered)
	}
}
