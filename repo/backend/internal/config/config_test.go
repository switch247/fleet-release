package config

import (
	"strings"
	"testing"
)

func setCommonEnv(t *testing.T) {
	t.Helper()
	t.Setenv("PORT", "8080")
	t.Setenv("JWT_SECRET", "jwt-secret-for-tests")
	t.Setenv("DB_PASSWORD", "db-password-for-tests")
	t.Setenv("DB_USER", "fleetlease")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "fleetlease")
	t.Setenv("DATABASE_URL", "")
}

func TestLoadFromEnvFailsWithoutAESKeyOutsideDevelopment(t *testing.T) {
	setCommonEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AES256_KEY", "")

	_, err := loadFromEnv()
	if err == nil || !strings.Contains(err.Error(), "AES256_KEY is required") {
		t.Fatalf("expected AES256_KEY missing error, got: %v", err)
	}
}

func TestLoadFromEnvUsesSecureDBSSLModeOutsideDevelopment(t *testing.T) {
	setCommonEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AES256_KEY", "12345678901234567890123456789012")
	t.Setenv("DB_SSL_MODE", "")

	cfg, err := loadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cfg.DatabaseURL, "sslmode=require") {
		t.Fatalf("expected secure sslmode=require in DSN, got: %s", cfg.DatabaseURL)
	}
}

func TestLoadFromEnvHonorsDevDBSSLModeOverride(t *testing.T) {
	setCommonEnv(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("AES256_KEY", "12345678901234567890123456789012")
	t.Setenv("DB_SSL_MODE", "disable")

	cfg, err := loadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
		t.Fatalf("expected sslmode=disable in development DSN, got: %s", cfg.DatabaseURL)
	}
}
