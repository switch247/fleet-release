package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
)

// ---------------------------------------------------------------------------
// Notifications: save, list, deduplication by fingerprint
// ---------------------------------------------------------------------------

func TestSaveAndListNotifications(t *testing.T) {
	st := newStore()
	n := models.Notification{
		ID:          "n1",
		UserID:      "u1",
		Title:       "Booking Confirmed",
		Body:        "Your booking is confirmed.",
		Status:      "pending",
		Fingerprint: "fp-booking-confirmed-b1",
	}
	st.SaveNotification(n)
	notifications := st.ListNotifications("u1")
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}
	if notifications[0].Title != "Booking Confirmed" {
		t.Fatalf("expected title 'Booking Confirmed', got %s", notifications[0].Title)
	}
}

func TestListNotifications_Empty(t *testing.T) {
	st := newStore()
	notifications := st.ListNotifications("no-user")
	if len(notifications) != 0 {
		t.Fatalf("expected 0 notifications, got %d", len(notifications))
	}
}

func TestNotificationDeduplication(t *testing.T) {
	// Saving a notification with the same fingerprint should update, not append
	st := newStore()
	fp := "fp-dedup-test"
	n1 := models.Notification{ID: "n1", UserID: "u1", Title: "v1", Status: "pending", Fingerprint: fp}
	n2 := models.Notification{ID: "n1", UserID: "u1", Title: "v2-updated", Status: "delivered", Fingerprint: fp}
	st.SaveNotification(n1)
	st.SaveNotification(n2)
	notifications := st.ListNotifications("u1")
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification after dedup, got %d", len(notifications))
	}
	if notifications[0].Title != "v2-updated" {
		t.Fatalf("expected updated title, got %s", notifications[0].Title)
	}
	if notifications[0].Status != "delivered" {
		t.Fatalf("expected status delivered, got %s", notifications[0].Status)
	}
}

func TestNotificationDeduplication_DifferentFingerprints(t *testing.T) {
	// Different fingerprints should produce separate notifications
	st := newStore()
	st.SaveNotification(models.Notification{ID: "n1", UserID: "u1", Fingerprint: "fp-a"})
	st.SaveNotification(models.Notification{ID: "n2", UserID: "u1", Fingerprint: "fp-b"})
	notifications := st.ListNotifications("u1")
	if len(notifications) != 2 {
		t.Fatalf("expected 2 distinct notifications, got %d", len(notifications))
	}
}

func TestListAllNotifications(t *testing.T) {
	st := newStore()
	st.SaveNotification(models.Notification{ID: "n1", UserID: "u1", Fingerprint: "fp1"})
	st.SaveNotification(models.Notification{ID: "n2", UserID: "u2", Fingerprint: "fp2"})
	all := st.ListAllNotifications()
	if len(all) != 2 {
		t.Fatalf("expected 2 total notifications, got %d", len(all))
	}
}

func TestListNotifications_IsolatedByUser(t *testing.T) {
	st := newStore()
	st.SaveNotification(models.Notification{ID: "n1", UserID: "u1", Fingerprint: "fp1"})
	st.SaveNotification(models.Notification{ID: "n2", UserID: "u2", Fingerprint: "fp2"})
	u1n := st.ListNotifications("u1")
	u2n := st.ListNotifications("u2")
	if len(u1n) != 1 || len(u2n) != 1 {
		t.Fatalf("expected notifications isolated by user, u1=%d u2=%d", len(u1n), len(u2n))
	}
}

// ---------------------------------------------------------------------------
// Notification Templates
// ---------------------------------------------------------------------------

func TestSaveAndGetNotificationTemplate(t *testing.T) {
	st := newStore()
	tmpl := models.NotificationTemplate{
		ID:         "t1",
		Name:       "booking_confirmed",
		Title:      "Booking Confirmed",
		Body:       "Your booking {{.BookingID}} is confirmed.",
		Channel:    "push",
		Enabled:    true,
		ModifiedAt: time.Now().UTC(),
	}
	st.SaveNotificationTemplate(tmpl)
	got, ok := st.GetNotificationTemplate("t1")
	if !ok {
		t.Fatal("expected template to be found")
	}
	if got.Name != "booking_confirmed" {
		t.Fatalf("expected name booking_confirmed, got %s", got.Name)
	}
}

func TestGetNotificationTemplate_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetNotificationTemplate("no-template")
	if ok {
		t.Fatal("expected false for missing template")
	}
}

func TestListNotificationTemplates(t *testing.T) {
	st := newStore()
	st.SaveNotificationTemplate(models.NotificationTemplate{ID: "t1", Name: "tmpl1"})
	st.SaveNotificationTemplate(models.NotificationTemplate{ID: "t2", Name: "tmpl2"})
	templates := st.ListNotificationTemplates()
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
}

func TestNotificationTemplate_Disabled(t *testing.T) {
	st := newStore()
	st.SaveNotificationTemplate(models.NotificationTemplate{ID: "t1", Name: "disabled-tmpl", Enabled: false})
	got, ok := st.GetNotificationTemplate("t1")
	if !ok {
		t.Fatal("expected template to be stored even when disabled")
	}
	if got.Enabled {
		t.Fatal("expected Enabled=false")
	}
}
