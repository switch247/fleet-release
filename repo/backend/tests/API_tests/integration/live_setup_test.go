// live_setup_test.go provides shared live-HTTP infrastructure for integration
// tests that have been converted to make real network calls.
//
// The following files remain in-process because they require capabilities that
// are impossible over real TCP:
//   - settlement_test.go   — h.TamperLedger() writes directly to the in-memory store
//   - postgres_runtime_test.go — sets req.RemoteAddr to inject IP addresses
package integration

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

var (
	intLiveServerURL string
	intLiveClient    *http.Client
	intAdminTOTP     string
)

const (
	intAdminUser  = "int-admin"
	intAdminPass  = "IntAdmin1234!"
	intCustUser   = "int-customer"
	intCustPass   = "IntCust1234!"
	intProvUser   = "int-provider"
	intProvPass   = "IntProv1234!"
	intAgentUser  = "int-agent"
	intAgentPass  = "IntAgent1234!"

	intAdminID    = "bbbb0000-0000-0001-0000-000000000001"
	intCustID     = "bbbb0000-0000-0001-0000-000000000002"
	intProvID     = "bbbb0000-0000-0001-0000-000000000003"
	intAgentID    = "bbbb0000-0000-0001-0000-000000000004"
	intCategoryID = "bbbb0000-0000-0000-0000-000000000001"
	intListingID  = "bbbb0000-0000-0000-0000-000000000002"
	intBookingID  = "bbbb0000-0000-0000-0000-000000000003"
)

// intMiniPNG is a minimal valid PNG used for attachment upload tests.
var intMiniPNG = func() []byte {
	b := make([]byte, 100)
	copy(b, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	return b
}()

// TestMain always runs in-process tests (settlement_test.go,
// postgres_runtime_test.go) and additionally enables live-HTTP tests when
// TEST_SERVER_URL is set.  Live tests call skipIfNoIntLive(t) at the top.
func TestMain(m *testing.M) {
	if url := os.Getenv("TEST_SERVER_URL"); url != "" {
		intLiveServerURL = url
		intLiveClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
			Timeout: 30 * time.Second,
		}
		intAdminTOTP = os.Getenv("TEST_ADMIN_TOTP_SECRET")
		if intAdminTOTP == "" {
			fmt.Fprintln(os.Stderr, "WARN: TEST_ADMIN_TOTP_SECRET not set — integration live tests will skip")
			intLiveServerURL = ""
		} else if err := seedIntTestData(); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: integration live seed failed: %v — live tests will skip\n", err)
			intLiveServerURL = ""
		}
	}
	os.Exit(m.Run())
}

func skipIfNoIntLive(t *testing.T) {
	t.Helper()
	if intLiveServerURL == "" {
		t.Skip("TEST_SERVER_URL not set — skipping live integration test")
	}
}

func seedIntTestData() error {
	dsn := intDSN()
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

	adminHash, _ := bcrypt.GenerateFromPassword([]byte(intAdminPass), bcrypt.MinCost)
	custHash, _ := bcrypt.GenerateFromPassword([]byte(intCustPass), bcrypt.MinCost)
	provHash, _ := bcrypt.GenerateFromPassword([]byte(intProvPass), bcrypt.MinCost)
	agentHash, _ := bcrypt.GenerateFromPassword([]byte(intAgentPass), bcrypt.MinCost)

	for _, u := range []struct {
		id, username, email, hash string
		totpEnabled                bool
		totpSecret                 string
	}{
		{intAdminID, intAdminUser, intAdminUser + "@test.local", string(adminHash), true, intAdminTOTP},
		{intCustID, intCustUser, intCustUser + "@test.local", string(custHash), false, ""},
		{intProvID, intProvUser, intProvUser + "@test.local", string(provHash), false, ""},
		{intAgentID, intAgentUser, intAgentUser + "@test.local", string(agentHash), false, ""},
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
		{intAdminID, "admin"}, {intCustID, "customer"},
		{intProvID, "provider"}, {intAgentID, "csa"},
	} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO user_roles (user_id,role) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			r.uid, r.role); err != nil {
			return fmt.Errorf("seed role: %w", err)
		}
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO categories (id,name) VALUES ($1,$2) ON CONFLICT (id) DO NOTHING`,
		intCategoryID, "IntTestCars"); err != nil {
		return fmt.Errorf("seed category: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO listings (id,category_id,provider_id,spu,sku,name,included_miles,deposit,available)
		VALUES ($1,$2,$3,'INT-SPU','INT-SKU','Int Test Sedan',2.0,75.0,true)
		ON CONFLICT (id) DO NOTHING`,
		intListingID, intCategoryID, intProvID); err != nil {
		return fmt.Errorf("seed listing: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO bookings (id,customer_id,provider_id,listing_id,status,
			estimated_amount,deposit_amount,start_at,end_at,odo_start,odo_end)
		VALUES ($1,$2,$3,$4,'booked',25.0,75.0,NOW(),NOW()+INTERVAL '2 hours',10.0,35.0)
		ON CONFLICT (id) DO NOTHING`,
		intBookingID, intCustID, intProvID, intListingID); err != nil {
		return fmt.Errorf("seed booking: %w", err)
	}
	return nil
}

func intDSN() string {
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

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func intAPI(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, intLiveServerURL+path, r)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := intLiveClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

func intReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b
}

func intMustStatus(t *testing.T, resp *http.Response, want int) []byte {
	t.Helper()
	b := intReadBody(t, resp)
	if resp.StatusCode != want {
		t.Fatalf("expected HTTP %d, got %d — body: %s", want, resp.StatusCode, b)
	}
	return b
}

func intLogin(t *testing.T, username, password string) string {
	t.Helper()
	resp := intAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": username, "password": password}, "")
	b := intMustStatus(t, resp, http.StatusOK)
	var out struct{ Token string `json:"token"` }
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("intLogin %s: no token in %s", username, b)
	}
	return out.Token
}

func intLoginAdmin(t *testing.T) string {
	t.Helper()
	code, err := totp.GenerateCode(intAdminTOTP, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate totp: %v", err)
	}
	resp := intAPI(t, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"username": intAdminUser, "password": intAdminPass, "totpCode": code}, "")
	b := intMustStatus(t, resp, http.StatusOK)
	var out struct{ Token string `json:"token"` }
	if err := json.Unmarshal(b, &out); err != nil || out.Token == "" {
		t.Fatalf("intLoginAdmin: no token in %s", b)
	}
	return out.Token
}

// intCreateBooking creates a booking via API and returns its ID.
func intCreateBooking(t *testing.T, custToken string) string {
	t.Helper()
	now := time.Now().UTC()
	resp := intAPI(t, http.MethodPost, "/api/v1/bookings", map[string]interface{}{
		"listingId": intListingID,
		"startAt":   now.Format(time.RFC3339),
		"endAt":     now.Add(2 * time.Hour).Format(time.RFC3339),
		"odoStart":  100.0, "odoEnd": 150.0,
	}, custToken)
	b := intMustStatus(t, resp, http.StatusCreated)
	var out struct {
		Booking struct{ ID string `json:"id"` } `json:"booking"`
	}
	if err := json.Unmarshal(b, &out); err != nil || out.Booking.ID == "" {
		t.Fatalf("intCreateBooking: %s", b)
	}
	return out.Booking.ID
}

// intUploadAttachment runs init→chunk→complete and returns the attachment ID.
func intUploadAttachment(t *testing.T, token, bookingID, fingerprint string, data []byte) string {
	t.Helper()
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	fp := fingerprint
	if fp == "" {
		fp = fmt.Sprintf("int-fp-%d", time.Now().UnixNano())
	}

	resp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bookingID, "type": "photo",
		"sizeBytes": len(data), "checksum": checksum, "fingerprint": fp,
	}, token)
	b := intMustStatus(t, resp, http.StatusCreated)
	var init struct{ UploadID string `json:"uploadId"` }
	if err := json.Unmarshal(b, &init); err != nil || init.UploadID == "" {
		t.Fatalf("intUploadAttachment init: %s", b)
	}

	resp2 := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/upload", map[string]interface{}{
		"uploadId": init.UploadID, "chunkBase64": base64.StdEncoding.EncodeToString(data),
	}, token)
	intMustStatus(t, resp2, http.StatusOK)

	resp3 := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/complete",
		map[string]string{"uploadId": init.UploadID}, token)
	intMustStatus(t, resp3, http.StatusOK)

	return init.UploadID
}
