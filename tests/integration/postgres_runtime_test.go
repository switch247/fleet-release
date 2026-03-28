package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fleetlease/backend/pkg/public"

	"github.com/jackc/pgx/v5/pgxpool"
)

func requirePostgresDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping postgres integration tests")
	}
	return dsn
}

func TestMigrationIdempotencyOnPostgres(t *testing.T) {
	dsn := requirePostgresDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("postgres connect failed: %v", err)
	}
	defer pool.Close()

	migrationPath := filepath.Join("..", "..", "backend", "migrations", "001_init.sql")
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("failed reading migration: %v", err)
	}
	if _, err = pool.Exec(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("first migration run failed: %v", err)
	}
	if _, err = pool.Exec(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("second migration run failed (not idempotent): %v", err)
	}
}

func TestPostgresBackedRuntimeViaPublicHarness(t *testing.T) {
	dsn := requirePostgresDSN(t)
	_ = dsn
	t.Setenv("TEST_STORE_BACKEND", "postgres")

	router := public.BuildSeededRouterForTests()
	body, _ := json.Marshal(map[string]string{"username": "customer", "password": "Customer1234!"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.RemoteAddr = "192.0.2.10:11000"
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", loginRec.Code, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	if loginResp.Token == "" {
		t.Fatalf("expected token")
	}

	bookingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	bookingsReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	bookingsReq.RemoteAddr = "192.0.2.10:11001"
	bookingsRec := httptest.NewRecorder()
	router.ServeHTTP(bookingsRec, bookingsReq)
	if bookingsRec.Code != http.StatusOK {
		t.Fatalf("bookings fetch failed: %d %s", bookingsRec.Code, bookingsRec.Body.String())
	}
}
