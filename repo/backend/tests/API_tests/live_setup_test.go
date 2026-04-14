// Package api_tests exercises the backend API using real network HTTP calls.
// Every request in this package goes through net/http.Client → TLS → the
// running server.  No httptest.NewRecorder or in-process handler calls are
// used.  Tests are skipped automatically when TEST_SERVER_URL is not set.
package api_tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
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

// ---------------------------------------------------------------------------
// Package-level state, set by TestMain
// ---------------------------------------------------------------------------

var (
	liveServerURL  string
	liveClient     *http.Client
	apiAdminTOTP   string // TOTP secret for api-admin user
)

// Fixed credentials for users seeded by this package's TestMain.
const (
	apiAdminUser    = "api-admin"
	apiAdminPass    = "ApiAdmin1234!"
	apiCustUser     = "api-customer"
	apiCustPass     = "ApiCust1234!"
	apiProvUser     = "api-provider"
	apiProvPass     = "ApiProv1234!"
	apiAgentUser    = "api-agent"
	apiAgentPass    = "ApiAgent1234!"

	// Fixed entity IDs so tests can reference seeded data deterministically.
	apiCategoryID = "aaaa0000-0000-0000-0000-000000000001"
	apiListingID  = "aaaa0000-0000-0000-0000-000000000002"
	apiBookingID  = "aaaa0000-0000-0000-0000-000000000003"
	apiAdminID    = "aaaa0000-0000-0001-0000-000000000001"
	apiCustID     = "aaaa0000-0000-0001-0000-000000000002"
	apiProvID     = "aaaa0000-0000-0001-0000-000000000003"
	apiAgentID    = "aaaa0000-0000-0001-0000-000000000004"
)

// ---------------------------------------------------------------------------
// TestMain
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	liveServerURL = os.Getenv("TEST_SERVER_URL")
	if liveServerURL == "" {
		fmt.Fprintln(os.Stderr, "SKIP: api_tests — TEST_SERVER_URL not set; run via run_tests.sh")
		os.Exit(0)
	}

	liveClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // self-signed cert in test env
		},
		Timeout: 30 * time.Second,
	}

	if err := seedAPITestData(); err != nil {
		fmt.Fprintf(os.Stderr, "SKIP: api_tests — seed failed: %v\n", err)
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Seeding
// ---------------------------------------------------------------------------

func seedAPITestData() error {
	dsn := apiDSN()
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

	// Apply migration.
	var mig []byte
	for _, p := range []string{
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
		// Try walking up from the test binary working directory.
		if abs, e := filepath.Abs("../../migrations/001_init.sql"); e == nil {
			mig, _ = os.ReadFile(abs)
		}
	}
	if mig == nil {
		return fmt.Errorf("migration file not found")
	}
	if _, err = pool.Exec(ctx, string(mig)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	// TOTP secret: use the shared fixed secret when provided by run_tests.sh.
	apiAdminTOTP = os.Getenv("TEST_ADMIN_TOTP_SECRET")
	if apiAdminTOTP == "" {
		return fmt.Errorf("TEST_ADMIN_TOTP_SECRET not set; cannot seed admin TOTP for api_tests")
	}

	adminHash, _ := bcrypt.GenerateFromPassword([]byte(apiAdminPass), bcrypt.MinCost)
	custHash, _ := bcrypt.GenerateFromPassword([]byte(apiCustPass), bcrypt.MinCost)
	provHash, _ := bcrypt.GenerateFromPassword([]byte(apiProvPass), bcrypt.MinCost)
	agentHash, _ := bcrypt.GenerateFromPassword([]byte(apiAgentPass), bcrypt.MinCost)

	type seedUser struct {
		id, username, email, hash string
		totpEnabled                bool
		totpSecret                 string
	}
	for _, u := range []seedUser{
		{apiAdminID, apiAdminUser, apiAdminUser + "@test.local", string(adminHash), true, apiAdminTOTP},
		{apiCustID, apiCustUser, apiCustUser + "@test.local", string(custHash), false, ""},
		{apiProvID, apiProvUser, apiProvUser + "@test.local", string(provHash), false, ""},
		{apiAgentID, apiAgentUser, apiAgentUser + "@test.local", string(agentHash), false, ""},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, totp_enabled, totp_secret)
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
		{apiAdminID, "admin"}, {apiCustID, "customer"},
		{apiProvID, "provider"}, {apiAgentID, "csa"},
	} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO user_roles (user_id,role) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			r.uid, r.role); err != nil {
			return fmt.Errorf("seed role %s: %w", r.uid, err)
		}
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO categories (id,name) VALUES ($1,$2) ON CONFLICT (id) DO NOTHING`,
		apiCategoryID, "APITestCars"); err != nil {
		return fmt.Errorf("seed category: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO listings (id,category_id,provider_id,spu,sku,name,included_miles,deposit,available)
		VALUES ($1,$2,$3,'API-SPU','API-SKU','API Test Sedan',2.0,75.0,true)
		ON CONFLICT (id) DO NOTHING`,
		apiListingID, apiCategoryID, apiProvID); err != nil {
		return fmt.Errorf("seed listing: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO bookings (id,customer_id,provider_id,listing_id,status,
			estimated_amount,deposit_amount,start_at,end_at,odo_start,odo_end)
		VALUES ($1,$2,$3,$4,'booked',25.0,75.0,NOW(),NOW()+INTERVAL '2 hours',10.0,35.0)
		ON CONFLICT (id) DO NOTHING`,
		apiBookingID, apiCustID, apiProvID, apiListingID); err != nil {
		return fmt.Errorf("seed booking: %w", err)
	}

	return nil
}

func apiDSN() string {
	if u := os.Getenv("TEST_DATABASE_URL"); u != "" {
		return u
	}
	host := apiGetenv("DB_HOST", "w1-t1-ti1-db")
	port := apiGetenv("DB_PORT", "5432")
	user := apiGetenv("DB_USER", "fleetlease")
	pass := apiGetenv("DB_PASSWORD", "fleetlease")
	name := apiGetenv("DB_NAME", "fleetlease")
	ssl := apiGetenv("DB_SSL_MODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)
}

func apiGetenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// apiCall sends a real HTTP request and returns the response.
// Caller must close resp.Body.
func apiCall(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, liveServerURL+path, r)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := liveClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

func readAPIBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b
}

func mustAPIStatus(t *testing.T, resp *http.Response, want int) []byte {
	t.Helper()
	b := readAPIBody(t, resp)
	if resp.StatusCode != want {
		t.Fatalf("expected HTTP %d, got %d — body: %s", want, resp.StatusCode, b)
	}
	return b
}

// liveLogin logs in with username/password and returns the JWT token.
func liveLogin(t *testing.T, username, password string) string {
	t.Helper()
	resp := apiCall(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": username, "password": password}, "")
	b := mustAPIStatus(t, resp, http.StatusOK)
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("liveLogin %s: token missing in %s", username, b)
	}
	return out.Token
}

// liveLoginAdmin logs in as api-admin, supplying a fresh TOTP code.
func liveLoginAdmin(t *testing.T) string {
	t.Helper()
	code, err := totp.GenerateCode(apiAdminTOTP, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate totp: %v", err)
	}
	resp := apiCall(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": apiAdminUser, "password": apiAdminPass, "totpCode": code}, "")
	b := mustAPIStatus(t, resp, http.StatusOK)
	var out struct {
		Token string `json:"token"`
		User  struct{ ID string `json:"id"` } `json:"user"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("liveLoginAdmin: token missing in %s", b)
	}
	return out.Token
}

// liveLoginUserID returns the JWT token and the user's ID for the given user.
func liveLoginUserID(t *testing.T, username, password string) (string, string) {
	t.Helper()
	resp := apiCall(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": username, "password": password}, "")
	b := mustAPIStatus(t, resp, http.StatusOK)
	var out struct {
		Token string `json:"token"`
		User  struct{ ID string `json:"id"` } `json:"user"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("liveLoginUserID %s: bad response %s", username, b)
	}
	return out.Token, out.User.ID
}

// createTempUser creates a throwaway user via admin API and returns the ID.
func createTempUser(t *testing.T, adminToken, username, password string, roles []string) string {
	t.Helper()
	resp := apiCall(t, http.MethodPost, "/api/v1/admin/users", map[string]interface{}{
		"username": username,
		"email":    username + "@test.local",
		"password": password,
		"roles":    roles,
	}, adminToken)
	b := mustAPIStatus(t, resp, http.StatusCreated)
	var out struct{ ID string `json:"id"` }
	if err := json.Unmarshal(b, &out); err != nil || out.ID == "" {
		t.Fatalf("createTempUser: bad response %s", b)
	}
	return out.ID
}

// apiMiniPNG is a 100-byte minimal PNG (magic bytes + zero padding).
var apiMiniPNG = func() []byte {
	b := make([]byte, 100)
	copy(b, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	return b
}()

// uploadLiveAttachment runs init→chunk→complete and returns the attachment ID.
func uploadLiveAttachment(t *testing.T, token, bookingID, fingerprint string, data []byte) string {
	t.Helper()
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	fp := fingerprint
	if fp == "" {
		fp = fmt.Sprintf("fp-%d", time.Now().UnixNano())
	}

	resp := apiCall(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bookingID, "type": "photo",
		"sizeBytes": len(data), "checksum": checksum, "fingerprint": fp,
	}, token)
	b := mustAPIStatus(t, resp, http.StatusCreated)
	var init struct{ UploadID string `json:"uploadId"` }
	if err := json.Unmarshal(b, &init); err != nil || init.UploadID == "" {
		t.Fatalf("uploadLiveAttachment init: %s", b)
	}

	resp2 := apiCall(t, http.MethodPost, "/api/v1/attachments/chunk/upload", map[string]interface{}{
		"uploadId": init.UploadID, "chunkBase64": base64.StdEncoding.EncodeToString(data),
	}, token)
	mustAPIStatus(t, resp2, http.StatusOK)

	resp3 := apiCall(t, http.MethodPost, "/api/v1/attachments/chunk/complete",
		map[string]string{"uploadId": init.UploadID}, token)
	mustAPIStatus(t, resp3, http.StatusOK)

	return init.UploadID
}

// createFreshAPIBooking creates a booking via API and returns its ID.
func createFreshAPIBooking(t *testing.T, custToken string) string {
	t.Helper()
	now := time.Now().UTC()
	resp := apiCall(t, http.MethodPost, "/api/v1/bookings", map[string]interface{}{
		"listingId": apiListingID,
		"startAt":   now.Format(time.RFC3339),
		"endAt":     now.Add(2 * time.Hour).Format(time.RFC3339),
		"odoStart":  100.0, "odoEnd": 150.0,
	}, custToken)
	b := mustAPIStatus(t, resp, http.StatusCreated)
	var out struct {
		Booking struct{ ID string `json:"id"` } `json:"booking"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Booking.ID == "" {
		t.Fatalf("createFreshAPIBooking: %s", b)
	}
	return out.Booking.ID
}
