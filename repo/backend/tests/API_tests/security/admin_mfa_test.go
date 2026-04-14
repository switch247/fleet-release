package security

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestAdminSensitiveEndpointsRequireMFAWhenEnabled verifies that an admin user
// who has NOT enrolled TOTP is blocked (403) from sensitive admin routes when
// REQUIRE_ADMIN_MFA=true is set on the server (configured via docker-compose).
//
// Flow: login as sec-admin (TOTP enrolled) → create throwaway admin (no TOTP)
// → login as throwaway admin → try GET /api/v1/admin/users → expect 403.
func TestAdminSensitiveEndpointsRequireMFAWhenEnabled(t *testing.T) {
	skipIfNoSecLive(t)

	// Get admin token (sec-admin has TOTP enrolled and can create users).
	adminToken := secLoginAdmin(t)

	// Create a throwaway admin without TOTP enrolled.
	ts := time.Now().UnixNano()
	tmpAdmin := fmt.Sprintf("no-mfa-admin-%d", ts)
	tmpPass := "NoMfa1234!Pass"
	secCreateUser(t, adminToken, tmpAdmin, tmpPass, []string{"admin"})

	// Login as the throwaway admin — no TOTP code supplied (none enrolled).
	// The login itself should succeed (MFA enforcement is in the route middleware,
	// not the login handler).
	noMFAToken := secLogin(t, tmpAdmin, tmpPass)

	// Admin endpoint must return 403 because MFA is required but not enrolled.
	resp := secAPI(t, http.MethodGet, "/api/v1/admin/users", nil, noMFAToken)
	secMustStatus(t, resp, http.StatusForbidden)
}
