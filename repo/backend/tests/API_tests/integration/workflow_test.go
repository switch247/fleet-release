package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestConsultationVisibilityForCustomer verifies that a csa_admin consultation
// is hidden when a customer lists consultations for the same booking.
func TestConsultationVisibilityForCustomer(t *testing.T) {
	skipIfNoIntLive(t)

	agentToken := intLogin(t, intAgentUser, intAgentPass)
	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	// Agent creates an internal (csa_admin) consultation.
	createResp := intAPI(t, http.MethodPost, "/api/v1/consultations", map[string]string{
		"bookingId": bID, "topic": "damage review",
		"keyPoints": "mirror crack", "recommendation": "collect proof",
		"followUp": "call both parties", "visibility": "csa_admin",
	}, agentToken)
	intMustStatus(t, createResp, http.StatusCreated)

	// Customer lists consultations — should see zero records.
	listResp := intAPI(t, http.MethodGet, "/api/v1/consultations?bookingId="+bID, nil, custToken)
	listBody := intMustStatus(t, listResp, http.StatusOK)
	var consults []map[string]interface{}
	_ = json.Unmarshal(listBody, &consults)
	if len(consults) != 0 {
		t.Fatalf("expected csa_admin consultation hidden from customer, got %d rows", len(consults))
	}
}

// TestConsultationAttachmentVisibilityForCustomer verifies that a customer
// cannot list attachments of a csa_admin consultation (403).
func TestConsultationAttachmentVisibilityForCustomer(t *testing.T) {
	skipIfNoIntLive(t)

	agentToken := intLogin(t, intAgentUser, intAgentPass)
	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	// Agent creates csa_admin consultation.
	createResp := intAPI(t, http.MethodPost, "/api/v1/consultations", map[string]string{
		"bookingId": bID, "topic": "escalation note",
		"keyPoints": "internal", "recommendation": "await supervisor",
		"followUp": "audit", "visibility": "csa_admin",
	}, agentToken)
	createBody := intMustStatus(t, createResp, http.StatusCreated)
	var consult struct{ ID string `json:"id"` }
	_ = json.Unmarshal(createBody, &consult)

	// Customer tries to list its attachments → 403.
	attResp := intAPI(t, http.MethodGet,
		"/api/v1/consultations/"+consult.ID+"/attachments", nil, custToken)
	intMustStatus(t, attResp, http.StatusForbidden)
}

// TestDisputePDFExportEndpoint verifies that a customer can export a dispute PDF.
func TestDisputePDFExportEndpoint(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	// Create a complaint.
	complaintResp := intAPI(t, http.MethodPost, "/api/v1/complaints",
		map[string]string{"bookingId": bID, "outcome": "wheel scratch"}, custToken)
	complaintBody := intMustStatus(t, complaintResp, http.StatusCreated)
	var complaint struct{ ID string `json:"id"` }
	_ = json.Unmarshal(complaintBody, &complaint)

	// Export as PDF.
	pdfResp := intAPI(t, http.MethodGet,
		"/api/v1/exports/dispute-pdf/"+complaint.ID, nil, custToken)
	pdfBody := intMustStatus(t, pdfResp, http.StatusOK)
	if pdfResp.Header.Get("Content-Type") != "application/pdf" {
		t.Fatalf("expected application/pdf, got %s", pdfResp.Header.Get("Content-Type"))
	}
	if len(pdfBody) == 0 {
		t.Fatal("expected non-empty PDF body")
	}
}
