package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestAdminSensitiveEndpointsRequireMFAWhenEnabled(t *testing.T) {
	t.Setenv("TEST_REQUIRE_ADMIN_MFA", "true")
	e := public.BuildSeededRouterForTests()

	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "Admin1234!Pass"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("admin login failed: %d %s", loginRec.Code, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)

	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	adminReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	adminRec := httptest.NewRecorder()
	e.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when admin MFA is required and not enabled, got %d", adminRec.Code)
	}
}
