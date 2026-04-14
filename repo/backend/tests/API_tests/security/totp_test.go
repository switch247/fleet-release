package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

// TestTOTPFlow exercises enroll → verify → login-without-code → login-with-code.
// A throwaway user is created so the shared sec-customer is not mutated.
func TestTOTPFlow(t *testing.T) {
	skipIfNoSecLive(t)

	adminToken := secLoginAdmin(t)

	ts := time.Now().UnixNano()
	totpUser := fmt.Sprintf("totp-user-%d", ts)
	totpPass := "TotpUser1234!"
	secCreateUser(t, adminToken, totpUser, totpPass, []string{"customer"})

	// 1. Login to get a token (TOTP not yet enrolled, no code needed).
	token := secLogin(t, totpUser, totpPass)

	// 2. Enroll TOTP — server returns the secret.
	enrollResp := secAPI(t, http.MethodPost, "/api/v1/auth/totp/enroll", nil, token)
	enrollBody := secMustStatus(t, enrollResp, http.StatusOK)
	var enroll struct{ Secret string `json:"secret"` }
	if err := json.Unmarshal(enrollBody, &enroll); err != nil || enroll.Secret == "" {
		t.Fatalf("enroll: missing secret in %s", enrollBody)
	}

	// 3. Verify with a fresh code to activate TOTP.
	code, err := totp.GenerateCode(enroll.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	verifyResp := secAPI(t, http.MethodPost, "/api/v1/auth/totp/verify",
		map[string]string{"code": code}, token)
	secMustStatus(t, verifyResp, http.StatusOK)

	// 4. Login without TOTP code → 401 (TOTP now required).
	noCodeResp := secAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": totpUser, "password": totpPass}, "")
	secMustStatus(t, noCodeResp, http.StatusUnauthorized)

	// 5. Login with a fresh TOTP code → 200.
	code2, _ := totp.GenerateCode(enroll.Secret, time.Now().UTC())
	withCodeResp := secAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": totpUser, "password": totpPass, "totpCode": code2}, "")
	secMustStatus(t, withCodeResp, http.StatusOK)
}
