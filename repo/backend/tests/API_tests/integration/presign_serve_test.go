package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"fleetlease/backend/pkg/public"
)

// TestPresignAndServe uploads a small payload, requests a presigned URL, then
// fetches the public URL and verifies headers and body.
func TestPresignAndServe(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")
	pngBytes, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8Xw8AAoMBgR5nYeEAAAAASUVORK5CYII=")

	// Init
	initBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   h.BookingID,
		"type":        "photo",
		"sizeBytes":   len(pngBytes),
		"checksum":    "",
		"fingerprint": "fp-presign",
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

	// Upload chunk
	chunkPayload := base64.StdEncoding.EncodeToString(pngBytes)
	chunkBody, _ := json.Marshal(map[string]string{"uploadId": initResp.UploadID, "chunkBase64": chunkPayload})
	chunkReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/upload", bytes.NewReader(chunkBody))
	chunkReq.Header.Set("Content-Type", "application/json")
	chunkReq.Header.Set("Authorization", "Bearer "+token)
	chunkRec := httptest.NewRecorder()
	h.Router.ServeHTTP(chunkRec, chunkReq)
	if chunkRec.Code != http.StatusOK {
		t.Fatalf("expected 200 chunk got %d body=%s", chunkRec.Code, chunkRec.Body.String())
	}

	// Complete
	completeBody, _ := json.Marshal(map[string]string{"uploadId": initResp.UploadID})
	completeReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/complete", bytes.NewReader(completeBody))
	completeReq.Header.Set("Content-Type", "application/json")
	completeReq.Header.Set("Authorization", "Bearer "+token)
	completeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("expected 200 complete got %d body=%s", completeRec.Code, completeRec.Body.String())
	}

	// Presign
	presignBody, _ := json.Marshal(map[string]int{"ttlSeconds": 60})
	presignReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/"+initResp.UploadID+"/presign", bytes.NewReader(presignBody))
	presignReq.Header.Set("Content-Type", "application/json")
	presignReq.Header.Set("Authorization", "Bearer "+token)
	presignRec := httptest.NewRecorder()
	h.Router.ServeHTTP(presignRec, presignReq)
	if presignRec.Code != http.StatusOK {
		t.Fatalf("expected 200 presign got %d body=%s", presignRec.Code, presignRec.Body.String())
	}
	var presignResp struct {
		Url string `json:"url"`
	}
	_ = json.Unmarshal(presignRec.Body.Bytes(), &presignResp)

	// Parse returned URL and call router with path+query
	u, err := url.Parse(presignResp.Url)
	if err != nil {
		t.Fatalf("invalid presign url: %v", err)
	}
	servePath := u.Path + "?" + u.RawQuery
	serveReq := httptest.NewRequest(http.MethodGet, servePath, nil)
	// include Host header matching url so handlers can build scheme/host if needed
	serveReq.Host = u.Host
	serveRec := httptest.NewRecorder()
	h.Router.ServeHTTP(serveRec, serveReq)
	if serveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 serve got %d body=%s", serveRec.Code, serveRec.Body.String())
	}
	// verify content-disposition and body
	cd := serveRec.Header().Get("Content-Disposition")
	if cd == "" {
		t.Fatalf("missing Content-Disposition header")
	}
	if !bytes.Equal(serveRec.Body.Bytes(), pngBytes) {
		t.Fatalf("unexpected body bytes length=%d", len(serveRec.Body.Bytes()))
	}
}
