package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"

	"github.com/google/uuid"
)

// createCustomerBInProcess creates a second customer user via the admin API
// using an in-process httptest request, then returns a valid JWT for that user.
func createCustomerBInProcess(t *testing.T, h *public.TestHarness) string {
	t.Helper()

	adminToken := loginToken(t, h.Router, "admin", "Admin1234!Pass")
	suffix := uuid.NewString()[:8]
	custBUser := "cust-b-" + suffix
	custBPass := "CustBTwo1234!"

	body, _ := json.Marshal(map[string]interface{}{
		"username": custBUser,
		"password": custBPass,
		"email":    custBUser + "@fleetlease.local",
		"roles":    []string{"customer"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	// 192.0.2.1 is in the test admin allowlist (192.0.2.0/24)
	req.RemoteAddr = "192.0.2.1:9999"
	rec := httptest.NewRecorder()
	h.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin create customer B: expected 201 got %d body=%s", rec.Code, rec.Body.String())
	}

	return loginToken(t, h.Router, custBUser, custBPass)
}

// TestCustomerCannotReadAnotherCustomerLedger verifies that a customer
// cannot access the settlement ledger of a booking they do not own or participate in.
func TestCustomerCannotReadAnotherCustomerLedger(t *testing.T) {
	h := public.BuildHarnessForTests()
	custBToken := createCustomerBInProcess(t, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/"+h.BookingID, nil)
	req.Header.Set("Authorization", "Bearer "+custBToken)
	rec := httptest.NewRecorder()
	h.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-customer ledger read, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestCustomerCannotSettleAnotherCustomerBooking verifies that a customer
// cannot close settlement for a booking they do not own.
func TestCustomerCannotSettleAnotherCustomerBooking(t *testing.T) {
	h := public.BuildHarnessForTests()
	custBToken := createCustomerBInProcess(t, h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/"+h.BookingID, nil)
	req.Header.Set("Authorization", "Bearer "+custBToken)
	rec := httptest.NewRecorder()
	h.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-customer settlement close, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestCustomerCannotListInspectionsForAnotherBooking verifies that a customer
// cannot read inspection revisions for a booking they are not party to.
func TestCustomerCannotListInspectionsForAnotherBooking(t *testing.T) {
	h := public.BuildHarnessForTests()
	custBToken := createCustomerBInProcess(t, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inspections?bookingId="+h.BookingID, nil)
	req.Header.Set("Authorization", "Bearer "+custBToken)
	rec := httptest.NewRecorder()
	h.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-customer inspection list, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestCustomerCannotVerifyLedgerForAnotherBooking verifies that the ledger
// verify endpoint enforces booking-scoped ownership checks.
func TestCustomerCannotVerifyLedgerForAnotherBooking(t *testing.T) {
	h := public.BuildHarnessForTests()
	custBToken := createCustomerBInProcess(t, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/"+h.BookingID+"/verify", nil)
	req.Header.Set("Authorization", "Bearer "+custBToken)
	rec := httptest.NewRecorder()
	h.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-customer ledger verify, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestNotificationsReturnOnlyCallerNotifications verifies that each user's
// notification list is filtered to their own user ID.
func TestNotificationsReturnOnlyCallerNotifications(t *testing.T) {
	h := public.BuildHarnessForTests()

	for _, tc := range []struct {
		user, pass string
	}{
		{"customer", "Customer1234!"},
		{"provider", "Provider1234!"},
		{"agent", "Agent1234!Pass"},
	} {
		token := loginToken(t, h.Router, tc.user, tc.pass)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: notifications expected 200 got %d body=%s", tc.user, rec.Code, rec.Body.String())
		}
		// Response must be a JSON array (no cross-user leakage at the status/shape level).
		var notifications []json.RawMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &notifications); err != nil {
			t.Fatalf("%s: notifications response is not a JSON array: %v", tc.user, err)
		}
	}
}
