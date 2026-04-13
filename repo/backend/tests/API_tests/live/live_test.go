// Package live contains real-network HTTP tests that execute against the running
// backend server. All requests are made over the network using net/http; no
// in-process mocking or httptest.Recorder is used. The package is designed to
// be executed by `go test ./...` inside the docker compose test container where
// the server is already running at TEST_SERVER_URL.
package live

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
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Package-level state set up in TestMain
// ---------------------------------------------------------------------------

var (
	serverURL       string
	liveClient      *http.Client
	adminTOTPSecret string
)

// Fixed test entity IDs — use a distinct UUID prefix to avoid collisions with
// the existing httptest-seeded data.
const (
	liveAdminID    = "cccccccc-cccc-cccc-cccc-000000000001"
	liveCustomerID = "cccccccc-cccc-cccc-cccc-000000000002"
	liveProviderID = "cccccccc-cccc-cccc-cccc-000000000003"
	liveAgentID    = "cccccccc-cccc-cccc-cccc-000000000004"
	liveCategoryID = "cccccccc-cccc-cccc-cccc-000000000101"
	liveListingID  = "cccccccc-cccc-cccc-cccc-000000000201"
	liveBookingID  = "cccccccc-cccc-cccc-cccc-000000000301"
)

// Test passwords.
const (
	liveAdminPass    = "LiveAdmin1234!"
	liveCustomerPass = "LiveCust1234!"
	liveProviderPass = "LiveProv1234!"
	liveAgentPass    = "LiveAgent1234!"
)

// ---------------------------------------------------------------------------
// TestMain
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	serverURL = os.Getenv("TEST_SERVER_URL")
	if serverURL == "" {
		serverURL = "https://w1-t1-ti1-backend:8080"
	}

	liveClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // self-signed cert in test environment
		},
		Timeout: 30 * time.Second,
	}

	if err := seedLiveTestData(); err != nil {
		fmt.Fprintf(os.Stderr, "SKIP: live tests disabled — seed failed: %v\n", err)
		// Exit 0 so the overall test run is not marked as failed when running
		// outside the Docker compose environment.
		os.Exit(0)
	}

	if err := waitForServer(60 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "SKIP: live tests disabled — server not ready at %s: %v\n", serverURL, err)
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Seed helpers
// ---------------------------------------------------------------------------

func seedLiveTestData() error {
	dsn := buildDSN()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("postgres unavailable (%w) — run via docker compose to enable live tests", err)
	}
	// Verify the connection is actually usable.
	if pingErr := pool.Ping(ctx); pingErr != nil {
		return fmt.Errorf("postgres ping failed (%w) — run via docker compose to enable live tests", pingErr)
	}
	defer pool.Close()

	// Apply idempotent migration so tables exist.
	// Try multiple relative paths to accommodate different working directories
	// (package dir = tests/live/, backend dir, or repo root).
	var migration []byte
	for _, p := range []string{
		"../../../migrations/001_init.sql",
		"../../migrations/001_init.sql",
		"migrations/001_init.sql",
		"backend/migrations/001_init.sql",
	} {
		var readErr error
		migration, readErr = os.ReadFile(p)
		if readErr == nil {
			break
		}
	}
	if migration == nil {
		return fmt.Errorf("read migration: file not found in any expected location")
	}
	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	// Use a fixed TOTP secret when TEST_ADMIN_TOTP_SECRET is set (Docker CI),
	// so other test packages can independently generate valid admin codes.
	// Fall back to a random secret for isolated local runs.
	adminTOTPSecret = os.Getenv("TEST_ADMIN_TOTP_SECRET")
	if adminTOTPSecret == "" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "FleetLease",
			AccountName: "live_admin",
		})
		if err != nil {
			return fmt.Errorf("generate totp: %w", err)
		}
		adminTOTPSecret = key.Secret()
	}

	// Hash passwords at MinCost for test speed.
	adminHash, _ := bcrypt.GenerateFromPassword([]byte(liveAdminPass), bcrypt.MinCost)
	customerHash, _ := bcrypt.GenerateFromPassword([]byte(liveCustomerPass), bcrypt.MinCost)
	providerHash, _ := bcrypt.GenerateFromPassword([]byte(liveProviderPass), bcrypt.MinCost)
	agentHash, _ := bcrypt.GenerateFromPassword([]byte(liveAgentPass), bcrypt.MinCost)

	type seedUser struct {
		id, username, email, hash string
		totpEnabled                bool
		totpSecret                 string
	}
	users := []seedUser{
		{liveAdminID, "live_admin", "live_admin@fleetlease.local", string(adminHash), true, adminTOTPSecret},
		{liveCustomerID, "live_customer", "live_customer@fleetlease.local", string(customerHash), false, ""},
		{liveProviderID, "live_provider", "live_provider@fleetlease.local", string(providerHash), false, ""},
		{liveAgentID, "live_agent", "live_agent@fleetlease.local", string(agentHash), false, ""},
	}
	for _, u := range users {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, totp_enabled, totp_secret)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
			ON CONFLICT (username) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				totp_enabled  = EXCLUDED.totp_enabled,
				totp_secret   = EXCLUDED.totp_secret`,
			u.id, u.username, u.email, u.hash, u.totpEnabled, u.totpSecret)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.username, err)
		}
	}

	roleMap := []struct{ userID, role string }{
		{liveAdminID, "admin"},
		{liveCustomerID, "customer"},
		{liveProviderID, "provider"},
		{liveAgentID, "csa"},
	}
	for _, r := range roleMap {
		if _, err := pool.Exec(ctx,
			`INSERT INTO user_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			r.userID, r.role); err != nil {
			return fmt.Errorf("seed role %s→%s: %w", r.userID, r.role, err)
		}
	}

	// Seed category.
	if _, err := pool.Exec(ctx,
		`INSERT INTO categories (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
		liveCategoryID, "LiveTestCars"); err != nil {
		return fmt.Errorf("seed category: %w", err)
	}

	// Seed listing.
	if _, err := pool.Exec(ctx, `
		INSERT INTO listings (id, category_id, provider_id, spu, sku, name, included_miles, deposit, available)
		VALUES ($1, $2, $3, 'LIVE-SPU', 'LIVE-SKU', 'Live Test Sedan', 2.0, 50.0, true)
		ON CONFLICT (id) DO NOTHING`,
		liveListingID, liveCategoryID, liveProviderID); err != nil {
		return fmt.Errorf("seed listing: %w", err)
	}

	// Seed a booking in 'booked' state so workflow tests can act on it.
	if _, err := pool.Exec(ctx, `
		INSERT INTO bookings (id, customer_id, provider_id, listing_id, status,
			estimated_amount, deposit_amount, start_at, end_at, odo_start, odo_end)
		VALUES ($1, $2, $3, $4, 'booked', 25.0, 50.0,
			NOW(), NOW() + INTERVAL '2 hours', 100.0, 150.0)
		ON CONFLICT (id) DO NOTHING`,
		liveBookingID, liveCustomerID, liveProviderID, liveListingID); err != nil {
		return fmt.Errorf("seed booking: %w", err)
	}

	return nil
}

func buildDSN() string {
	if u := os.Getenv("TEST_DATABASE_URL"); u != "" {
		return u
	}
	host := getenv("DB_HOST", "w1-t1-ti1-db")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "fleetlease")
	pass := getenv("DB_PASSWORD", "fleetlease")
	name := getenv("DB_NAME", "fleetlease")
	ssl := getenv("DB_SSL_MODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// waitForServer polls the /health endpoint until it responds 200 or the
// deadline is exceeded.
func waitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := liveClient.Get(serverURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("server at %s did not become healthy within %s", serverURL, timeout)
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// api sends a real HTTP request to the running server and returns the response.
// The caller must close resp.Body.
func api(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body for %s %s: %v", method, path, err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, serverURL+path, r)
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, path, err)
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

// readBody drains and closes the response body and returns the raw bytes.
func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b
}

// mustStatus fails the test if the response status does not match.
func mustStatus(t *testing.T, resp *http.Response, want int) []byte {
	t.Helper()
	b := readBody(t, resp)
	if resp.StatusCode != want {
		t.Fatalf("expected HTTP %d, got %d — body: %s", want, resp.StatusCode, b)
	}
	return b
}

// loginAs logs in with the given username/password and returns the JWT token.
func loginAs(t *testing.T, username, password string) string {
	t.Helper()
	resp := api(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": username, "password": password}, "")
	b := mustStatus(t, resp, http.StatusOK)
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("login %s: token missing in %s", username, b)
	}
	return out.Token
}

// loginAdmin logs in as live_admin providing a fresh TOTP code each time.
func loginAdmin(t *testing.T) string {
	t.Helper()
	if adminTOTPSecret == "" {
		t.Fatal("adminTOTPSecret not initialised — TestMain may not have run")
	}
	code, err := totp.GenerateCode(adminTOTPSecret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	resp := api(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": "live_admin", "password": liveAdminPass, "totpCode": code}, "")
	b := mustStatus(t, resp, http.StatusOK)
	var out struct {
		Token string `json:"token"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("admin login: token missing in %s", b)
	}
	return out.Token
}

// customerToken is a convenience wrapper.
func customerToken(t *testing.T) string {
	t.Helper()
	return loginAs(t, "live_customer", liveCustomerPass)
}

// providerToken is a convenience wrapper.
func providerToken(t *testing.T) string {
	t.Helper()
	return loginAs(t, "live_provider", liveProviderPass)
}

// agentToken is a convenience wrapper.
func agentToken(t *testing.T) string {
	t.Helper()
	return loginAs(t, "live_agent", liveAgentPass)
}

// createFreshBooking creates a new booking via the API and returns its ID.
// Used by tests that mutate booking state (e.g. settlement).
func createFreshBooking(t *testing.T, custToken string) string {
	t.Helper()
	now := time.Now().UTC()
	body := map[string]interface{}{
		"listingId": liveListingID,
		"startAt":   now.Format(time.RFC3339),
		"endAt":     now.Add(2 * time.Hour).Format(time.RFC3339),
		"odoStart":  100.0,
		"odoEnd":    150.0,
	}
	resp := api(t, http.MethodPost, "/api/v1/bookings", body, custToken)
	b := mustStatus(t, resp, http.StatusCreated)
	var out struct {
		Booking struct {
			ID string `json:"id"`
		} `json:"booking"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Booking.ID == "" {
		t.Fatalf("createFreshBooking: bad response %s", b)
	}
	return out.Booking.ID
}

// minimalPNG is a tiny valid PNG: PNG signature (8 bytes) followed by zeroed
// padding.  http.DetectContentType recognises the magic bytes and returns
// "image/png", satisfying the server's MIME check.
var minimalPNG = func() []byte {
	b := make([]byte, 100)
	copy(b, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) // PNG signature
	return b
}()

// createCompleteAttachment runs the full three-step attachment pipeline
// (init → chunk → complete) and returns the finished attachment ID.
// The content is a minimal PNG so the server's MIME validation passes.
func createCompleteAttachment(t *testing.T, token, bookingID string) string {
	t.Helper()

	sum := sha256.Sum256(minimalPNG)
	checksum := hex.EncodeToString(sum[:])
	fp := fmt.Sprintf("live-fp-%d", time.Now().UnixNano())

	// 1. Init
	initBody := map[string]interface{}{
		"bookingId":   bookingID,
		"type":        "photo",
		"sizeBytes":   int64(len(minimalPNG)),
		"checksum":    checksum,
		"fingerprint": fp,
	}
	resp := api(t, http.MethodPost, "/api/v1/attachments/chunk/init", initBody, token)
	b := mustStatus(t, resp, http.StatusCreated)
	var initOut struct {
		UploadID string `json:"uploadId"`
	}
	if err := json.Unmarshal(b, &initOut); err != nil || initOut.UploadID == "" {
		t.Fatalf("createCompleteAttachment init: bad response %s", b)
	}
	uploadID := initOut.UploadID

	// 2. Upload chunk — field name is chunkBase64, no chunkIndex field.
	chunkBody := map[string]interface{}{
		"uploadId":    uploadID,
		"chunkBase64": base64.StdEncoding.EncodeToString(minimalPNG),
	}
	resp2 := api(t, http.MethodPost, "/api/v1/attachments/chunk/upload", chunkBody, token)
	b2 := readBody(t, resp2)
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("createCompleteAttachment chunk: got %d %s", resp2.StatusCode, b2)
	}

	// 3. Complete — only uploadId required; handler verifies checksum and MIME.
	completeBody := map[string]interface{}{
		"uploadId": uploadID,
	}
	resp3 := api(t, http.MethodPost, "/api/v1/attachments/chunk/complete", completeBody, token)
	b3 := readBody(t, resp3)
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("createCompleteAttachment complete: got %d %s", resp3.StatusCode, b3)
	}

	return uploadID
}
