package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestAttachmentIntegrityValidationAndDedup(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")

	largeInitBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   h.BookingID,
		"type":        "photo",
		"sizeBytes":   11 * 1024 * 1024,
		"checksum":    "abc",
		"fingerprint": "fp-large",
	})
	largeReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(largeInitBody))
	largeReq.Header.Set("Content-Type", "application/json")
	largeReq.Header.Set("Authorization", "Bearer "+token)
	largeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(largeRec, largeReq)
	if largeRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for large photo, got %d body=%s", largeRec.Code, largeRec.Body.String())
	}

	initBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   h.BookingID,
		"type":        "photo",
		"sizeBytes":   10,
		"checksum":    "deadbeef",
		"fingerprint": "fp-a",
	})
	initReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(initBody))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Authorization", "Bearer "+token)
	initRec := httptest.NewRecorder()
	h.Router.ServeHTTP(initRec, initReq)
	if initRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 init got %d body=%s", initRec.Code, initRec.Body.String())
	}
	var initResp struct {
		UploadID string `json:"uploadId"`
	}
	_ = json.Unmarshal(initRec.Body.Bytes(), &initResp)

	chunkPayload := base64.StdEncoding.EncodeToString([]byte("hello"))
	chunkBody, _ := json.Marshal(map[string]string{"uploadId": initResp.UploadID, "chunkBase64": chunkPayload})
	chunkReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/upload", bytes.NewReader(chunkBody))
	chunkReq.Header.Set("Content-Type", "application/json")
	chunkReq.Header.Set("Authorization", "Bearer "+token)
	chunkRec := httptest.NewRecorder()
	h.Router.ServeHTTP(chunkRec, chunkReq)
	if chunkRec.Code != http.StatusOK {
		t.Fatalf("expected 200 chunk got %d body=%s", chunkRec.Code, chunkRec.Body.String())
	}

	completeBody, _ := json.Marshal(map[string]string{"uploadId": initResp.UploadID})
	completeReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/complete", bytes.NewReader(completeBody))
	completeReq.Header.Set("Content-Type", "application/json")
	completeReq.Header.Set("Authorization", "Bearer "+token)
	completeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 checksum mismatch got %d body=%s", completeRec.Code, completeRec.Body.String())
	}
}

func TestAttachmentFingerprintDedupScopedByBooking(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")

	bookingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	bookingsReq.Header.Set("Authorization", "Bearer "+token)
	bookingsRec := httptest.NewRecorder()
	h.Router.ServeHTTP(bookingsRec, bookingsReq)
	if bookingsRec.Code != http.StatusOK {
		t.Fatalf("list bookings failed status=%d body=%s", bookingsRec.Code, bookingsRec.Body.String())
	}
	var existingBookings []struct {
		ListingID string `json:"listingId"`
	}
	_ = json.Unmarshal(bookingsRec.Body.Bytes(), &existingBookings)
	if len(existingBookings) == 0 || existingBookings[0].ListingID == "" {
		t.Fatalf("expected seeded booking with listing id")
	}

	createBody, _ := json.Marshal(map[string]interface{}{
		"listingId": existingBookings[0].ListingID,
		"startAt":   "2026-03-28T09:00:00Z",
		"endAt":     "2026-03-28T11:00:00Z",
		"odoStart":  40,
		"odoEnd":    55,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	h.Router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create second booking failed status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		Booking struct {
			ID string `json:"id"`
		} `json:"booking"`
	}
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	if created.Booking.ID == "" {
		t.Fatalf("expected created booking id")
	}

	firstInitBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   h.BookingID,
		"type":        "photo",
		"sizeBytes":   100,
		"checksum":    "deadbeef",
		"fingerprint": "shared-fingerprint",
	})
	firstInitReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(firstInitBody))
	firstInitReq.Header.Set("Content-Type", "application/json")
	firstInitReq.Header.Set("Authorization", "Bearer "+token)
	firstInitRec := httptest.NewRecorder()
	h.Router.ServeHTTP(firstInitRec, firstInitReq)
	if firstInitRec.Code != http.StatusCreated {
		t.Fatalf("first attachment init failed status=%d body=%s", firstInitRec.Code, firstInitRec.Body.String())
	}

	secondInitBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   created.Booking.ID,
		"type":        "photo",
		"sizeBytes":   100,
		"checksum":    "deadbeef",
		"fingerprint": "shared-fingerprint",
	})
	secondInitReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(secondInitBody))
	secondInitReq.Header.Set("Content-Type", "application/json")
	secondInitReq.Header.Set("Authorization", "Bearer "+token)
	secondInitRec := httptest.NewRecorder()
	h.Router.ServeHTTP(secondInitRec, secondInitReq)
	if secondInitRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for cross-booking fingerprint reuse, got %d body=%s", secondInitRec.Code, secondInitRec.Body.String())
	}
}
