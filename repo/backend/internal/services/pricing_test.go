package services

import (
	"testing"
	"time"
)

func TestEstimateFareNightWindow(t *testing.T) {
	cfg := DefaultPricingConfig()
	start := time.Date(2026, 3, 27, 22, 15, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)
	res := EstimateFare(cfg, EstimateInput{StartAt: start, EndAt: end, OdoStart: 100, OdoEnd: 112, Deposit: 75})
	if res.NightSurcharge <= 0 {
		t.Fatalf("expected night surcharge > 0, got %v", res.NightSurcharge)
	}
	if res.Total <= res.BaseAmount {
		t.Fatalf("expected total > base")
	}
}

func TestEstimateFareIncludedMilesAndMinFare(t *testing.T) {
	cfg := DefaultPricingConfig()
	cfg.IncludedMiles = 10
	cfg.MinFare = 5
	start := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Minute)
	res := EstimateFare(cfg, EstimateInput{StartAt: start, EndAt: end, OdoStart: 10, OdoEnd: 12, Deposit: 50})
	if res.MileageAmount != 0 {
		t.Fatalf("expected zero mileage charge with included miles, got %v", res.MileageAmount)
	}
	if res.Total < cfg.MinFare {
		t.Fatalf("expected min fare to apply")
	}
}

func TestEstimateFareNegativeAndNoNight(t *testing.T) {
	cfg := DefaultPricingConfig()
	start := time.Date(2026, 3, 27, 8, 0, 0, 0, time.UTC)
	end := start.Add(-1 * time.Minute)
	res := EstimateFare(cfg, EstimateInput{StartAt: start, EndAt: end, OdoStart: 20, OdoEnd: 10, Deposit: 0})
	if res.MileageAmount != 0 {
		t.Fatalf("expected zero mileage for negative odo diff")
	}
	if res.NightSurcharge != 0 {
		t.Fatalf("expected no night surcharge")
	}
	if res.Deposit != cfg.DepositDefault {
		t.Fatalf("expected deposit default fallback")
	}
}
