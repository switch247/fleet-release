package integration

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
		t.Fatalf("login failed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid login response: %v", err)
	}
	return payload.Token
}

func TestSettlementHashChainDetectsTampering(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")

	closeReq := httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/"+h.BookingID, nil)
	closeReq.Header.Set("Authorization", "Bearer "+token)
	closeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(closeRec, closeReq)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("close settlement expected 200 got %d body=%s", closeRec.Code, closeRec.Body.String())
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/"+h.BookingID+"/verify", nil)
	verifyReq.Header.Set("Authorization", "Bearer "+token)
	verifyRec := httptest.NewRecorder()
	h.Router.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify expected 200 got %d body=%s", verifyRec.Code, verifyRec.Body.String())
	}
	var before struct {
		Valid bool `json:"valid"`
	}
	_ = json.Unmarshal(verifyRec.Body.Bytes(), &before)
	if !before.Valid {
		t.Fatalf("expected chain to be valid before tamper")
	}

	h.TamperLedger(h.BookingID)

	afterReq := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/"+h.BookingID+"/verify", nil)
	afterReq.Header.Set("Authorization", "Bearer "+token)
	afterRec := httptest.NewRecorder()
	h.Router.ServeHTTP(afterRec, afterReq)
	if afterRec.Code != http.StatusOK {
		t.Fatalf("verify after tamper expected 200 got %d body=%s", afterRec.Code, afterRec.Body.String())
	}
	var after struct {
		Valid bool `json:"valid"`
	}
	_ = json.Unmarshal(afterRec.Body.Bytes(), &after)
	if after.Valid {
		t.Fatalf("expected chain to be invalid after tampering")
	}
}
