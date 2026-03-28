package api

import (
	"log/slog"
	"net/http"

	"fleetlease/backend/internal/config"
	"fleetlease/backend/internal/handlers"
	appmw "fleetlease/backend/internal/middleware"
	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"
	"fleetlease/backend/internal/store"

	"github.com/labstack/echo/v4"
	echoMW "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func NewRouter(cfg config.Config, st store.Repository, logger *slog.Logger, securityLogger *zap.Logger) *echo.Echo {
	e := echo.New()
	e.Use(echoMW.Recover())
	e.Use(echoMW.Logger())
	e.Use(echoMW.CORSWithConfig(echoMW.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:4173", "http://127.0.0.1:4173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
	}))
	e.Use(appmw.RequireTLSOrAllowlistedCIDRWithTrustedProxies(cfg.AllowInsecureCIDR, cfg.TrustedProxiesCIDR))
	e.Use(appmw.SecurityAuditLogWithTrustedProxies(securityLogger, cfg.TrustedProxiesCIDR))

	authSvc := services.NewAuthService(cfg.JWTSecret, cfg.IdleTimeout, cfg.AbsoluteTimeout, st)
	workerMetrics := services.NewWorkerMetrics()
	services.StartNotificationRetryWorker(st, cfg, logger, workerMetrics)
	h := handlers.New(cfg, st, authSvc, logger, workerMetrics)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/docs", h.OpenAPIDocsPage)
	e.GET("/docs/spec", h.OpenAPISpec)

	v1 := e.Group("/api/v1")
	v1.POST("/auth/login", h.Login)

	secured := v1.Group("")
	secured.Use(appmw.JWTAuth(authSvc))
	secured.POST("/auth/refresh", h.Refresh)
	secured.POST("/auth/logout", h.Logout)
	secured.POST("/auth/totp/enroll", h.TOTPEnroll)
	secured.POST("/auth/totp/verify", h.TOTPVerify)
	secured.GET("/auth/me", h.Me)
	secured.PATCH("/auth/me", h.UpdateMe)
	secured.GET("/auth/login-history", h.LoginHistory)

	adminOnly := secured.Group("/auth")
	adminOnlyMiddlewares := []echo.MiddlewareFunc{appmw.RequireRoles(models.RoleAdmin)}
	if cfg.RequireAdminMFA {
		adminOnlyMiddlewares = append(adminOnlyMiddlewares, appmw.RequireMFAForRoles(st, models.RoleAdmin))
	}
	adminOnly.Use(adminOnlyMiddlewares...)
	adminOnly.POST("/admin-reset", h.AdminResetPassword)

	secured.GET("/categories", h.Categories)
	secured.GET("/stats/summary", h.StatsSummary)
	secured.GET("/listings", h.Listings)
	secured.GET("/bookings", h.Bookings)
	secured.POST("/bookings/estimate", h.EstimateBooking)
	secured.POST("/bookings", h.CreateBooking)
	secured.POST("/coupons/redeem", h.RedeemCoupon)
	secured.POST("/inspections", h.UpsertInspection)
	secured.GET("/inspections", h.ListInspections)
	secured.GET("/inspections/verify/:bookingID", h.VerifyInspection)
	secured.POST("/attachments/chunk/init", h.AttachmentInit)
	secured.POST("/attachments/chunk/upload", h.AttachmentChunk)
	secured.POST("/attachments/chunk/complete", h.AttachmentComplete)
	secured.POST("/attachments/:id/presign", h.AttachmentPresign)
	// public serve endpoint for presigned attachments (signature validated in handler)
	v1.GET("/attachments/:id", h.AttachmentServe)
	secured.POST("/settlements/close/:bookingID", h.CloseSettlement)
	secured.GET("/ledger/:bookingID", h.Ledger)
	secured.GET("/ledger/:bookingID/verify", h.VerifyLedger)
	secured.POST("/complaints", h.CreateComplaint)
	secured.GET("/complaints", h.ListComplaints)
	secured.PATCH("/complaints/:id/arbitrate", h.ArbitrateComplaint)
	secured.POST("/consultations", h.CreateConsultation)
	secured.GET("/consultations", h.ListConsultations)
	secured.POST("/consultations/attachments", h.AttachConsultationEvidence)
	secured.GET("/consultations/:id/attachments", h.ListConsultationEvidence)
	secured.POST("/ratings", h.CreateRating)
	secured.GET("/ratings", h.ListRatings)
	secured.GET("/notifications", h.Notifications)
	secured.POST("/sync/reconcile", h.SyncReconcile)
	secured.GET("/exports/dispute-pdf/:id", h.ExportDisputePDF)

	admin := secured.Group("/admin")
	adminMiddlewares := []echo.MiddlewareFunc{
		appmw.RequireRoles(models.RoleAdmin),
		appmw.AdminIPAllowlist(cfg.AdminAllowlistCIDR, cfg.TrustedProxiesCIDR),
	}
	if cfg.RequireAdminMFA {
		adminMiddlewares = append(adminMiddlewares, appmw.RequireMFAForRoles(st, models.RoleAdmin))
	}
	admin.Use(adminMiddlewares...)
	admin.GET("/retention", h.AdminRetention)
	admin.POST("/retention/purge", h.AdminRunRetentionPurge)
	admin.POST("/backup/now", h.AdminBackupNow)
	admin.POST("/restore/now", h.AdminRestoreNow)
	admin.GET("/backup/jobs", h.AdminBackupJobs)
	admin.POST("/categories", h.AdminCreateCategory)
	admin.GET("/categories", h.AdminListCategories)
	admin.PATCH("/categories/:categoryID", h.AdminUpdateCategory)
	admin.DELETE("/categories/:categoryID", h.AdminDeleteCategory)
	admin.POST("/listings", h.AdminCreateListing)
	admin.GET("/listings", h.AdminListListings)
	admin.PATCH("/listings/:listingID", h.AdminUpdateListing)
	admin.DELETE("/listings/:listingID", h.AdminDeleteListing)
	admin.POST("/listings/bulk", h.AdminBulkListings)
	admin.GET("/listings/search", h.AdminSearchListings)
	admin.GET("/notification-templates", h.AdminListNotificationTemplates)
	admin.POST("/notification-templates", h.AdminCreateNotificationTemplate)
	admin.POST("/notifications/send", h.AdminSendNotification)
	admin.POST("/notifications/retry", h.AdminRetryNotifications)
	admin.GET("/workers/metrics", h.AdminWorkerMetrics)

	adminUsers := admin.Group("/users")
	adminUsers.GET("", h.AdminListUsers)
	adminUsers.POST("", h.AdminCreateUser)
	adminUsers.PATCH("/:userID", h.AdminUpdateUser)
	adminUsers.DELETE("/:userID", h.AdminDeleteUser)

	return e
}
