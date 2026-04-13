package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCouponRedemptionConcurrentDedup fires 8 concurrent redemption requests
// for the same coupon code and asserts that exactly one succeeds.
func TestCouponRedemptionConcurrentDedup(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	couponCode := fmt.Sprintf("CONCUR-%d", time.Now().UnixNano())

	payload, _ := json.Marshal(map[string]string{
		"code":      couponCode,
		"bookingId": bID,
	})

	var wg sync.WaitGroup
	var success int32
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodPost,
				intLiveServerURL+"/api/v1/coupons/redeem",
				bytes.NewReader(payload))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+custToken)
			resp, err := intLiveClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Fatalf("expected exactly 1 successful coupon redemption, got %d", success)
	}
}
