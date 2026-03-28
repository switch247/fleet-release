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

func (h *Handler) AdminResetPassword(c echo.Context) error {
	var req struct {
		Username    string `json:"username"`
		NewPassword string `json:"newPassword"`
		CheckedBy   string `json:"checkedBy"`
		Method      string `json:"method"`
		EvidenceRef string `json:"evidenceRef"`
		Reason      string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	actorID, _ := c.Get("userID").(string)
	if strings.TrimSpace(req.CheckedBy) == "" || strings.TrimSpace(req.Method) == "" || strings.TrimSpace(req.EvidenceRef) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "checkedBy, method and evidenceRef are required"})
	}
	if req.CheckedBy != actorID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "checkedBy must match authenticated administrator"})
	}
	if err := services.ValidatePasswordComplexity(req.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	user, ok := h.Store.GetUserByUsername(req.Username)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.FailedAttempts = 0
	user.LockedUntil = time.Time{}
	h.Store.SaveUser(user)
	h.Store.SavePasswordResetEvidence(models.PasswordResetEvidence{
		ID:           uuid.NewString(),
		TargetUserID: user.ID,
		CheckedBy:    req.CheckedBy,
		Method:       req.Method,
		EvidenceRef:  req.EvidenceRef,
		Reason:       req.Reason,
		CreatedAt:    time.Now().UTC(),
	})
	return c.JSON(http.StatusOK, map[string]string{"status": "password reset"})
}

func (h *Handler) AdminListUsers(c echo.Context) error {
	users := h.Store.ListUsers()
	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, sanitizeUser(u))
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) AdminCreateUser(c echo.Context) error {
	var req createUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username required"})
	}
	if h.Store.UsernameExists(username) {
		return c.JSON(http.StatusConflict, map[string]string{"error": "username already exists"})
	}
	if err := services.ValidatePasswordComplexity(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	roles := req.Roles
	if len(roles) == 0 {
		roles = []models.Role{models.RoleCustomer}
	}
	if containsAdminRole(roles) && h.Store.HasAdminExcluding("") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "only one admin is permitted"})
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	gov := ""
	if req.GovernmentID != "" {
		enc, err := services.EncryptAES256([]byte(h.Cfg.EncryptionKey), req.GovernmentID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encrypt government ID"})
		}
		gov = services.MaskSensitive(enc)
	}
	user := models.User{ID: uuid.NewString(), Username: username, Email: strings.TrimSpace(req.Email), PasswordHash: string(hash), Roles: roles, GovernmentIDEnc: gov}
	h.Store.SaveUser(user)
	return c.JSON(http.StatusCreated, sanitizeUser(user))
}

func (h *Handler) AdminUpdateUser(c echo.Context) error {
	id := c.Param("userID")
	user, ok := h.Store.GetUserByID(id)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	var req updateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if len(req.Roles) > 0 {
		if containsAdminRole(req.Roles) && h.Store.HasAdminExcluding(id) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "only one admin is permitted"})
		}
		user.Roles = req.Roles
	}
	if req.Email != "" {
		user.Email = strings.TrimSpace(req.Email)
	}
	if req.Password != "" {
		if err := services.ValidatePasswordComplexity(req.Password); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.PasswordHash = string(hash)
	}
	if req.GovernmentID != "" {
		enc, err := services.EncryptAES256([]byte(h.Cfg.EncryptionKey), req.GovernmentID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encrypt government ID"})
		}
		user.GovernmentIDEnc = services.MaskSensitive(enc)
	}
	h.Store.SaveUser(user)
	return c.JSON(http.StatusOK, sanitizeUser(user))
}

func (h *Handler) AdminDeleteUser(c echo.Context) error {
	target := c.Param("userID")
	actor, _ := c.Get("userID").(string)
	if target == actor {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot delete self"})
	}
	if _, ok := h.Store.GetUserByID(target); !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	h.Store.DeleteUser(target)
	return c.NoContent(http.StatusNoContent)
}
