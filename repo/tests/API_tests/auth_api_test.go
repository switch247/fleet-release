package api_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestLoginSuccess(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	body, _ := json.Marshal(map[string]string{"username": "customer", "password": "Customer1234!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminResetPasswordRequiresIdentityEvidence(t *testing.T) {
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
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	_ = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)

	missingEvidenceBody, _ := json.Marshal(map[string]string{"username": "customer", "newPassword": "Customer4567!Pass"})
	missingEvidenceReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin-reset", bytes.NewReader(missingEvidenceBody))
	missingEvidenceReq.Header.Set("Content-Type", "application/json")
	missingEvidenceReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	missingEvidenceRec := httptest.NewRecorder()
	e.ServeHTTP(missingEvidenceRec, missingEvidenceReq)
	if missingEvidenceRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing evidence, got %d", missingEvidenceRec.Code)
	}

	validBody, _ := json.Marshal(map[string]string{
		"username":    "customer",
		"newPassword": "Customer4567!Pass",
		"checkedBy":   loginResp.User.ID,
		"method":      "government_id_match",
		"evidenceRef": "case-123",
		"reason":      "identity verified at desk",
	})
	validReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin-reset", bytes.NewReader(validBody))
	validReq.Header.Set("Content-Type", "application/json")
	validReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	validRec := httptest.NewRecorder()
	e.ServeHTTP(validRec, validReq)
	if validRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid reset, got %d body=%s", validRec.Code, validRec.Body.String())
	}
}
