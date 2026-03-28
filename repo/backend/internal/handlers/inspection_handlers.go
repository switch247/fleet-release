package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) UpsertInspection(c echo.Context) error {
	var req struct {
		BookingID string                  `json:"bookingId"`
		Stage     string                  `json:"stage"`
		Items     []models.InspectionItem `json:"items"`
		Notes     string                  `json:"notes"`
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
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "inspection items required"})
	}
	for _, item := range req.Items {
		if len(item.EvidenceIDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "camera evidence required per checklist item"})
		}
	}
	prevHash := ""
	revisions := h.Store.ListInspections(req.BookingID)
	if len(revisions) > 0 {
		prevHash = revisions[len(revisions)-1].Hash
	}
	now := time.Now().UTC()
	payload, _ := json.Marshal(req)
	hash := services.ChainHash(prevHash, string(payload), now)
	rev := models.InspectionRevision{RevisionID: uuid.NewString(), BookingID: req.BookingID, Stage: req.Stage, Items: req.Items, Notes: req.Notes, CreatedBy: actor, CreatedAt: now, PrevHash: prevHash, Hash: hash}
	h.Store.SaveInspection(req.BookingID, rev)
	return c.JSON(http.StatusCreated, rev)
}

func (h *Handler) AttachmentInit(c echo.Context) error {
	var req struct {
		BookingID   string `json:"bookingId"`
		Type        string `json:"type"`
		SizeBytes   int64  `json:"sizeBytes"`
		Checksum    string `json:"checksum"`
		Fingerprint string `json:"fingerprint"`
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
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if req.Type == "photo" && req.SizeBytes > 10*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "photo size exceeds 10MB"})
	}
	if req.Type == "video" && req.SizeBytes > 100*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "video size exceeds 100MB"})
	}
	if existing, ok := h.Store.FindAttachmentByFingerprint(req.Fingerprint); ok {
		return c.JSON(http.StatusOK, map[string]interface{}{"deduplicated": true, "attachment": existing})
	}
	_ = os.MkdirAll(h.Cfg.AttachmentDir, 0o755)
	uploadID := uuid.NewString()
	path := filepath.Join(h.Cfg.AttachmentDir, uploadID+".part")
	att := models.Attachment{ID: uploadID, BookingID: req.BookingID, Type: req.Type, SizeBytes: req.SizeBytes, Checksum: req.Checksum, Fingerprint: req.Fingerprint, Path: path}
	h.Store.SaveAttachment(att)
	return c.JSON(http.StatusCreated, map[string]interface{}{"uploadId": att.ID, "path": att.Path, "checksum": att.Checksum})
}

func (h *Handler) AttachmentChunk(c echo.Context) error {
	var req struct {
		UploadID    string `json:"uploadId"`
		ChunkBase64 string `json:"chunkBase64"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	att, ok := h.Store.GetAttachment(req.UploadID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "upload not found"})
	}
	booking, ok := h.Store.GetBooking(att.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	chunk, err := base64.StdEncoding.DecodeString(req.ChunkBase64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid chunk encoding"})
	}
	f, err := os.OpenFile(att.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store chunk"})
	}
	defer f.Close()
	if _, err = f.Write(chunk); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to write chunk"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "chunk accepted"})
}

func (h *Handler) AttachmentComplete(c echo.Context) error {
	var req struct {
		UploadID string `json:"uploadId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	att, ok := h.Store.GetAttachment(req.UploadID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "upload not found"})
	}
	booking, ok := h.Store.GetBooking(att.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	bytes, err := os.ReadFile(att.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "upload file missing"})
	}
	hash := sha256.Sum256(bytes)
	computed := hex.EncodeToString(hash[:])
	if att.Checksum != "" && !strings.EqualFold(att.Checksum, computed) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "checksum mismatch"})
	}
	finalPath := filepath.Join(h.Cfg.AttachmentDir, req.UploadID)
	if err = os.Rename(att.Path, finalPath); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to finalize upload"})
	}
	att.Path = finalPath
	h.Store.SaveAttachment(att)
	return c.JSON(http.StatusOK, map[string]string{"status": "upload complete"})
}

func (h *Handler) VerifyInspection(c echo.Context) error {
	bookingID := c.Param("bookingID")
	booking, ok := h.Store.GetBooking(bookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	revisions := h.Store.ListInspections(bookingID)
	type verifiedItem struct {
		Name       string   `json:"name"`
		Verified   bool     `json:"verified"`
		EvidenceID []string `json:"evidenceIds"`
	}
	type verifiedRevision struct {
		RevisionID string         `json:"revisionId"`
		Stage      string         `json:"stage"`
		Hash       string         `json:"hash"`
		PrevHash   string         `json:"prevHash"`
		HashValid  bool           `json:"hashValid"`
		Items      []verifiedItem `json:"items"`
	}
	resp := make([]verifiedRevision, 0, len(revisions))
	prev := ""
	for _, rev := range revisions {
		payload := struct {
			BookingID string                  `json:"bookingId"`
			Stage     string                  `json:"stage"`
			Items     []models.InspectionItem `json:"items"`
			Notes     string                  `json:"notes"`
		}{BookingID: rev.BookingID, Stage: rev.Stage, Items: rev.Items, Notes: rev.Notes}
		body, _ := json.Marshal(payload)
		expected := services.ChainHash(prev, string(body), rev.CreatedAt)
		revisionResult := verifiedRevision{
			RevisionID: rev.RevisionID,
			Stage:      rev.Stage,
			Hash:       rev.Hash,
			PrevHash:   rev.PrevHash,
			HashValid:  rev.PrevHash == prev && strings.EqualFold(rev.Hash, expected),
			Items:      make([]verifiedItem, 0, len(rev.Items)),
		}
		for _, item := range rev.Items {
			itemValid := true
			for _, evidenceID := range item.EvidenceIDs {
				att, ok := h.Store.GetAttachment(evidenceID)
				if !ok {
					itemValid = false
					break
				}
				fileBytes, err := os.ReadFile(att.Path)
				if err != nil {
					itemValid = false
					break
				}
				sum := sha256.Sum256(fileBytes)
				computed := hex.EncodeToString(sum[:])
				if att.Checksum != "" && !strings.EqualFold(att.Checksum, computed) {
					itemValid = false
					break
				}
			}
			revisionResult.Items = append(revisionResult.Items, verifiedItem{Name: item.Name, Verified: itemValid, EvidenceID: item.EvidenceIDs})
		}
		resp = append(resp, revisionResult)
		prev = rev.Hash
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"bookingId": bookingID, "revisions": resp})
}

func (h *Handler) ListInspections(c echo.Context) error {
	bookingID := c.QueryParam("bookingId")
	if bookingID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bookingId is required"})
	}
	booking, ok := h.Store.GetBooking(bookingID)
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)

	// If booking doesn't exist, return 404 (booking not found)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}

	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	revisions := h.Store.ListInspections(bookingID)
	if revisions == nil {
		revisions = make([]models.InspectionRevision, 0)
	}
	return c.JSON(http.StatusOK, revisions)
}

// AttachmentPresign returns a temporary signed URL for previewing an attachment.
func (h *Handler) AttachmentPresign(c echo.Context) error {
	id := c.Param("id")
	// require authenticated user and access to booking
	attachment, ok := h.Store.GetAttachment(id)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "attachment not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if attachment.BookingID != "" {
		booking, ok := h.Store.GetBooking(attachment.BookingID)
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
		}
		if !canAccessBooking(actor, roles, booking) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
	}
	// optional ttl in body
	var req struct {
		TTLSeconds int `json:"ttlSeconds"`
	}
	_ = c.Bind(&req)
	ttl := 60
	if req.TTLSeconds > 0 && req.TTLSeconds <= 3600 {
		ttl = req.TTLSeconds
	}
	exp := time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	msg := fmt.Sprintf("%s:%d", id, exp)
	mac := hmac.New(sha256.New, []byte(h.Cfg.AttachmentSigningSecret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	// Build fully-qualified URL using request scheme+host so frontend can open it directly.
	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	} else if p := c.Request().Header.Get("X-Forwarded-Proto"); p != "" {
		scheme = p
	}
	host := c.Request().Host
	url := fmt.Sprintf("%s://%s/api/v1/attachments/%s?exp=%d&sig=%s", scheme, host, id, exp, sig)
	return c.JSON(http.StatusOK, map[string]string{"url": url})
}

// AttachmentServe serves an attachment file when presented with a valid signature and expiry.
func (h *Handler) AttachmentServe(c echo.Context) error {
	id := c.Param("id")
	expStr := c.QueryParam("exp")
	sig := c.QueryParam("sig")
	if expStr == "" || sig == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing signature or expiry"})
	}
	expInt, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid expiry"})
	}
	if time.Now().Unix() > expInt {
		return c.JSON(http.StatusGone, map[string]string{"error": "url expired"})
	}
	msg := fmt.Sprintf("%s:%d", id, expInt)
	mac := hmac.New(sha256.New, []byte(h.Cfg.AttachmentSigningSecret))
	mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "invalid signature"})
	}
	att, ok := h.Store.GetAttachment(id)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "attachment not found"})
	}
	// read file and detect content-type
	bytes, err := os.ReadFile(att.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to read file"})
	}
	contentType := "application/octet-stream"
	if len(bytes) > 0 {
		contentType = http.DetectContentType(bytes[:512])
	}
	// set inline disposition so browser renders images/videos directly
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", id))
	c.Response().Header().Set("Cache-Control", "private, max-age=60")
	return c.Blob(http.StatusOK, contentType, bytes)
}
