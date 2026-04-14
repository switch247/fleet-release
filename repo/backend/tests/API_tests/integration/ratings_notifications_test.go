package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRatingsAndNotificationRetryFlow(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	adminToken := intLoginAdmin(t)

	// Create a fresh booking for this test so we can rate it independently.
	bID := intCreateBooking(t, custToken)

	// Customer submits a rating.
	resp := intAPI(t, http.MethodPost, "/api/v1/ratings",
		map[string]interface{}{"bookingId": bID, "score": 5, "comment": "great handover"},
		custToken)
	intMustStatus(t, resp, http.StatusCreated)

	// List ratings for the booking.
	resp2 := intAPI(t, http.MethodGet, "/api/v1/ratings?bookingId="+bID, nil, custToken)
	intMustStatus(t, resp2, http.StatusOK)

	// Find int-customer ID via admin user list.
	usersResp := intAPI(t, http.MethodGet, "/api/v1/admin/users", nil, adminToken)
	usersBody := intMustStatus(t, usersResp, http.StatusOK)
	var users []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	_ = json.Unmarshal(usersBody, &users)
	custID := ""
	for _, u := range users {
		if u.Username == intCustUser {
			custID = u.ID
		}
	}
	if custID == "" {
		t.Fatalf("int-customer not found in user list")
	}

	// Create notification template.
	tmplResp := intAPI(t, http.MethodPost, "/api/v1/admin/notification-templates",
		map[string]interface{}{
			"name": "Email Notice", "title": "Email Template",
			"body": "hello", "channel": "email", "enabled": true,
		}, adminToken)
	tmplBody := intMustStatus(t, tmplResp, http.StatusCreated)
	var tmpl struct{ ID string `json:"id"` }
	_ = json.Unmarshal(tmplBody, &tmpl)

	// Send notification.
	sendResp := intAPI(t, http.MethodPost, "/api/v1/admin/notifications/send",
		map[string]string{"templateId": tmpl.ID, "userId": custID}, adminToken)
	intMustStatus(t, sendResp, http.StatusCreated)

	// Retry notifications.
	retryResp := intAPI(t, http.MethodPost, "/api/v1/admin/notifications/retry", nil, adminToken)
	intMustStatus(t, retryResp, http.StatusOK)

	// Customer sees their notifications.
	notifResp := intAPI(t, http.MethodGet, "/api/v1/notifications", nil, custToken)
	notifBody := intMustStatus(t, notifResp, http.StatusOK)
	var notifs []struct {
		Title    string `json:"title"`
		Status   string `json:"status"`
		Attempts int    `json:"attempts"`
	}
	_ = json.Unmarshal(notifBody, &notifs)
	if len(notifs) == 0 {
		t.Fatal("expected notification records")
	}
	found := false
	for _, n := range notifs {
		if n.Title == "Email Template" {
			found = true
			if n.Status != "disabled_offline" {
				t.Fatalf("expected disabled_offline status, got %s", n.Status)
			}
			if n.Attempts < 2 {
				t.Fatalf("expected attempts >= 2 after retry, got %d", n.Attempts)
			}
		}
	}
	if !found {
		t.Fatal("Email Template notification not found")
	}
}
