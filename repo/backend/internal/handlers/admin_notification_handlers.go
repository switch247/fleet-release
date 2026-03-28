package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"fleetlease/backend/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) AdminListNotificationTemplates(c echo.Context) error {
	return c.JSON(http.StatusOK, h.Store.ListNotificationTemplates())
}

func (h *Handler) AdminCreateNotificationTemplate(c echo.Context) error {
	var req struct {
		Name    string `json:"name"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Channel string `json:"channel"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.Channel = strings.ToLower(strings.TrimSpace(req.Channel))
	if req.Name == "" || req.Title == "" || req.Body == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name, title, and body are required"})
	}
	if req.Channel == "" {
		req.Channel = "in_app"
	}
	switch req.Channel {
	case "in_app", "email", "sms":
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "channel must be one of: in_app, email, sms"})
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	actor, _ := c.Get("userID").(string)
	template := models.NotificationTemplate{
		ID:         uuid.NewString(),
		Name:       req.Name,
		Title:      req.Title,
		Body:       req.Body,
		Channel:    req.Channel,
		Enabled:    enabled,
		CreatedBy:  actor,
		ModifiedAt: time.Now().UTC(),
	}
	h.Store.SaveNotificationTemplate(template)
	h.Logger.Info("admin_notification_template_created", "templateID", template.ID, "channel", template.Channel)
	return c.JSON(http.StatusCreated, template)
}

func (h *Handler) AdminSendNotification(c echo.Context) error {
	var req struct {
		UserID      string `json:"userId"`
		TemplateID  string `json:"templateId"`
		Title       string `json:"title"`
		Body        string `json:"body"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.TemplateID = strings.TrimSpace(req.TemplateID)
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.Fingerprint = strings.TrimSpace(req.Fingerprint)
	if req.UserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "userId is required"})
	}
	if _, ok := h.Store.GetUserByID(req.UserID); !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}

	channel := "in_app"
	if req.TemplateID != "" {
		template, ok := h.Store.GetNotificationTemplate(req.TemplateID)
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "template not found"})
		}
		if !template.Enabled {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "template is disabled"})
		}
		if req.Title == "" {
			req.Title = template.Title
		}
		if req.Body == "" {
			req.Body = template.Body
		}
		channel = template.Channel
	}
	if req.Title == "" || req.Body == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title and body are required"})
	}
	if req.Fingerprint == "" {
		sum := sha256.Sum256([]byte(req.UserID + "|" + req.TemplateID + "|" + req.Title + "|" + req.Body))
		req.Fingerprint = hex.EncodeToString(sum[:])
	}

	notification := models.Notification{
		ID:          uuid.NewString(),
		UserID:      req.UserID,
		TemplateID:  req.TemplateID,
		Title:       req.Title,
		Body:        req.Body,
		Status:      "delivered",
		Attempts:    1,
		Fingerprint: req.Fingerprint,
		DeliveredAt: time.Now().UTC(),
	}
	if channel == "email" || channel == "sms" {
		notification.Status = "disabled_offline"
		notification.DeliveredAt = time.Time{}
		// RULES: no 3rd-party integrations in offline mode.
		// Email/SMS delivery remains disabled for offline deployments.
	}
	h.Store.SaveNotification(notification)
	h.Logger.Info("admin_notification_sent", "notificationID", notification.ID, "userID", notification.UserID, "channel", channel)
	return c.JSON(http.StatusCreated, notification)
}

func (h *Handler) AdminRetryNotifications(c echo.Context) error {
	all := h.Store.ListAllNotifications()
	retried := 0
	delivered := 0
	deadLetter := 0
	for _, notification := range all {
		switch notification.Status {
		case "queued", "failed", "retry_pending", "disabled_offline":
		default:
			continue
		}
		retried++
		notification.Attempts++
		if notification.Status == "disabled_offline" {
			h.Store.SaveNotification(notification)
			continue
		}
		if notification.Attempts >= h.Cfg.NotificationRetryMax {
			notification.Status = "dead_letter"
			deadLetter++
			h.Store.SaveNotification(notification)
			continue
		}
		notification.Status = "delivered"
		notification.DeliveredAt = time.Now().UTC()
		delivered++
		h.Store.SaveNotification(notification)
	}
	return c.JSON(http.StatusOK, map[string]int{
		"retried":    retried,
		"delivered":  delivered,
		"deadLetter": deadLetter,
	})
}
