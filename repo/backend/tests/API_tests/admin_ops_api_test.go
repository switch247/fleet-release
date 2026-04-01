package api_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestBackupRestoreDegradedWhenScriptsMissing(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	adminToken := loginForEndpoint(t, e, "admin", "Admin1234!Pass")

	tests := []string{
		"/api/v1/admin/backup/now",
		"/api/v1/admin/restore/now",
	}
	for _, path := range tests {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 for %s got %d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}
