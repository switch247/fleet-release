package security

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestAccountLockoutAfterFiveFailures creates a throwaway user so the shared
// sec-customer account is not left locked for subsequent tests.
func TestAccountLockoutAfterFiveFailures(t *testing.T) {
	skipIfNoSecLive(t)

	adminToken := secLoginAdmin(t)

	ts := time.Now().UnixNano()
	lockUser := fmt.Sprintf("lockout-%d", ts)
	lockPass := "Lockout1234!"
	secCreateUser(t, adminToken, lockUser, lockPass, []string{"customer"})

	for i := 0; i < 5; i++ {
		resp := secAPI(t, http.MethodPost, "/api/v1/auth/login",
			map[string]string{"username": lockUser, "password": "WrongPass999!"}, "")
		secMustStatus(t, resp, http.StatusUnauthorized)
	}

	// Correct credentials should now return 423 Locked.
	resp := secAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": lockUser, "password": lockPass}, "")
	secMustStatus(t, resp, http.StatusLocked)
}
