package services

import "time"

type PricingConfig struct {
	BaseFare       float64
	PerMile        float64
	PerMinute      float64
	NightSurcharge float64
	IncludedMiles  float64
	MinFare        float64
	DepositDefault float64
}

type EstimateInput struct {
	StartAt  time.Time
	EndAt    time.Time
	OdoStart float64
	OdoEnd   float64
	Deposit  float64
}

type EstimateResult struct {
	BaseAmount     float64 `json:"baseAmount"`
	MileageAmount  float64 `json:"mileageAmount"`
	TimeAmount     float64 `json:"timeAmount"`
	NightSurcharge float64 `json:"nightSurcharge"`
	Total          float64 `json:"total"`
	Deposit        float64 `json:"deposit"`
}

func DefaultPricingConfig() PricingConfig {
	return PricingConfig{
		BaseFare:       1.80,
		PerMile:        0.65,
		PerMinute:      0.22,
		NightSurcharge: 0.20,
		IncludedMiles:  2,
		MinFare:        5,
		DepositDefault: 75,
	}
}

func EstimateFare(cfg PricingConfig, in EstimateInput) EstimateResult {
	miles := in.OdoEnd - in.OdoStart
	if miles < 0 {
		miles = 0
	}
	billableMiles := miles - cfg.IncludedMiles
	if billableMiles < 0 {
		billableMiles = 0
	}
	minutes := in.EndAt.Sub(in.StartAt).Minutes()
	if minutes < 0 {
		minutes = 0
	}

	base := cfg.BaseFare
	mileage := billableMiles * cfg.PerMile
	timeAmount := minutes * cfg.PerMinute
	subtotal := base + mileage + timeAmount
	if subtotal < cfg.MinFare {
		subtotal = cfg.MinFare
	}

	night := 0.0
	if intersectsNightWindow(in.StartAt, in.EndAt) {
		night = subtotal * cfg.NightSurcharge
	}
	total := subtotal + night
	deposit := in.Deposit
	if deposit <= 0 {
		deposit = cfg.DepositDefault
	}

	return EstimateResult{
		BaseAmount:     round2(base),
		MileageAmount:  round2(mileage),
		TimeAmount:     round2(timeAmount),
		NightSurcharge: round2(night),
		Total:          round2(total),
		Deposit:        round2(deposit),
	}
}

func intersectsNightWindow(start, end time.Time) bool {
	if !end.After(start) {
		return false
	}
	for t := start; t.Before(end); t = t.Add(time.Minute * 30) {
		h := t.Hour()
		if h >= 22 || h <= 5 {
			return true
		}
	}
	return false
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
