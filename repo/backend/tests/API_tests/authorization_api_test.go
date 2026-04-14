package api_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestProviderCannotSettleUnownedBooking verifies that a provider who does not
// own a booking receives 403 when attempting to settle it.
func TestProviderCannotSettleUnownedBooking(t *testing.T) {
	adminToken := liveLoginAdmin(t)

	// Create a throwaway second provider so we have someone who definitely does
	// not own apiBookingID (which belongs to apiProvID / api-provider).
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	prov2User := "prov2-" + suffix
	prov2Pass := "ProvTwo1234!"
	createTempUser(t, adminToken, prov2User, prov2Pass, []string{"provider"})

	prov2Token := liveLogin(t, prov2User, prov2Pass)

	resp := apiCall(t, http.MethodPost, "/api/v1/settlements/close/"+apiBookingID, nil, prov2Token)
	mustAPIStatus(t, resp, http.StatusForbidden)
}

// TestComplaintArbitrationRequiresCSAOrAdmin verifies that a customer who
// creates a complaint cannot arbitrate it — only CSA/admin may do so.
func TestComplaintArbitrationRequiresCSAOrAdmin(t *testing.T) {
	custToken := liveLogin(t, apiCustUser, apiCustPass)

	// Create a complaint on the seeded booking.
	resp := apiCall(t, http.MethodPost, "/api/v1/complaints", map[string]string{
		"bookingId": apiBookingID,
		"outcome":   "broken mirror",
	}, custToken)
	b := mustAPIStatus(t, resp, http.StatusCreated)

	var complaint struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &complaint); err != nil || complaint.ID == "" {
		t.Fatalf("create complaint: bad response %s", b)
	}

	// Customer attempts to arbitrate their own complaint → must be 403.
	resp2 := apiCall(t, http.MethodPatch, "/api/v1/complaints/"+complaint.ID+"/arbitrate",
		map[string]string{"status": "closed", "outcome": "denied"},
		custToken)
	mustAPIStatus(t, resp2, http.StatusForbidden)
}

// TestNonCustomerCannotCreateBooking verifies that a provider cannot create a
// booking (only customers may).
func TestNonCustomerCannotCreateBooking(t *testing.T) {
	provToken := liveLogin(t, apiProvUser, apiProvPass)

	now := time.Now().UTC()
	resp := apiCall(t, http.MethodPost, "/api/v1/bookings", map[string]interface{}{
		"listingId": apiListingID,
		"startAt":   now.Format(time.RFC3339),
		"endAt":     now.Add(2 * time.Hour).Format(time.RFC3339),
		"odoStart":  1000.0,
		"odoEnd":    1100.0,
	}, provToken)
	mustAPIStatus(t, resp, http.StatusForbidden)
}
