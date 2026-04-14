// live_setup_test.go provides shared live-HTTP infrastructure for the security
// package tests that have been converted to make real network calls.
// transport_test.go and admin_allowlist_spoof_test.go remain in-process because
// they inject RemoteAddr values that are impossible to forge over a real TCP
// connection.
package security

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var (
	secLiveServerURL string
	secLiveClient    *http.Client
	secAdminTOTP     string
)

const (
	secAdminUser  = "sec-admin"
	secAdminPass  = "SecAdmin1234!"
	secCustUser   = "sec-customer"
	secCustPass   = "SecCust1234!"

	secAdminID = "5555aaaa-0000-0000-0000-000000000001"
	secCustID  = "5555aaaa-0000-0000-0000-000000000002"
)

// TestMain runs in-process tests unconditionally and enables live-HTTP tests
// when TEST_SERVER_URL is set.  Tests that require live access call
// skipIfNoSecLive(t) at the top.
func TestMain(m *testing.M) {
	if url := os.Getenv("TEST_SERVER_URL"); url != "" {
		secLiveServerURL = url
		secLiveClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
			Timeout: 30 * time.Second,
		}
		secAdminTOTP = os.Getenv("TEST_ADMIN_TOTP_SECRET")
		if secAdminTOTP == "" {
			fmt.Fprintln(os.Stderr, "WARN: TEST_ADMIN_TOTP_SECRET not set — security live tests will skip")
			secLiveServerURL = ""
		} else if err := seedSecTestUsers(); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: security live seed failed: %v — live tests will skip\n", err)
			secLiveServerURL = ""
		}
	}
	os.Exit(m.Run())
}

func skipIfNoSecLive(t *testing.T) {
	t.Helper()
	if secLiveServerURL == "" {
		t.Skip("TEST_SERVER_URL / TEST_ADMIN_TOTP_SECRET not set — skipping live security test")
	}
}

func seedSecTestUsers() error {
	dsn := secDSN()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("postgres unavailable: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	defer pool.Close()

	var mig []byte
	for _, p := range []string{
		"../../../migrations/001_init.sql",
		"../../migrations/001_init.sql",
		"migrations/001_init.sql",
		"backend/migrations/001_init.sql",
	} {
		mig, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if mig == nil {
		if abs, e := filepath.Abs("../../../migrations/001_init.sql"); e == nil {
			mig, _ = os.ReadFile(abs)
		}
	}
	if mig == nil {
		return fmt.Errorf("migration not found")
	}
	if _, err = pool.Exec(ctx, string(mig)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	adminHash, _ := bcrypt.GenerateFromPassword([]byte(secAdminPass), bcrypt.MinCost)
	custHash, _ := bcrypt.GenerateFromPassword([]byte(secCustPass), bcrypt.MinCost)

	for _, u := range []struct {
		id, username, email, hash string
		totpEnabled                bool
		totpSecret                 string
	}{
		{secAdminID, secAdminUser, secAdminUser + "@test.local", string(adminHash), true, secAdminTOTP},
		{secCustID, secCustUser, secCustUser + "@test.local", string(custHash), false, ""},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id,username,email,password_hash,totp_enabled,totp_secret)
			VALUES ($1,$2,$3,$4,$5,NULLIF($6,''))
			ON CONFLICT (username) DO UPDATE SET
				password_hash=EXCLUDED.password_hash,
				totp_enabled=EXCLUDED.totp_enabled,
				totp_secret=EXCLUDED.totp_secret`,
			u.id, u.username, u.email, u.hash, u.totpEnabled, u.totpSecret); err != nil {
			return fmt.Errorf("seed user %s: %w", u.username, err)
		}
	}
	for _, r := range []struct{ uid, role string }{
		{secAdminID, "admin"}, {secCustID, "customer"},
	} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO user_roles (user_id,role) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			r.uid, r.role); err != nil {
			return fmt.Errorf("seed role: %w", err)
		}
	}
	return nil
}

func secDSN() string {
	if u := os.Getenv("TEST_DATABASE_URL"); u != "" {
		return u
	}
	get := func(k, d string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return d
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		get("DB_USER", "fleetlease"), get("DB_PASSWORD", "fleetlease"),
		get("DB_HOST", "w1-t1-ti1-db"), get("DB_PORT", "5432"),
		get("DB_NAME", "fleetlease"), get("DB_SSL_MODE", "disable"))
}

// secAPI sends a real HTTP request to the running server.
func secAPI(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, secLiveServerURL+path, r)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := secLiveClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

func secReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b
}

func secMustStatus(t *testing.T, resp *http.Response, want int) []byte {
	t.Helper()
	b := secReadBody(t, resp)
	if resp.StatusCode != want {
		t.Fatalf("expected HTTP %d, got %d — body: %s", want, resp.StatusCode, b)
	}
	return b
}

func secLogin(t *testing.T, username, password string) string {
	t.Helper()
	resp := secAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": username, "password": password}, "")
	b := secMustStatus(t, resp, http.StatusOK)
	var out struct{ Token string `json:"token"` }
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("secLogin %s: no token in %s", username, b)
	}
	return out.Token
}

func secLoginAdmin(t *testing.T) string {
	t.Helper()
	code, err := totp.GenerateCode(secAdminTOTP, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate totp: %v", err)
	}
	resp := secAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": secAdminUser, "password": secAdminPass, "totpCode": code}, "")
	b := secMustStatus(t, resp, http.StatusOK)
	var out struct{ Token string `json:"token"` }
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("secLoginAdmin: no token in %s", b)
	}
	return out.Token
}

func secCreateUser(t *testing.T, adminToken, username, password string, roles []string) string {
	t.Helper()
	resp := secAPI(t, http.MethodPost, "/api/v1/admin/users", map[string]interface{}{
		"username": username, "email": username + "@test.local",
		"password": password, "roles": roles,
	}, adminToken)
	b := secMustStatus(t, resp, http.StatusCreated)
	var out struct{ ID string `json:"id"` }
	if err := json.Unmarshal(b, &out); err != nil || out.ID == "" {
		t.Fatalf("secCreateUser: bad response %s", b)
	}
	return out.ID
}
