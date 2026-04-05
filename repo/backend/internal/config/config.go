package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
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
	DisableTLSEnforcement     bool
}

func Load() Config {
	cfg, err := loadFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func loadFromEnv() (Config, error) {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	isDevelopment := appEnv == "development"

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	dbPassword := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	encryptionKey := strings.TrimSpace(os.Getenv("AES256_KEY"))
	dbSSLMode := strings.ToLower(strings.TrimSpace(os.Getenv("DB_SSL_MODE")))

	if dbSSLMode == "" {
		if isDevelopment {
			dbSSLMode = "disable"
		} else {
			dbSSLMode = "require"
		}
	}

	if jwtSecret == "" {
		if isDevelopment {
			jwtSecret = randomSecret()
		} else {
			return Config{}, fmt.Errorf("JWT_SECRET is required when APP_ENV=%s", envLabel(appEnv))
		}
	}

	if encryptionKey == "" {
		if isDevelopment {
			encryptionKey = randomAES256Key()
		} else {
			return Config{}, fmt.Errorf("AES256_KEY is required when APP_ENV=%s", envLabel(appEnv))
		}
	}
	if len([]byte(encryptionKey)) != 32 {
		return Config{}, fmt.Errorf("AES256_KEY must be exactly 32 bytes")
	}

	if dbPassword == "" && !isDevelopment {
		return Config{}, fmt.Errorf("DB_PASSWORD is required when APP_ENV=%s", envLabel(appEnv))
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		if dbPassword == "" {
			if isDevelopment {
				dbPassword = "fleetlease"
			} else {
				return Config{}, fmt.Errorf("DB_PASSWORD must be provided when DATABASE_URL is not set")
			}
		}

		if !isDevelopment && !isSecureDBSSLMode(dbSSLMode) {
			return Config{}, fmt.Errorf("DB_SSL_MODE=%s is not allowed when APP_ENV=%s", dbSSLMode, envLabel(appEnv))
		}

		databaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("DB_USER", "fleetlease"),
			dbPassword,
			getEnv("DB_HOST", "db"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "fleetlease"),
			url.QueryEscape(dbSSLMode),
		)
	} else if !isDevelopment && strings.Contains(strings.ToLower(databaseURL), "sslmode=disable") {
		return Config{}, fmt.Errorf("DATABASE_URL cannot use sslmode=disable when APP_ENV=%s", envLabel(appEnv))
	}

	return Config{
		Port:                      getEnv("PORT", "8080"),
		JWTSecret:                 jwtSecret,
		AttachmentSigningSecret:   getEnv("ATTACHMENT_SIGNING_SECRET", jwtSecret),
		IdleTimeout:               time.Duration(getEnvInt("JWT_IDLE_MINUTES", 30)) * time.Minute,
		AbsoluteTimeout:           time.Duration(getEnvInt("JWT_ABSOLUTE_HOURS", 12)) * time.Hour,
		AdminAllowlistCIDR:        splitCSV(getEnv("ADMIN_ALLOWLIST", "127.0.0.1/32,::1/128")),
		TrustedProxiesCIDR:        splitCSV(getEnv("TRUSTED_PROXIES", "")),
		EncryptionKey:             encryptionKey,
		AttachmentDir:             getEnv("ATTACHMENT_DIR", "./data/attachments"),
		DatabaseURL:               databaseURL,
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
		DisableTLSEnforcement:     getEnvBool("DISABLE_TLS_ENFORCEMENT", false),
	}, nil
}

func randomSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Printf("failed to generate random JWT secret: %v", err)
		return fmt.Sprintf("dev-secret-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func randomAES256Key() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%032x", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func isSecureDBSSLMode(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "require", "verify-ca", "verify-full":
		return true
	default:
		return false
	}
}

func envLabel(appEnv string) string {
	if strings.TrimSpace(appEnv) == "" {
		return "<unset>"
	}
	return appEnv
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
