package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fleetlease/backend/pkg/public"

	"github.com/pquerna/otp/totp"
)

func TestTOTPFlow(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	loginBody, _ := json.Marshal(map[string]string{"username": "customer", "password": "Customer1234!"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed %d %s", loginRec.Code, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)

	enrollReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/enroll", nil)
	enrollReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	enrollRec := httptest.NewRecorder()
	e.ServeHTTP(enrollRec, enrollReq)
	if enrollRec.Code != http.StatusOK {
		t.Fatalf("enroll failed %d %s", enrollRec.Code, enrollRec.Body.String())
	}
	var enrollResp struct {
		Secret string `json:"secret"`
	}
	_ = json.Unmarshal(enrollRec.Body.Bytes(), &enrollResp)

	code, err := totp.GenerateCode(enrollResp.Secret, time.Now())
	if err != nil {
		t.Fatalf("failed generating totp: %v", err)
	}
	verifyBody, _ := json.Marshal(map[string]string{"code": code})
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	verifyRec := httptest.NewRecorder()
	e.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify failed %d %s", verifyRec.Code, verifyRec.Body.String())
	}

	noTotpLoginBody, _ := json.Marshal(map[string]string{"username": "customer", "password": "Customer1234!"})
	noTotpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(noTotpLoginBody))
	noTotpReq.Header.Set("Content-Type", "application/json")
	noTotpRec := httptest.NewRecorder()
	e.ServeHTTP(noTotpRec, noTotpReq)
	if noTotpRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without totp got %d", noTotpRec.Code)
	}

	withTotpCode, _ := totp.GenerateCode(enrollResp.Secret, time.Now())
	withTotpBody, _ := json.Marshal(map[string]string{"username": "customer", "password": "Customer1234!", "totpCode": withTotpCode})
	withTotpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(withTotpBody))
	withTotpReq.Header.Set("Content-Type", "application/json")
	withTotpRec := httptest.NewRecorder()
	e.ServeHTTP(withTotpRec, withTotpReq)
	if withTotpRec.Code != http.StatusOK {
		t.Fatalf("expected 200 with totp got %d body=%s", withTotpRec.Code, withTotpRec.Body.String())
	}
}
