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
