package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fleetlease/backend/internal/api"
	"fleetlease/backend/internal/config"
	applogger "fleetlease/backend/internal/logger"
	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"
	"fleetlease/backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	securityLogger, err := applogger.New()
	if err != nil {
		log.Fatalf("failed to create security logger: %v", err)
	}
	defer securityLogger.Sync()

	repo := initRepository(cfg, logger)
	seed(repo, cfg)
	startNightlyBackupScheduler(repo, logger)
	startRetentionPurgeScheduler(repo, cfg, logger)

	e := api.NewRouter(cfg, repo, logger, securityLogger)
	addr := ":" + cfg.Port
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		logger.Info("starting HTTPS server", "addr", addr)
		if err := e.StartTLS(addr, cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
			log.Fatal(err)
		}
		return
	}
	logger.Warn("TLS certificates not configured; HTTP allowed only for allowlisted CIDRs", "allowlist", strings.Join(cfg.AllowInsecureCIDR, ","))
	if err := e.Start(addr); err != nil {
		log.Fatal(err)
	}
}

func startNightlyBackupScheduler(st store.Repository, logger *slog.Logger) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next))
			job := models.BackupJob{
				ID:        uuid.NewString(),
				Type:      "backup",
				Status:    "running",
				CreatedAt: time.Now().UTC(),
			}
			st.SaveBackupJob(job)
			cmd := exec.Command("sh", "scripts/backup.sh")
			if out, err := cmd.CombinedOutput(); err != nil {
				job.Status = "failed"
				job.Error = string(out)
				job.FinishedAt = time.Now().UTC()
				st.SaveBackupJob(job)
				logger.Error("nightly backup failed", "error", err)
				continue
			}
			job.Status = "completed"
			job.Artifact = "local-backup"
			job.FinishedAt = time.Now().UTC()
			st.SaveBackupJob(job)
			logger.Info("nightly backup completed")
		}
	}()
}

func startRetentionPurgeScheduler(st store.Repository, cfg config.Config, logger *slog.Logger) {
	go func() {
		interval := time.Duration(cfg.RetentionPurgeMinutes) * time.Minute
		if interval <= 0 {
			interval = 24 * time.Hour
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		// Run immediately on startup.
		forceRunRetention(st, cfg, logger)
		for range ticker.C {
			forceRunRetention(st, cfg, logger)
		}
	}()
}

func forceRunRetention(st store.Repository, cfg config.Config, logger *slog.Logger) {
	result := services.RunRetentionPurge(st, cfg, logger)
	logger.Info("scheduled_retention_purge", "reportID", result.ID, "attachmentsDeleted", result.AttachmentsDeleted, "ledgerDeleted", result.LedgerDeleted)
}

func initRepository(cfg config.Config, logger *slog.Logger) store.Repository {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if cfg.StoreBackend == "memory" {
		if env != "test" {
			log.Fatal("memory backend is only permitted when APP_ENV=test")
		}
		logger.Warn("using in-memory store for test environment")
		return store.NewMemoryStore()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pool.Close()

	migrationPath := filepath.Join("backend", "migrations", "001_init.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		migrationPath = filepath.Join("migrations", "001_init.sql")
		migrationSQL, err = os.ReadFile(migrationPath)
	}
	if err != nil {
		log.Fatalf("failed to read migration: %v", err)
	}
	if _, err = pool.Exec(ctx, string(migrationSQL)); err != nil {
		log.Fatalf("failed applying migration: %v", err)
	}
	logger.Info("postgres schema ensured")

	repo, err := store.NewPostgresStore(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to initialize postgres store: %v", err)
	}
	return repo
}

func seed(st store.Repository, cfg config.Config) {
	if st.UsernameExists("admin") {
		return
	}
	seedUser := func(username, password, govID string, roles ...models.Role) models.User {
		if err := services.ValidatePasswordComplexity(password); err != nil {
			panic(err)
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		encGovID, _ := services.EncryptAES256([]byte(cfg.EncryptionKey), govID)
		u := models.User{ID: uuid.NewString(), Username: username, Email: username + "@fleetlease.local", PasswordHash: string(hash), Roles: roles, GovernmentIDEnc: services.MaskSensitive(encGovID)}
		st.SaveUser(u)
		return u
	}

	seedUser("admin", "Admin1234!Pass", "A-ADMIN-7788", models.RoleAdmin)
	seedUser("customer", "Customer1234!", "A-CUST-1122", models.RoleCustomer)
	provider := seedUser("provider", "Provider1234!", "A-PROV-9900", models.RoleProvider)
	seedUser("agent", "Agent1234!Pass", "A-CSA-5566", models.RoleCSA)

	cat1 := models.Category{ID: uuid.NewString(), Name: "Cars"}
	cat2 := models.Category{ID: uuid.NewString(), Name: "Vans"}
	st.SaveCategory(cat1)
	st.SaveCategory(cat2)

	st.SaveListing(models.Listing{ID: uuid.NewString(), CategoryID: cat1.ID, ProviderID: provider.ID, SPU: "SEDAN-SPU", SKU: "SEDAN-SKU-A", Name: "City Sedan", IncludedMiles: 2, Deposit: 75, Available: true})
	st.SaveListing(models.Listing{ID: uuid.NewString(), CategoryID: cat2.ID, ProviderID: provider.ID, SPU: "VAN-SPU", SKU: "VAN-SKU-X", Name: "Cargo Van", IncludedMiles: 3, Deposit: 90, Available: true})
}
