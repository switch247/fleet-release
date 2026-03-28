package handlers

import (
	"log/slog"

	"fleetlease/backend/internal/config"
	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"
	"fleetlease/backend/internal/store"
)

type Handler struct {
	Cfg     config.Config
	Store   store.Repository
	AuthSvc *services.AuthService
	Pricing services.PricingConfig
	Logger  *slog.Logger
	Metrics *services.WorkerMetrics
}

type userResponse struct {
	ID                     string        `json:"id"`
	Username               string        `json:"username"`
	Email                  string        `json:"email"`
	Roles                  []models.Role `json:"roles"`
	GovernmentIDMasked     string        `json:"governmentIdMasked"`
	PaymentReferenceMasked string        `json:"paymentReferenceMasked"`
	AddressMasked          string        `json:"addressMasked"`
}

type createUserRequest struct {
	Username         string        `json:"username"`
	Email            string        `json:"email"`
	Password         string        `json:"password"`
	Roles            []models.Role `json:"roles"`
	GovernmentID     string        `json:"governmentId"`
	PaymentReference string        `json:"paymentReference"`
	Address          string        `json:"address"`
}

type updateUserRequest struct {
	Email            string        `json:"email"`
	Password         string        `json:"password"`
	Roles            []models.Role `json:"roles"`
	GovernmentID     string        `json:"governmentId"`
	PaymentReference string        `json:"paymentReference"`
	Address          string        `json:"address"`
}

func sanitizeUser(u models.User) userResponse {
	return userResponse{
		ID:                     u.ID,
		Username:               u.Username,
		Email:                  u.Email,
		Roles:                  u.Roles,
		GovernmentIDMasked:     services.MaskSensitive(u.GovernmentIDEnc),
		PaymentReferenceMasked: services.MaskSensitive(u.PaymentReferenceEnc),
		AddressMasked:          services.MaskSensitive(u.AddressEnc),
	}
}

func containsAdminRole(roles []models.Role) bool {
	for _, r := range roles {
		if r == models.RoleAdmin {
			return true
		}
	}
	return false
}

func hasRole(roles []models.Role, target models.Role) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

func New(cfg config.Config, st store.Repository, auth *services.AuthService, logger *slog.Logger, metrics *services.WorkerMetrics) *Handler {
	return &Handler{Cfg: cfg, Store: st, AuthSvc: auth, Pricing: services.DefaultPricingConfig(), Logger: logger, Metrics: metrics}
}
