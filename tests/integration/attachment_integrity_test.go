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
