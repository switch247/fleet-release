package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"fleetlease/backend/pkg/public"
)

func TestCouponRedemptionConcurrentDedup(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")
	payload, _ := json.Marshal(map[string]string{
		"code":      "OFFLINE-COUPON-1",
		"bookingId": h.BookingID,
	})

	var wg sync.WaitGroup
	var success int32
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/coupons/redeem", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			h.Router.ServeHTTP(rec, req)
			if rec.Code == http.StatusOK {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Fatalf("expected exactly one successful redemption, got %d", success)
	}
}
