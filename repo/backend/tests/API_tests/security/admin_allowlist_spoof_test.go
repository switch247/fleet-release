package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestAdminAllowlistIgnoresUntrustedForwardedHeaders(t *testing.T) {
	e := public.BuildSeededRouterForTests()

	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "Admin1234!Pass"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.RemoteAddr = "192.0.2.10:4321"
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("admin login failed: %d %s", loginRec.Code, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	if loginResp.Token == "" {
		t.Fatalf("missing token")
	}

	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	adminReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	adminReq.Header.Set("X-Forwarded-For", "127.0.0.1")
	adminReq.RemoteAddr = "203.0.113.11:5555"
	adminRec := httptest.NewRecorder()
	e.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for disallowed remote addr with spoofed X-Forwarded-For, got %d body=%s", adminRec.Code, adminRec.Body.String())
	}
}
