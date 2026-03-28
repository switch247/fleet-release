package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/pkg/public"
)

func TestEstimateFareNightWindow(t *testing.T) {
	start := time.Date(2026, 3, 27, 22, 15, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)
	res := public.EstimateFare(public.EstimateInput{StartAt: start, EndAt: end, OdoStart: 100, OdoEnd: 112, Deposit: 75})
	if res.NightSurcharge <= 0 {
		t.Fatalf("expected night surcharge > 0, got %v", res.NightSurcharge)
	}
	if res.Total <= res.BaseAmount {
		t.Fatalf("expected total > base")
	}
}
