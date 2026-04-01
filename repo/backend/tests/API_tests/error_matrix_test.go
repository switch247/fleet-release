package api_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestAPIErrorMatrix(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	customerToken := loginForEndpoint(t, e, "customer", "Customer1234!")
	agentToken := loginForEndpoint(t, e, "agent", "Agent1234!Pass")

	consultationBody := `{"bookingId":"22222222-2222-2222-2222-222222222222","topic":"lane","visibility":"csa_admin"}`
	createConsult := httptest.NewRequest(http.MethodPost, "/api/v1/consultations", bytes.NewBufferString(consultationBody))
	createConsult.Header.Set("Content-Type", "application/json")
	createConsult.Header.Set("Authorization", "Bearer "+agentToken)
	createConsultRec := httptest.NewRecorder()
	e.ServeHTTP(createConsultRec, createConsult)
	if createConsultRec.Code != http.StatusCreated {
		t.Fatalf("consultation setup failed: %d %s", createConsultRec.Code, createConsultRec.Body.String())
	}
	var consultResult struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createConsultRec.Body.Bytes(), &consultResult)

	// settle booking once to trigger conflict on second attempt
	settleReq := httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/22222222-2222-2222-2222-222222222222", nil)
	settleReq.Header.Set("Authorization", "Bearer "+customerToken)
	settleRec := httptest.NewRecorder()
	e.ServeHTTP(settleRec, settleReq)
	if settleRec.Code != http.StatusOK {
		t.Fatalf("initial settlement failed: %d %s", settleRec.Code, settleRec.Body.String())
	}

	testCases := []struct {
		name   string
		req    *http.Request
		token  string
		expect int
	}{
		{
			name:   "bookings_unauthenticated",
			req:    httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil),
			token:  "",
			expect: http.StatusUnauthorized,
		},
		{
			name:   "consultation_evidence_forbidden",
			req:    httptest.NewRequest(http.MethodGet, "/api/v1/consultations/"+consultResult.ID+"/attachments", nil),
			token:  customerToken,
			expect: http.StatusForbidden,
		},
		{
			name:   "inspections_missing_booking",
			req:    httptest.NewRequest(http.MethodGet, "/api/v1/inspections?bookingId=missing-id", nil),
			token:  customerToken,
			expect: http.StatusNotFound,
		},
		{
			name:   "settlement_conflict",
			req:    httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/22222222-2222-2222-2222-222222222222", nil),
			token:  customerToken,
			expect: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.token != "" {
				tc.req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, tc.req)
			if rec.Code != tc.expect {
				t.Fatalf("%s expected %d got %d body=%s", tc.name, tc.expect, rec.Code, rec.Body.String())
			}
		})
	}
}
