package public

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type TestHarness struct {
	Router    *echo.Echo
	BookingID string
	tamper    func(bookingID string)
}

func (h *TestHarness) TamperLedger(bookingID string) {
	if h == nil || h.tamper == nil {
		return
	}
	h.tamper(bookingID)
}

type EstimateInput struct {
	StartAt  time.Time
	EndAt    time.Time
	OdoStart float64
	OdoEnd   float64
	Deposit  float64
}

func EstimateFare(input EstimateInput) services.EstimateResult {
	return services.EstimateFare(services.DefaultPricingConfig(), services.EstimateInput{
		StartAt:  input.StartAt,
		EndAt:    input.EndAt,
		OdoStart: input.OdoStart,
		OdoEnd:   input.OdoEnd,
		Deposit:  input.Deposit,
	})
}

func BuildSeededRouterForTests() *echo.Echo {
	cfg := config.Load()
	cfg.AllowInsecureCIDR = []string{"127.0.0.1/32", "::1/128", "192.0.2.0/24"}
	cfg.AdminAllowlistCIDR = []string{"127.0.0.1/32", "::1/128", "192.0.2.0/24"}
	cfg.RequireAdminMFA = strings.EqualFold(os.Getenv("TEST_REQUIRE_ADMIN_MFA"), "true")
	st := buildTestStoreWithFallback(cfg)
	customerID := uuid.NewString()
	providerID := uuid.NewString()
	csaID := uuid.NewString()
	adminID := uuid.NewString()
	seedUser := func(id, username, password, gov string, roles ...models.Role) models.User {
		if existing, ok := st.GetUserByUsername(username); ok {
			return existing
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		encrypted, _ := services.EncryptAES256([]byte(cfg.EncryptionKey), gov)
		user := models.User{ID: id, Username: username, Email: username + "@fleetlease.local", PasswordHash: string(hash), Roles: roles, GovernmentIDEnc: services.MaskSensitive(encrypted)}
		st.SaveUser(user)
		return user
	}
	customer := seedUser(customerID, "customer", "Customer1234!", "A-CUST-1122", models.RoleCustomer)
	provider := seedUser(providerID, "provider", "Provider1234!", "A-PROV-1122", models.RoleProvider)
	seedUser(csaID, "agent", "Agent1234!Pass", "A-CSA-1122", models.RoleCSA)
	seedUser(adminID, "admin", "Admin1234!Pass", "A-ADMIN-1122", models.RoleAdmin)
	categoryID := uuid.NewString()
	listingID := "11111111-1111-1111-1111-111111111111"
	st.SaveCategory(models.Category{ID: categoryID, Name: "Cars"})
	st.SaveListing(models.Listing{ID: listingID, CategoryID: categoryID, ProviderID: provider.ID, SPU: "SEDAN-SPU", SKU: "SEDAN-SKU-A", Name: "City Sedan", IncludedMiles: 2, Deposit: 75, Available: true})
	st.SaveBooking(models.Booking{ID: "22222222-2222-2222-2222-222222222222", CustomerID: customer.ID, ProviderID: provider.ID, ListingID: listingID, Status: "booked", DepositAmount: 75, EstimatedAmount: 25})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	securityLogger, _ := applogger.New()
	return api.NewRouter(cfg, st, logger, securityLogger)
}

func BuildHarnessForTests() *TestHarness {
	cfg := config.Load()
	cfg.AllowInsecureCIDR = []string{"127.0.0.1/32", "::1/128", "192.0.2.0/24"}
	cfg.AdminAllowlistCIDR = []string{"127.0.0.1/32", "::1/128", "192.0.2.0/24"}
	cfg.RequireAdminMFA = strings.EqualFold(os.Getenv("TEST_REQUIRE_ADMIN_MFA"), "true")
	st := buildTestStoreWithFallback(cfg)
	customerID := uuid.NewString()
	providerID := uuid.NewString()
	seedUser := func(id, username, password string, roles ...models.Role) models.User {
		if existing, ok := st.GetUserByUsername(username); ok {
			return existing
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		encrypted, _ := services.EncryptAES256([]byte(cfg.EncryptionKey), username+"-id")
		user := models.User{ID: id, Username: username, Email: username + "@fleetlease.local", PasswordHash: string(hash), Roles: roles, GovernmentIDEnc: services.MaskSensitive(encrypted)}
		st.SaveUser(user)
		return user
	}
	customer := seedUser(customerID, "customer", "Customer1234!", models.RoleCustomer)
	provider := seedUser(providerID, "provider", "Provider1234!", models.RoleProvider)
	seedUser(uuid.NewString(), "agent", "Agent1234!Pass", models.RoleCSA)
	seedUser(uuid.NewString(), "admin", "Admin1234!Pass", models.RoleAdmin)
	categoryID := uuid.NewString()
	listingID := uuid.NewString()
	bookingID := uuid.NewString()
	st.SaveCategory(models.Category{ID: categoryID, Name: "Cars"})
	st.SaveListing(models.Listing{ID: listingID, CategoryID: categoryID, ProviderID: provider.ID, SPU: "SEDAN-SPU", SKU: "SEDAN-SKU-A", Name: "City Sedan", IncludedMiles: 2, Deposit: 75, Available: true})
	st.SaveBooking(models.Booking{
		ID: bookingID, CustomerID: customer.ID, ProviderID: provider.ID, ListingID: listingID, CouponCode: "",
		StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(2 * time.Hour), OdoStart: 10, OdoEnd: 35, Status: "booked", EstimatedAmount: 25, DepositAmount: 75,
	})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	securityLogger, _ := applogger.New()
	router := api.NewRouter(cfg, st, logger, securityLogger)
	return &TestHarness{
		Router:    router,
		BookingID: bookingID,
		tamper: func(bID string) {
			st.AppendLedger(bID, models.LedgerEntry{
				ID:          uuid.NewString(),
				BookingID:   bID,
				Type:        "tamper",
				Amount:      9999,
				Description: "tamper",
				CreatedAt:   time.Now().UTC(),
				PrevHash:    "invalid-prev-hash",
				Hash:        "invalid-hash",
			})
		},
	}
}

func buildTestStoreWithFallback(cfg config.Config) store.Repository {
	backend := strings.ToLower(os.Getenv("TEST_STORE_BACKEND"))
	if backend == "" {
		backend = "postgres"
	}
	if backend == "memory" {
		return store.NewMemoryStore()
	}

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = cfg.DatabaseURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Println("TEST HINT: postgres unavailable; falling back to memory. Set TEST_STORE_BACKEND=memory or start PostgreSQL and set TEST_DATABASE_URL.")
		return store.NewMemoryStore()
	}
	defer pool.Close()

	migrationPath := filepath.Join("backend", "migrations", "001_init.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		migrationPath = filepath.Join("migrations", "001_init.sql")
		migrationSQL, err = os.ReadFile(migrationPath)
	}
	if err != nil {
		fmt.Println("TEST HINT: migration not found; falling back to memory.")
		return store.NewMemoryStore()
	}
	if _, err = pool.Exec(ctx, string(migrationSQL)); err != nil {
		fmt.Println("TEST HINT: postgres migration failed; falling back to memory.")
		return store.NewMemoryStore()
	}
	pgStore, err := store.NewPostgresStore(context.Background(), dsn)
	if err != nil {
		fmt.Println("TEST HINT: postgres store init failed; falling back to memory.")
		return store.NewMemoryStore()
	}
	return pgStore
}
