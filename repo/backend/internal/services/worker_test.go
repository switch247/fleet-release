package services

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"fleetlease/backend/internal/config"
	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/store"
)

func TestProcessNotificationRetries(t *testing.T) {
	st := store.NewMemoryStore()
	st.SaveNotification(models.Notification{ID: "n1", UserID: "u1", Fingerprint: "f1", Status: "queued", Attempts: 0})
	st.SaveNotification(models.Notification{ID: "n2", UserID: "u1", Fingerprint: "f2", Status: "failed", Attempts: 2})
	cfg := config.Config{NotificationRetryMax: 3}
	metrics := NewWorkerMetrics()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	processNotificationRetries(st, cfg, metrics, logger)

	items := st.ListNotifications("u1")
	var delivered, dead int
	for _, item := range items {
		if item.Status == "delivered" {
			delivered++
		}
		if item.Status == "dead_letter" {
			dead++
		}
	}
	if delivered == 0 || dead == 0 {
		t.Fatalf("expected delivered and dead letter outcomes, got delivered=%d dead=%d", delivered, dead)
	}
	snapshot := metrics.Snapshot()
	if snapshot.LastRunAt.IsZero() {
		t.Fatalf("expected metrics timestamp")
	}
}

func TestStartNotificationRetryWorker(t *testing.T) {
	st := store.NewMemoryStore()
	st.SaveNotification(models.Notification{ID: "n3", UserID: "u2", Fingerprint: "f3", Status: "queued"})
	cfg := config.Config{NotificationRetryMax: 3, NotificationRetryBackoffS: 1}
	metrics := NewWorkerMetrics()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	StartNotificationRetryWorker(st, cfg, logger, metrics)
	time.Sleep(1100 * time.Millisecond)
	if metrics.Snapshot().LastRunAt.IsZero() {
		t.Fatalf("expected worker to run at least once")
	}
}

func TestStartNotificationRetryWorkerWithNilMetrics(t *testing.T) {
	st := store.NewMemoryStore()
	cfg := config.Config{NotificationRetryMax: 2, NotificationRetryBackoffS: 1}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	StartNotificationRetryWorker(st, cfg, logger, nil)
}
