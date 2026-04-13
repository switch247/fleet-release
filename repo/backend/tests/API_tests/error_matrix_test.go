package api_tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPIErrorMatrix(t *testing.T) {
	custToken := liveLogin(t, apiCustUser, apiCustPass)
	agentToken := liveLogin(t, apiAgentUser, apiAgentPass)

	// Create a consultation on the seeded booking via agent.
	resp := apiCall(t, http.MethodPost, "/api/v1/consultations", map[string]interface{}{
		"bookingId":  apiBookingID,
		"topic":      "lane",
		"visibility": "csa_admin",
	}, agentToken)
	b := mustAPIStatus(t, resp, http.StatusCreated)
	var consultResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &consultResult); err != nil || consultResult.ID == "" {
		t.Fatalf("create consultation: bad response %s", b)
	}

	// Settle a fresh booking so we can provoke a 409 on a second attempt.
	// We need a booking we can settle; use createFreshAPIBooking so the seeded
	// apiBookingID is left intact for other tests.
	freshID := createFreshAPIBooking(t, custToken)

	resp2 := apiCall(t, http.MethodPost, "/api/v1/settlements/close/"+freshID, nil, custToken)
	mustAPIStatus(t, resp2, http.StatusOK)

	// -----------------------------------------------------------------------
	// Error matrix
	// -----------------------------------------------------------------------
	testCases := []struct {
		name   string
		method string
		path   string
		body   interface{}
		token  string
		expect int
	}{
		{
			name:   "bookings_unauthenticated",
			method: http.MethodGet,
			path:   "/api/v1/bookings",
			token:  "",
			expect: http.StatusUnauthorized,
		},
		{
			// Consultation was created with visibility=csa_admin; customer should
			// be forbidden from reading its attachments.
			name:   "consultation_evidence_forbidden",
			method: http.MethodGet,
			path:   "/api/v1/consultations/" + consultResult.ID + "/attachments",
			token:  custToken,
			expect: http.StatusForbidden,
		},
		{
			name:   "inspections_missing_booking",
			method: http.MethodGet,
			path:   "/api/v1/inspections?bookingId=missing-id",
			token:  custToken,
			expect: http.StatusNotFound,
		},
		{
			// Second settlement attempt on the already-settled fresh booking → 409.
			name:   "settlement_conflict",
			method: http.MethodPost,
			path:   "/api/v1/settlements/close/" + freshID,
			token:  custToken,
			expect: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp := apiCall(t, tc.method, tc.path, tc.body, tc.token)
			mustAPIStatus(t, resp, tc.expect)
		})
	}
}
