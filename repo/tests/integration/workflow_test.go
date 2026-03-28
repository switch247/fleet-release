package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestConsultationVisibilityForCustomer(t *testing.T) {
	h := public.BuildHarnessForTests()
	agentToken := loginToken(t, h.Router, "agent", "Agent1234!Pass")
	customerToken := loginToken(t, h.Router, "customer", "Customer1234!")

	createPayload, _ := json.Marshal(map[string]string{
		"bookingId":      h.BookingID,
		"topic":          "damage review",
		"keyPoints":      "mirror crack",
		"recommendation": "collect proof",
		"followUp":       "call both parties",
		"visibility":     "csa_admin",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/consultations", bytes.NewReader(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+agentToken)
	createRec := httptest.NewRecorder()
	h.Router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 create consultation got %d body=%s", createRec.Code, createRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/consultations?bookingId="+h.BookingID, nil)
	listReq.Header.Set("Authorization", "Bearer "+customerToken)
	listRec := httptest.NewRecorder()
	h.Router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 list consultations got %d body=%s", listRec.Code, listRec.Body.String())
	}
	var consultations []map[string]interface{}
	_ = json.Unmarshal(listRec.Body.Bytes(), &consultations)
	if len(consultations) != 0 {
		t.Fatalf("expected csa_admin consultation to be hidden from customer; got %d rows", len(consultations))
	}
}

func TestConsultationAttachmentVisibilityForCustomer(t *testing.T) {
	h := public.BuildHarnessForTests()
	agentToken := loginToken(t, h.Router, "agent", "Agent1234!Pass")
	customerToken := loginToken(t, h.Router, "customer", "Customer1234!")

	createPayload, _ := json.Marshal(map[string]string{
		"bookingId":      h.BookingID,
		"topic":          "escalation note",
		"keyPoints":      "internal evidence",
		"recommendation": "await supervisor",
		"followUp":       "audit review",
		"visibility":     "csa_admin",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/consultations", bytes.NewReader(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+agentToken)
	createRec := httptest.NewRecorder()
	h.Router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 consultation got %d body=%s", createRec.Code, createRec.Body.String())
	}
	var consultation struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createRec.Body.Bytes(), &consultation)

	attachmentInitBody, _ := json.Marshal(map[string]interface{}{
		"bookingId":   h.BookingID,
		"type":        "photo",
		"sizeBytes":   10,
		"checksum":    "abc123",
		"fingerprint": "consultation-evidence-fp",
	})
	attachmentInitReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/chunk/init", bytes.NewReader(attachmentInitBody))
	attachmentInitReq.Header.Set("Content-Type", "application/json")
	attachmentInitReq.Header.Set("Authorization", "Bearer "+agentToken)
	attachmentInitRec := httptest.NewRecorder()
	h.Router.ServeHTTP(attachmentInitRec, attachmentInitReq)
	if attachmentInitRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 attachment init got %d body=%s", attachmentInitRec.Code, attachmentInitRec.Body.String())
	}
	var attachmentResp struct {
		UploadID string `json:"uploadId"`
	}
	_ = json.Unmarshal(attachmentInitRec.Body.Bytes(), &attachmentResp)

	attachBody, _ := json.Marshal(map[string]string{
		"consultationId": consultation.ID,
		"attachmentId":   attachmentResp.UploadID,
	})
	attachReq := httptest.NewRequest(http.MethodPost, "/api/v1/consultations/attachments", bytes.NewReader(attachBody))
	attachReq.Header.Set("Content-Type", "application/json")
	attachReq.Header.Set("Authorization", "Bearer "+agentToken)
	attachRec := httptest.NewRecorder()
	h.Router.ServeHTTP(attachRec, attachReq)
	if attachRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 consultation attachment got %d body=%s", attachRec.Code, attachRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/consultations/"+consultation.ID+"/attachments", nil)
	listReq.Header.Set("Authorization", "Bearer "+customerToken)
	listRec := httptest.NewRecorder()
	h.Router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for customer consultation evidence got %d body=%s", listRec.Code, listRec.Body.String())
	}
}

func TestDisputePDFExportEndpoint(t *testing.T) {
	h := public.BuildHarnessForTests()
	customerToken := loginToken(t, h.Router, "customer", "Customer1234!")

	createComplaintBody, _ := json.Marshal(map[string]string{
		"bookingId": h.BookingID,
		"outcome":   "wheel scratch",
	})
	createComplaintReq := httptest.NewRequest(http.MethodPost, "/api/v1/complaints", bytes.NewReader(createComplaintBody))
	createComplaintReq.Header.Set("Content-Type", "application/json")
	createComplaintReq.Header.Set("Authorization", "Bearer "+customerToken)
	createComplaintRec := httptest.NewRecorder()
	h.Router.ServeHTTP(createComplaintRec, createComplaintReq)
	if createComplaintRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 complaint got %d body=%s", createComplaintRec.Code, createComplaintRec.Body.String())
	}
	var complaint struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createComplaintRec.Body.Bytes(), &complaint)

	exportReq := httptest.NewRequest(http.MethodGet, "/api/v1/exports/dispute-pdf/"+complaint.ID, nil)
	exportReq.Header.Set("Authorization", "Bearer "+customerToken)
	exportRec := httptest.NewRecorder()
	h.Router.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected 200 dispute pdf export got %d body=%s", exportRec.Code, exportRec.Body.String())
	}
	if exportRec.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("expected application/pdf content type, got %s", exportRec.Header().Get("Content-Type"))
	}
	if exportRec.Body.Len() == 0 {
		t.Fatalf("expected non-empty pdf body")
	}
}
