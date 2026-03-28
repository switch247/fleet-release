package api_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func loginToken(t *testing.T, e http.Handler, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed for %s status=%d body=%s", username, rec.Code, rec.Body.String())
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid login response: %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("missing token for %s", username)
	}
	return payload.Token
}

func TestProviderCannotSettleUnownedBooking(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	adminToken := loginToken(t, e, "admin", "Admin1234!Pass")

	createBody, _ := json.Marshal(map[string]interface{}{
		"username": "provider_two",
		"password": "ProviderTwo1234!",
		"roles":    []string{"provider"},
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+adminToken)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create user failed status=%d body=%s", createRec.Code, createRec.Body.String())
	}

	provider2Token := loginToken(t, e, "provider_two", "ProviderTwo1234!")
	settleReq := httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/22222222-2222-2222-2222-222222222222", nil)
	settleReq.Header.Set("Authorization", "Bearer "+provider2Token)
	settleRec := httptest.NewRecorder()
	e.ServeHTTP(settleRec, settleReq)
	if settleRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d body=%s", settleRec.Code, settleRec.Body.String())
	}
}

func TestComplaintArbitrationRequiresCSAOrAdmin(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	customerToken := loginToken(t, e, "customer", "Customer1234!")

	complaintBody, _ := json.Marshal(map[string]string{"bookingId": "22222222-2222-2222-2222-222222222222", "outcome": "broken mirror"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/complaints", bytes.NewReader(complaintBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+customerToken)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create complaint failed status=%d body=%s", createRec.Code, createRec.Body.String())
	}

	var complaint struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &complaint); err != nil {
		t.Fatalf("invalid complaint response: %v", err)
	}
	arbBody, _ := json.Marshal(map[string]string{"status": "closed", "outcome": "denied"})
	arbReq := httptest.NewRequest(http.MethodPatch, "/api/v1/complaints/"+complaint.ID+"/arbitrate", bytes.NewReader(arbBody))
	arbReq.Header.Set("Content-Type", "application/json")
	arbReq.Header.Set("Authorization", "Bearer "+customerToken)
	arbRec := httptest.NewRecorder()
	e.ServeHTTP(arbRec, arbReq)
	if arbRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d body=%s", arbRec.Code, arbRec.Body.String())
	}
}
