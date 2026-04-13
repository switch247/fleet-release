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

// TestInspectionRejectsCrossBookingEvidenceID verifies that an evidence
// attachment belonging to booking B cannot be referenced in booking A's
// inspection.
func TestInspectionRejectsCrossBookingEvidenceID(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	provToken := intLogin(t, intProvUser, intProvPass)

	bID1 := intCreateBooking(t, custToken) // booking the provider owns
	bID2 := intCreateBooking(t, custToken) // second booking with cross evidence

	// Upload evidence tied to bID2.
	crossEvidID := intUploadAttachment(t, custToken, bID2,
		fmt.Sprintf("cross-evid-%d", time.Now().UnixNano()), intMiniPNG)

	// Attempt to use bID2's evidence in bID1's inspection → 400.
	inspBody := map[string]interface{}{
		"bookingId": bID1, "stage": "return",
		"items": []map[string]interface{}{
			{"name": "Exterior", "condition": "good", "evidenceIds": []string{crossEvidID}},
		},
		"notes": "cross-booking evidence should fail",
	}
	inspResp := intAPI(t, http.MethodPost, "/api/v1/inspections", inspBody, provToken)
	intMustStatus(t, inspResp, http.StatusBadRequest)
}

// TestConsultationVersionIsolationByBookingThread verifies that consultation
// versions are tracked independently per booking+topic thread.
func TestConsultationVersionIsolationByBookingThread(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	agentToken := intLogin(t, intAgentUser, intAgentPass)

	bID1 := intCreateBooking(t, custToken)
	bID2 := intCreateBooking(t, custToken)

	createConsult := func(bookingID string) int {
		resp := intAPI(t, http.MethodPost, "/api/v1/consultations", map[string]string{
			"bookingId": bookingID, "topic": "damage review",
			"keyPoints": "notes", "recommendation": "follow process",
			"followUp": "call both parties", "changeReason": "inspection_feedback",
			"visibility": "parties",
		}, agentToken)
		body := intMustStatus(t, resp, http.StatusCreated)
		var out struct{ Version int `json:"version"` }
		_ = json.Unmarshal(body, &out)
		return out.Version
	}

	v1a := createConsult(bID1)
	v2a := createConsult(bID1)
	v1b := createConsult(bID2)
	v2b := createConsult(bID2)

	if v1a != 1 || v2a != 2 || v1b != 1 || v2b != 2 {
		t.Fatalf("expected per-thread versions [1,2] and [1,2], got [%d,%d] and [%d,%d]",
			v1a, v2a, v1b, v2b)
	}
}

// TestAttachmentMIMEBypassReturnsUnsupportedMediaType verifies that uploading
// a plain-text file as "photo" type is rejected at completion with 415.
func TestAttachmentMIMEBypassReturnsUnsupportedMediaType(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	payload := []byte("plain text pretending to be a photo")
	sum := sha256.Sum256(payload)
	checksum := hex.EncodeToString(sum[:])

	fp := fmt.Sprintf("mime-bypass-%d", time.Now().UnixNano())
	initResp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/init", map[string]interface{}{
		"bookingId": bID, "type": "photo",
		"sizeBytes": len(payload), "checksum": checksum, "fingerprint": fp,
	}, custToken)
	initBody := intMustStatus(t, initResp, http.StatusCreated)
	var initOut struct{ UploadID string `json:"uploadId"` }
	_ = json.Unmarshal(initBody, &initOut)

	chunkResp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/upload", map[string]interface{}{
		"uploadId": initOut.UploadID, "chunkBase64": base64.StdEncoding.EncodeToString(payload),
	}, custToken)
	intMustStatus(t, chunkResp, http.StatusOK)

	// Complete — MIME check fails, expect 415 Unsupported Media Type.
	completeResp := intAPI(t, http.MethodPost, "/api/v1/attachments/chunk/complete",
		map[string]string{"uploadId": initOut.UploadID}, custToken)
	intMustStatus(t, completeResp, http.StatusUnsupportedMediaType)
}
