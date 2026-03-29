package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                      string
	JWTSecret                 string
	AttachmentSigningSecret   string
	IdleTimeout               time.Duration
	AbsoluteTimeout           time.Duration
	AdminAllowlistCIDR        []string
	TrustedProxiesCIDR        []string
	EncryptionKey             string
	AttachmentDir             string
	DatabaseURL               string
	StoreBackend              string
	TLSCertFile               string
	TLSKeyFile                string
	AllowInsecureCIDR         []string
	BackupRetentionDays       int
	AttachmentRetentionDays   int
	LedgerRetentionYears      int
	NotificationRetryMax      int
	NotificationRetryBackoffS int
	RetentionPurgeMinutes     int
	RequireAdminMFA           bool
}

func Load() Config {
	jwt := getEnv("JWT_SECRET", "dev-secret-change-me")
	return Config{
		Port:                      getEnv("PORT", "8080"),
		JWTSecret:                 jwt,
		AttachmentSigningSecret:   getEnv("ATTACHMENT_SIGNING_SECRET", jwt),
		IdleTimeout:               time.Duration(getEnvInt("JWT_IDLE_MINUTES", 30)) * time.Minute,
		AbsoluteTimeout:           time.Duration(getEnvInt("JWT_ABSOLUTE_HOURS", 12)) * time.Hour,
		AdminAllowlistCIDR:        splitCSV(getEnv("ADMIN_ALLOWLIST", "127.0.0.1/32,::1/128")),
		TrustedProxiesCIDR:        splitCSV(getEnv("TRUSTED_PROXIES", "")),
		EncryptionKey:             getEnv("AES256_KEY", "01234567890123456789012345678901"),
		AttachmentDir:             getEnv("ATTACHMENT_DIR", "./data/attachments"),
		DatabaseURL:               getEnv("DATABASE_URL", "postgres://fleetlease:fleetlease@db:5432/fleetlease?sslmode=disable"),
		StoreBackend:              strings.ToLower(getEnv("STORE_BACKEND", "postgres")),
		TLSCertFile:               getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:                getEnv("TLS_KEY_FILE", ""),
		AllowInsecureCIDR:         splitCSV(getEnv("ALLOW_INSECURE_HTTP_CIDRS", "127.0.0.1/32,::1/128")),
		BackupRetentionDays:       getEnvInt("BACKUP_RETENTION_DAYS", 30),
		AttachmentRetentionDays:   getEnvInt("ATTACHMENT_RETENTION_DAYS", 365),
		LedgerRetentionYears:      getEnvInt("LEDGER_RETENTION_YEARS", 7),
		NotificationRetryMax:      getEnvInt("NOTIFICATION_RETRY_MAX", 3),
		NotificationRetryBackoffS: getEnvInt("NOTIFICATION_RETRY_BACKOFF_SECONDS", 30),
		RetentionPurgeMinutes:     getEnvInt("RETENTION_PURGE_INTERVAL_MINUTES", 1440),
		RequireAdminMFA:           getEnvBool("REQUIRE_ADMIN_MFA", true),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
