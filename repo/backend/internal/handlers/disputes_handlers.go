package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateComplaint(c echo.Context) error {
	var req struct {
		BookingID string `json:"bookingId"`
		Outcome   string `json:"outcome"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	booking, ok := h.Store.GetBooking(req.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !(hasRole(roles, models.RoleCustomer) || hasRole(roles, models.RoleProvider) || hasRole(roles, models.RoleAdmin)) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "role not allowed to open complaint"})
	}
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	item := models.Complaint{ID: uuid.NewString(), BookingID: req.BookingID, OpenedBy: actor, Status: "open", Outcome: req.Outcome, CreatedAt: time.Now().UTC()}
	h.Store.SaveComplaint(item)
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListComplaints(c echo.Context) error {
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	bookingFilter := strings.TrimSpace(c.QueryParam("bookingId"))
	all := h.Store.ListComplaints()
	out := make([]models.Complaint, 0, len(all))
	for _, item := range all {
		if bookingFilter != "" && item.BookingID != bookingFilter {
			continue
		}
		booking, ok := h.Store.GetBooking(item.BookingID)
		if !ok {
			continue
		}
		if canAccessBooking(actor, roles, booking) || item.OpenedBy == actor {
			out = append(out, item)
		}
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) ArbitrateComplaint(c echo.Context) error {
	roles, _ := c.Get("roles").([]models.Role)
	if !(hasRole(roles, models.RoleCSA) || hasRole(roles, models.RoleAdmin)) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "CSA/Admin required"})
	}
	var req struct {
		Status  string `json:"status"`
		Outcome string `json:"outcome"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	id := c.Param("id")
	complaint, ok := h.Store.GetComplaint(id)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "complaint not found"})
	}
	if req.Status != "" {
		complaint.Status = req.Status
	}
	complaint.Outcome = req.Outcome
	h.Store.SaveComplaint(complaint)
	return c.JSON(http.StatusOK, complaint)
}

func (h *Handler) CreateConsultation(c echo.Context) error {
	roles, _ := c.Get("roles").([]models.Role)
	if !(hasRole(roles, models.RoleCSA) || hasRole(roles, models.RoleAdmin)) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "CSA/Admin required"})
	}
	var req struct {
		BookingID      string `json:"bookingId"`
		Topic          string `json:"topic"`
		KeyPoints      string `json:"keyPoints"`
		Recommendation string `json:"recommendation"`
		FollowUp       string `json:"followUp"`
		Visibility     string `json:"visibility"`
		ChangeReason   string `json:"changeReason"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if strings.TrimSpace(req.Topic) == "" || strings.TrimSpace(req.BookingID) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bookingId and topic are required"})
	}
	booking, ok := h.Store.GetBooking(req.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	userID, _ := c.Get("userID").(string)
	if !canAccessBooking(userID, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	visibility := strings.ToLower(strings.TrimSpace(req.Visibility))
	if visibility == "" {
		visibility = "csa_admin"
	}
	switch visibility {
	case "csa_admin", "parties", "all":
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid visibility"})
	}
	threadID := req.BookingID + "::" + req.Topic
	versions := h.Store.ListConsultationsByThread(threadID)
	nextVersion := len(versions) + 1
	item := models.Consultation{
		ID:             uuid.NewString(),
		BookingID:      req.BookingID,
		Topic:          req.Topic,
		KeyPoints:      req.KeyPoints,
		Recommendation: req.Recommendation,
		FollowUp:       req.FollowUp,
		Visibility:     visibility,
		ChangeReason:   req.ChangeReason,
		Version:        nextVersion,
		CreatedBy:      userID,
		CreatedAt:      time.Now().UTC(),
	}
	h.Store.SaveConsultation(item)
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListConsultations(c echo.Context) error {
	bookingID := strings.TrimSpace(c.QueryParam("bookingId"))
	roles, _ := c.Get("roles").([]models.Role)
	actor, _ := c.Get("userID").(string)

	var out []models.Consultation

	// If bookingId provided, restrict to that booking (existing behaviour)
	if bookingID != "" {
		booking, ok := h.Store.GetBooking(bookingID)
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
		}
		if !canAccessBooking(actor, roles, booking) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		all := h.Store.ListConsultationsByBooking(bookingID)
		for _, item := range all {
			switch item.Visibility {
			case "all":
				out = append(out, item)
			case "parties":
				if hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) || actor == booking.CustomerID || actor == booking.ProviderID {
					out = append(out, item)
				}
			default:
				if hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) {
					out = append(out, item)
				}
			}
		}
		return c.JSON(http.StatusOK, out)
	}

	// No bookingId: return consultations the actor may access across bookings.
	// Admin/CSA: return all consultations; others: return consultations for bookings they are a party of (customer or provider).
	allBookings := h.Store.ListBookings()
	for _, booking := range allBookings {
		if !(hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) || actor == booking.CustomerID || actor == booking.ProviderID) {
			continue
		}
		consults := h.Store.ListConsultationsByBooking(booking.ID)
		for _, item := range consults {
			switch item.Visibility {
			case "all":
				out = append(out, item)
			case "parties":
				if hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) || actor == booking.CustomerID || actor == booking.ProviderID {
					out = append(out, item)
				}
			default:
				if hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) {
					out = append(out, item)
				}
			}
		}
	}

	return c.JSON(http.StatusOK, out)
}

func (h *Handler) AttachConsultationEvidence(c echo.Context) error {
	roles, _ := c.Get("roles").([]models.Role)
	if !(hasRole(roles, models.RoleCSA) || hasRole(roles, models.RoleAdmin)) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "CSA/Admin required"})
	}
	var req struct {
		ConsultationID string `json:"consultationId"`
		AttachmentID   string `json:"attachmentId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	consultation, ok := h.Store.GetConsultation(req.ConsultationID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "consultation not found"})
	}
	attachment, ok := h.Store.GetAttachment(req.AttachmentID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "attachment not found"})
	}
	if consultation.BookingID != "" && attachment.BookingID != consultation.BookingID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "attachment must belong to consultation booking"})
	}
	actor, _ := c.Get("userID").(string)
	item := models.ConsultationAttachment{
		ID:             uuid.NewString(),
		ConsultationID: req.ConsultationID,
		AttachmentID:   req.AttachmentID,
		CreatedBy:      actor,
		CreatedAt:      time.Now().UTC(),
	}
	h.Store.SaveConsultationAttachment(item)
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListConsultationEvidence(c echo.Context) error {
	consultationID := c.Param("id")
	consultation, ok := h.Store.GetConsultation(consultationID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "consultation not found"})
	}
	roles, _ := c.Get("roles").([]models.Role)
	actor, _ := c.Get("userID").(string)
	if consultation.BookingID != "" {
		booking, ok := h.Store.GetBooking(consultation.BookingID)
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
		}
		if !canAccessBooking(actor, roles, booking) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		switch consultation.Visibility {
		case "all":
			// allowed once booking access passes
		case "parties":
			if !(hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) || actor == booking.CustomerID || actor == booking.ProviderID) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden by consultation visibility"})
			}
		default:
			if !(hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA)) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden by consultation visibility"})
			}
		}
	}
	return c.JSON(http.StatusOK, h.Store.ListConsultationAttachments(consultationID))
}

func (h *Handler) CreateRating(c echo.Context) error {
	var req struct {
		BookingID string `json:"bookingId"`
		Score     int    `json:"score"`
		Comment   string `json:"comment"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if req.Score < 1 || req.Score > 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "score must be between 1 and 5"})
	}
	booking, ok := h.Store.GetBooking(req.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	target := booking.ProviderID
	if actor == booking.ProviderID {
		target = booking.CustomerID
	}
	item := models.Rating{
		ID:         uuid.NewString(),
		BookingID:  req.BookingID,
		FromUserID: actor,
		ToUserID:   target,
		Score:      req.Score,
		Comment:    req.Comment,
		CreatedAt:  time.Now().UTC(),
	}
	h.Store.SaveRating(item)
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListRatings(c echo.Context) error {
	bookingID := c.QueryParam("bookingId")
	if bookingID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bookingId is required"})
	}
	booking, ok := h.Store.GetBooking(bookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	return c.JSON(http.StatusOK, h.Store.ListRatings(bookingID))
}

func (h *Handler) ExportDisputePDF(c echo.Context) error {
	roles, _ := c.Get("roles").([]models.Role)
	if !(hasRole(roles, models.RoleCSA) || hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCustomer) || hasRole(roles, models.RoleProvider)) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "role not allowed"})
	}
	id := c.Param("id")
	complaint, ok := h.Store.GetComplaint(id)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "complaint not found"})
	}
	booking, ok := h.Store.GetBooking(complaint.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	if !canAccessBooking(actor, roles, booking) && !hasRole(roles, models.RoleCSA) && !hasRole(roles, models.RoleAdmin) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	inspections := h.Store.ListInspections(booking.ID)
	inspectionRows := make([]string, 0, len(inspections))
	for _, rev := range inspections {
		inspectionRows = append(inspectionRows, fmt.Sprintf("revision=%s stage=%s createdAt=%s prevHash=%s hash=%s", rev.RevisionID, rev.Stage, rev.CreatedAt.UTC().Format(time.RFC3339), rev.PrevHash, rev.Hash))
	}
	entries := h.Store.ListLedger(booking.ID)
	ledgerRows := make([]string, 0, len(entries))
	for _, entry := range entries {
		ledgerRows = append(ledgerRows, fmt.Sprintf("entry=%s type=%s amount=%.2f createdAt=%s prevHash=%s hash=%s", entry.ID, entry.Type, entry.Amount, entry.CreatedAt.UTC().Format(time.RFC3339), entry.PrevHash, entry.Hash))
	}

	pdfBytes, err := services.GenerateDisputePDF(services.DisputePDFData{
		ComplaintID:    complaint.ID,
		BookingID:      booking.ID,
		Status:         complaint.Status,
		Outcome:        complaint.Outcome,
		OpenedBy:       complaint.OpenedBy,
		GeneratedAt:    time.Now().UTC(),
		InspectionRows: inspectionRows,
		LedgerRows:     ledgerRows,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate PDF"})
	}
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) Notifications(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	return c.JSON(http.StatusOK, h.Store.ListNotifications(userID))
}

func (h *Handler) SyncReconcile(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "reconciled"})
}
