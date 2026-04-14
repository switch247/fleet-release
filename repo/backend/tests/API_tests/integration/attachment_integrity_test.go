package integration

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAttachmentIntegrityValidationAndDedup(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	// Oversized photo (>10 MB) should be rejected with 400.
	resp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bID, "type": "photo",
		"sizeBytes": 11 * 1024 * 1024, "checksum": "abc", "fingerprint": "fp-large",
	}, custToken)
	intMustStatus(t, resp, http.StatusBadRequest)

	// Valid init.
	fp := fmt.Sprintf("fp-integrity-%d", time.Now().UnixNano())
	initResp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bID, "type": "photo",
		"sizeBytes": 10, "checksum": "deadbeef", "fingerprint": fp,
	}, custToken)
	initBody := intMustStatus(t, initResp, http.StatusCreated)
	var initOut struct{ UploadID string `json:"uploadId"` }
	_ = json.Unmarshal(initBody, &initOut)

	// Upload chunk.
	chunkResp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/upload", map[string]interface{}{
		"uploadId": initOut.UploadID, "chunkBase64": base64.StdEncoding.EncodeToString([]byte("hello")),
	}, custToken)
	intMustStatus(t, chunkResp, http.StatusOK)

	// Complete with wrong checksum → 400.
	completeResp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/complete",
		map[string]string{"uploadId": initOut.UploadID}, custToken)
	intMustStatus(t, completeResp, http.StatusBadRequest)
}

func TestAttachmentFingerprintDedupScopedByBooking(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	bID1 := intCreateBooking(t, custToken)
	bID2 := intCreateBooking(t, custToken)

	sharedFP := fmt.Sprintf("shared-fp-%d", time.Now().UnixNano())
	sum := sha256.Sum256(intMiniPNG)
	checksum := hex.EncodeToString(sum[:])

	// Init attachment on booking 1.
	r1 := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bID1, "type": "photo",
		"sizeBytes": len(intMiniPNG), "checksum": checksum, "fingerprint": sharedFP,
	}, custToken)
	intMustStatus(t, r1, http.StatusCreated)

	// Attempting the same fingerprint on booking 2 → 409 Conflict.
	r2 := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bID2, "type": "photo",
		"sizeBytes": len(intMiniPNG), "checksum": checksum, "fingerprint": sharedFP,
	}, custToken)
	intMustStatus(t, r2, http.StatusConflict)
}
