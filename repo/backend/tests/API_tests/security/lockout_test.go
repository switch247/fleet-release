package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestAccountLockoutAfterFiveFailures(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	for i := 0; i < 5; i++ {
		body, _ := json.Marshal(map[string]string{"username": "customer", "password": "WrongPass123!"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected 401 got %d body=%s", i+1, rec.Code, rec.Body.String())
		}
	}

	validBody, _ := json.Marshal(map[string]string{"username": "customer", "password": "Customer1234!"})
	validReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(validBody))
	validReq.Header.Set("Content-Type", "application/json")
	validRec := httptest.NewRecorder()
	e.ServeHTTP(validRec, validReq)
	if validRec.Code != http.StatusLocked {
		t.Fatalf("expected 423 lockout got %d body=%s", validRec.Code, validRec.Body.String())
	}
}
