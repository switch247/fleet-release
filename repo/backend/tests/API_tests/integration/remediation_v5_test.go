package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func createBookingForToken(t *testing.T, router http.Handler, token, listingID, startAt, endAt string, odoStart, odoEnd float64) string {
	t.Helper()
	payload, _ := json.Marshal(map[string]interface{}{
		"listingId": listingID,
		"startAt":   startAt,
		"endAt":     endAt,
		"odoStart":  odoStart,
		"odoEnd":    odoEnd,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create booking failed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Booking struct {
			ID string `json:"id"`
		} `json:"booking"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Booking.ID == "" {
		t.Fatalf("expected created booking id")
	}
	return body.Booking.ID
}

func uploadAttachment(t *testing.T, router http.Handler, token, bookingID, fileType, fingerprint string, bytesIn []byte) string {
	t.Helper()
	sum := sha256.Sum256(bytesIn)
	checksum := hex.EncodeToString(sum[:])

	initBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   bookingID,
		"type":        fileType,
		"sizeBytes":   len(bytesIn),
		"checksum":    checksum,
		"fingerprint": fingerprint,
	})
	initReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(initBody))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Authorization", "Bearer "+token)
	initRec := httptest.NewRecorder()
	router.ServeHTTP(initRec, initReq)
	if initRec.Code != http.StatusCreated {
		t.Fatalf("attachment init failed status=%d body=%s", initRec.Code, initRec.Body.String())
	}
	var initResp struct {
		UploadID string `json:"uploadId"`
	}
	_ = json.Unmarshal(initRec.Body.Bytes(), &initResp)

	chunkBody, _ := json.Marshal(map[string]string{
		"uploadId":    initResp.UploadID,
		"chunkBase64": base64.StdEncoding.EncodeToString(bytesIn),
	})
	chunkReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/upload", bytes.NewReader(chunkBody))
	chunkReq.Header.Set("Content-Type", "application/json")
	chunkReq.Header.Set("Authorization", "Bearer "+token)
	chunkRec := httptest.NewRecorder()
	router.ServeHTTP(chunkRec, chunkReq)
	if chunkRec.Code != http.StatusOK {
		t.Fatalf("attachment chunk failed status=%d body=%s", chunkRec.Code, chunkRec.Body.String())
	}

	completeBody, _ := json.Marshal(map[string]string{"uploadId": initResp.UploadID})
	completeReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/complete", bytes.NewReader(completeBody))
	completeReq.Header.Set("Content-Type", "application/json")
	completeReq.Header.Set("Authorization", "Bearer "+token)
	completeRec := httptest.NewRecorder()
	router.ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("attachment complete failed status=%d body=%s", completeRec.Code, completeRec.Body.String())
	}

	return initResp.UploadID
}

func TestInspectionRejectsCrossBookingEvidenceID(t *testing.T) {
	h := public.BuildHarnessForTests()
	customerToken := loginToken(t, h.Router, "customer", "Customer1234!")
	providerToken := loginToken(t, h.Router, "provider", "Provider1234!")

	bookingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	bookingsReq.Header.Set("Authorization", "Bearer "+customerToken)
	bookingsRec := httptest.NewRecorder()
	h.Router.ServeHTTP(bookingsRec, bookingsReq)
	if bookingsRec.Code != http.StatusOK {
		t.Fatalf("list bookings failed status=%d body=%s", bookingsRec.Code, bookingsRec.Body.String())
	}
	var seeded []struct {
		ListingID string `json:"listingId"`
	}
	_ = json.Unmarshal(bookingsRec.Body.Bytes(), &seeded)

	secondBookingID := createBookingForToken(t, h.Router, customerToken, seeded[0].ListingID, "2026-04-02T09:00:00Z", "2026-04-02T11:00:00Z", 50, 62)
	pngBytes, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8Xw8AAoMBgR5nYeEAAAAASUVORK5CYII=")
	crossEvidenceID := uploadAttachment(t, h.Router, customerToken, secondBookingID, "photo", "cross-booking-evidence", pngBytes)

	inspectionBody, _ := json.Marshal(map[string]interface{}{
		"bookingId": h.BookingID,
		"stage":     "return",
		"items": []map[string]interface{}{
			{"name": "Exterior bodywork", "condition": "good", "evidenceIds": []string{crossEvidenceID}},
		},
		"notes": "cross-booking evidence should fail",
	})
	inspectionReq := httptest.NewRequest(http.MethodPost, "/api/v1/inspections", bytes.NewReader(inspectionBody))
	inspectionReq.Header.Set("Content-Type", "application/json")
	inspectionReq.Header.Set("Authorization", "Bearer "+providerToken)
	inspectionRec := httptest.NewRecorder()
	h.Router.ServeHTTP(inspectionRec, inspectionReq)
	if inspectionRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for cross-booking evidence got %d body=%s", inspectionRec.Code, inspectionRec.Body.String())
	}
}

func TestConsultationVersionIsolationByBookingThread(t *testing.T) {
	h := public.BuildHarnessForTests()
	customerToken := loginToken(t, h.Router, "customer", "Customer1234!")
	agentToken := loginToken(t, h.Router, "agent", "Agent1234!Pass")

	bookingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	bookingsReq.Header.Set("Authorization", "Bearer "+customerToken)
	bookingsRec := httptest.NewRecorder()
	h.Router.ServeHTTP(bookingsRec, bookingsReq)
	if bookingsRec.Code != http.StatusOK {
		t.Fatalf("list bookings failed status=%d body=%s", bookingsRec.Code, bookingsRec.Body.String())
	}
	var seeded []struct {
		ListingID string `json:"listingId"`
	}
	_ = json.Unmarshal(bookingsRec.Body.Bytes(), &seeded)

	secondBookingID := createBookingForToken(t, h.Router, customerToken, seeded[0].ListingID, "2026-04-03T09:00:00Z", "2026-04-03T11:00:00Z", 65, 75)

	createConsult := func(bookingID string) int {
		payload, _ := json.Marshal(map[string]string{
			"bookingId":      bookingID,
			"topic":          "damage review",
			"keyPoints":      "inspection notes",
			"recommendation": "follow process",
			"followUp":       "call both parties",
			"changeReason":   "inspection_feedback",
			"visibility":     "parties",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/consultations", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+agentToken)
		rec := httptest.NewRecorder()
		h.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create consultation failed status=%d body=%s", rec.Code, rec.Body.String())
		}
		var body struct {
			Version int `json:"version"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return body.Version
	}

	v1a := createConsult(h.BookingID)
	v2a := createConsult(h.BookingID)
	v1b := createConsult(secondBookingID)
	v2b := createConsult(secondBookingID)

	if v1a != 1 || v2a != 2 || v1b != 1 || v2b != 2 {
		t.Fatalf("expected per-thread versioning [1,2] and [1,2], got [%d,%d] and [%d,%d]", v1a, v2a, v1b, v2b)
	}
}

func TestAttachmentMIMEBypassReturnsUnsupportedMediaType(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")
	payloadBytes := []byte("plain text pretending to be photo")
	sum := sha256.Sum256(payloadBytes)
	checksum := hex.EncodeToString(sum[:])

	initBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   h.BookingID,
		"type":        "photo",
		"sizeBytes":   len(payloadBytes),
		"checksum":    checksum,
		"fingerprint": "mime-bypass-text-file",
	})
	initReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(initBody))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Authorization", "Bearer "+token)
	initRec := httptest.NewRecorder()
	h.Router.ServeHTTP(initRec, initReq)
	if initRec.Code != http.StatusCreated {
		t.Fatalf("attachment init failed status=%d body=%s", initRec.Code, initRec.Body.String())
	}
	var initResp struct {
		UploadID string `json:"uploadId"`
	}
	_ = json.Unmarshal(initRec.Body.Bytes(), &initResp)

	chunkBody, _ := json.Marshal(map[string]string{
		"uploadId":    initResp.UploadID,
		"chunkBase64": base64.StdEncoding.EncodeToString(payloadBytes),
	})
	chunkReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/upload", bytes.NewReader(chunkBody))
	chunkReq.Header.Set("Content-Type", "application/json")
	chunkReq.Header.Set("Authorization", "Bearer "+token)
	chunkRec := httptest.NewRecorder()
	h.Router.ServeHTTP(chunkRec, chunkReq)
	if chunkRec.Code != http.StatusOK {
		t.Fatalf("attachment chunk failed status=%d body=%s", chunkRec.Code, chunkRec.Body.String())
	}

	completeBody, _ := json.Marshal(map[string]string{"uploadId": initResp.UploadID})
	completeReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/complete", bytes.NewReader(completeBody))
	completeReq.Header.Set("Content-Type", "application/json")
	completeReq.Header.Set("Authorization", "Bearer "+token)
	completeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 for MIME bypass got %d body=%s", completeRec.Code, completeRec.Body.String())
	}
}
