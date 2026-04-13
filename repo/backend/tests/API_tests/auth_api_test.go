package api_tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestLoginSuccess(t *testing.T) {
	resp := apiCall(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": apiCustUser, "password": apiCustPass}, "")
	mustAPIStatus(t, resp, http.StatusOK)
}

func TestAdminResetPasswordRequiresIdentityEvidence(t *testing.T) {
	adminToken := liveLoginAdmin(t)

	// Missing evidence fields → 400
	resp := apiCall(t, http.MethodPost, "/api/v1/auth/admin-reset",
		map[string]string{"username": apiCustUser, "newPassword": "Customer4567!Pass"},
		adminToken)
	mustAPIStatus(t, resp, http.StatusBadRequest)

	// Throwaway customer so we don't mutate the shared api-customer password
	ts := time.Now().UnixNano()
	tmpUser := fmt.Sprintf("reset-target-%d", ts)
	tmpPass := "TmpTarget1234!"
	createTempUser(t, adminToken, tmpUser, tmpPass, []string{"customer"})

	// Valid reset with all evidence fields → 200
	resp2 := apiCall(t, http.MethodPost, "/api/v1/auth/admin-reset", map[string]string{
		"username":    tmpUser,
		"newPassword": "Target4567!Pass",
		"checkedBy":   apiAdminID,
		"method":      "government_id_match",
		"evidenceRef": "case-api-123",
		"reason":      "identity verified",
	}, adminToken)
	mustAPIStatus(t, resp2, http.StatusOK)
}
