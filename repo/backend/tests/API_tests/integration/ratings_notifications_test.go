package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestRatingsAndNotificationRetryFlow(t *testing.T) {
	h := public.BuildHarnessForTests()
	customerToken := loginToken(t, h.Router, "customer", "Customer1234!")
	adminToken := loginToken(t, h.Router, "admin", "Admin1234!Pass")

	ratingBody, _ := json.Marshal(map[string]interface{}{"bookingId": h.BookingID, "score": 5, "comment": "great handover"})
	ratingReq := httptest.NewRequest(http.MethodPost, "/api/v1/ratings", bytes.NewReader(ratingBody))
	ratingReq.Header.Set("Content-Type", "application/json")
	ratingReq.Header.Set("Authorization", "Bearer "+customerToken)
	ratingRec := httptest.NewRecorder()
	h.Router.ServeHTTP(ratingRec, ratingReq)
	if ratingRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 rating got %d body=%s", ratingRec.Code, ratingRec.Body.String())
	}

	listRatingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/ratings?bookingId="+h.BookingID, nil)
	listRatingsReq.Header.Set("Authorization", "Bearer "+customerToken)
	listRatingsRec := httptest.NewRecorder()
	h.Router.ServeHTTP(listRatingsRec, listRatingsReq)
	if listRatingsRec.Code != http.StatusOK {
		t.Fatalf("expected 200 ratings list got %d", listRatingsRec.Code)
	}

	usersReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	usersReq.Header.Set("Authorization", "Bearer "+adminToken)
	usersRec := httptest.NewRecorder()
	h.Router.ServeHTTP(usersRec, usersReq)
	if usersRec.Code != http.StatusOK {
		t.Fatalf("users list failed: %d %s", usersRec.Code, usersRec.Body.String())
	}
	var users []struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}
	_ = json.Unmarshal(usersRec.Body.Bytes(), &users)
	customerID := ""
	for _, u := range users {
		if u.Username == "customer" {
			customerID = u.ID
		}
	}
	if customerID == "" {
		t.Fatalf("customer user not found")
	}

	templateBody, _ := json.Marshal(map[string]interface{}{"name": "Email Notice", "title": "Email Template", "body": "hello", "channel": "email", "enabled": true})
	templateReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notification-templates", bytes.NewReader(templateBody))
	templateReq.Header.Set("Content-Type", "application/json")
	templateReq.Header.Set("Authorization", "Bearer "+adminToken)
	templateRec := httptest.NewRecorder()
	h.Router.ServeHTTP(templateRec, templateReq)
	if templateRec.Code != http.StatusCreated {
		t.Fatalf("template create failed: %d %s", templateRec.Code, templateRec.Body.String())
	}
	var templateResp struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(templateRec.Body.Bytes(), &templateResp)

	sendBody, _ := json.Marshal(map[string]string{"templateId": templateResp.ID, "userId": customerID})
	sendReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications/send", bytes.NewReader(sendBody))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+adminToken)
	sendRec := httptest.NewRecorder()
	h.Router.ServeHTTP(sendRec, sendReq)
	if sendRec.Code != http.StatusCreated {
		t.Fatalf("send notification failed: %d %s", sendRec.Code, sendRec.Body.String())
	}

	retryReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications/retry", nil)
	retryReq.Header.Set("Authorization", "Bearer "+adminToken)
	retryRec := httptest.NewRecorder()
	h.Router.ServeHTTP(retryRec, retryReq)
	if retryRec.Code != http.StatusOK {
		t.Fatalf("retry failed: %d %s", retryRec.Code, retryRec.Body.String())
	}

	listNotifReq := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	listNotifReq.Header.Set("Authorization", "Bearer "+customerToken)
	listNotifRec := httptest.NewRecorder()
	h.Router.ServeHTTP(listNotifRec, listNotifReq)
	if listNotifRec.Code != http.StatusOK {
		t.Fatalf("notifications list failed: %d", listNotifRec.Code)
	}
	var notifs []struct {
		Title    string `json:"title"`
		Status   string `json:"status"`
		Attempts int    `json:"attempts"`
	}
	_ = json.Unmarshal(listNotifRec.Body.Bytes(), &notifs)
	if len(notifs) == 0 {
		t.Fatalf("expected notification records")
	}
	found := false
	for _, n := range notifs {
		if n.Title == "Email Template" {
			found = true
			if n.Status != "disabled_offline" {
				t.Fatalf("expected disabled_offline status for email template, got %s", n.Status)
			}
			if n.Attempts < 2 {
				t.Fatalf("expected attempts increment after retry, got %d", n.Attempts)
			}
		}
	}
	if !found {
		t.Fatalf("expected to find Email Template notification")
	}
}
