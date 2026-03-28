package handlers

import (
	"net/http"
	"strings"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totpCode"`
}

func (h *Handler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	user, ok := h.Store.GetUserByUsername(strings.TrimSpace(req.Username))
	if !ok {
		h.Store.SaveAuthEvent(models.AuthEvent{ID: uuid.NewString(), Username: strings.TrimSpace(req.Username), IP: requesterIP(c), EventType: "login_failed_unknown_user", CreatedAt: time.Now().UTC()})
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}
	if time.Now().UTC().Before(user.LockedUntil) {
		h.Store.SaveAuthEvent(models.AuthEvent{ID: uuid.NewString(), UserID: user.ID, Username: user.Username, IP: requesterIP(c), EventType: "login_blocked_lockout", CreatedAt: time.Now().UTC()})
		return c.JSON(http.StatusLocked, map[string]string{"error": "account locked for 15 minutes"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.FailedAttempts++
		if user.FailedAttempts >= 5 {
			user.LockedUntil = time.Now().UTC().Add(15 * time.Minute)
			user.FailedAttempts = 0
		}
		h.Store.SaveUser(user)
		h.Store.SaveAuthEvent(models.AuthEvent{ID: uuid.NewString(), UserID: user.ID, Username: user.Username, IP: requesterIP(c), EventType: "login_failed_invalid_password", CreatedAt: time.Now().UTC()})
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}
	if user.TOTPEnabled && !services.ValidateTOTPCode(user.TOTPSecret, req.TOTPCode) {
		h.Store.SaveAuthEvent(models.AuthEvent{ID: uuid.NewString(), UserID: user.ID, Username: user.Username, IP: requesterIP(c), EventType: "login_failed_invalid_totp", CreatedAt: time.Now().UTC()})
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid TOTP code"})
	}
	user.FailedAttempts = 0
	h.Store.SaveUser(user)
	token, session, err := h.AuthSvc.IssueToken(user)
	if err != nil {
		h.Logger.Error("failed to issue token", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
	}
	h.Store.SaveAuthEvent(models.AuthEvent{ID: uuid.NewString(), UserID: user.ID, Username: user.Username, IP: requesterIP(c), EventType: "login_success", CreatedAt: time.Now().UTC()})
	h.Logger.Info("login success", "user", user.Username)
	return c.JSON(http.StatusOK, map[string]interface{}{"token": token, "sessionId": session.ID, "user": sanitizeUser(user)})
}

func (h *Handler) Refresh(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	user, ok := h.Store.GetUserByID(userID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user"})
	}
	token, session, err := h.AuthSvc.IssueToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to refresh"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"token": token, "sessionId": session.ID})
}

func (h *Handler) Me(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	user, ok := h.Store.GetUserByID(userID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
	}
	return c.JSON(http.StatusOK, sanitizeUser(user))
}

func (h *Handler) UpdateMe(c echo.Context) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	userID, _ := c.Get("userID").(string)
	user, ok := h.Store.GetUserByID(userID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
	}
	user.Email = strings.TrimSpace(req.Email)
	h.Store.SaveUser(user)
	return c.JSON(http.StatusOK, sanitizeUser(user))
}

func (h *Handler) LoginHistory(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	return c.JSON(http.StatusOK, h.Store.ListAuthEventsByUser(userID, 50))
}

func (h *Handler) Logout(c echo.Context) error {
	sid, _ := c.Get("sid").(string)
	userID, _ := c.Get("userID").(string)
	user, _ := h.Store.GetUserByID(userID)
	h.AuthSvc.RevokeSession(sid)
	h.Store.SaveAuthEvent(models.AuthEvent{ID: uuid.NewString(), UserID: userID, Username: user.Username, IP: requesterIP(c), EventType: "logout", CreatedAt: time.Now().UTC()})
	return c.JSON(http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *Handler) TOTPEnroll(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	user, ok := h.Store.GetUserByID(userID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	secret, uri, err := services.GenerateTOTPSecret(user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate TOTP"})
	}
	user.TOTPSecret = secret
	h.Store.SaveUser(user)
	return c.JSON(http.StatusOK, map[string]string{"secret": secret, "uri": uri})
}

func (h *Handler) TOTPVerify(c echo.Context) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	userID, _ := c.Get("userID").(string)
	user, ok := h.Store.GetUserByID(userID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	if !services.ValidateTOTPCode(user.TOTPSecret, req.Code) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid code"})
	}
	user.TOTPEnabled = true
	h.Store.SaveUser(user)
	return c.JSON(http.StatusOK, map[string]string{"status": "enabled"})
}
