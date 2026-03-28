package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestTransportRejectsNonWhitelistedHTTP(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-whitelisted HTTP got %d", rec.Code)
	}
}
