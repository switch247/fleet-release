package unit_tests

import (
	"strings"
	"testing"
	"time"

	"fleetlease/backend/internal/services"
)

func pt(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// ---------------------------------------------------------------------------
// AES-256 encryption / decryption
// ---------------------------------------------------------------------------

var testKey32 = []byte("0123456789abcdef0123456789abcdef")

func TestEncryptAES256_ProducesOutput(t *testing.T) {
	ciphertext, err := services.EncryptAES256(testKey32, "sensitive-data")
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("expected non-empty ciphertext")
	}
	if ciphertext == "sensitive-data" {
		t.Fatal("ciphertext must differ from plaintext")
	}
}

func TestEncryptAES256_WrongKeyLength(t *testing.T) {
	shortKey := []byte("too-short")
	_, err := services.EncryptAES256(shortKey, "data")
	if err == nil {
		t.Fatal("expected error for wrong key length")
	}
	if !strings.Contains(err.Error(), "32 bytes") {
		t.Fatalf("expected '32 bytes' in error, got: %v", err)
	}
}

func TestEncryptAES256_EmptyKey(t *testing.T) {
	_, err := services.EncryptAES256([]byte{}, "data")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestEncryptAES256_EmptyPlaintext(t *testing.T) {
	ciphertext, err := services.EncryptAES256(testKey32, "")
	if err != nil {
		t.Fatalf("unexpected error for empty plaintext: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("expected non-empty ciphertext even for empty plaintext")
	}
}

func TestEncryptAES256_Nondeterministic(t *testing.T) {
	// Each call should produce different ciphertext due to random nonce
	c1, _ := services.EncryptAES256(testKey32, "same-data")
	c2, _ := services.EncryptAES256(testKey32, "same-data")
	if c1 == c2 {
		t.Fatal("AES-GCM encryption must be nondeterministic (random nonce)")
	}
}

// ---------------------------------------------------------------------------
// MaskSensitive
// ---------------------------------------------------------------------------

func TestMaskSensitive_ShortString(t *testing.T) {
	result := services.MaskSensitive("abc")
	if result != "****" {
		t.Fatalf("expected **** for short string, got %s", result)
	}
}

func TestMaskSensitive_EmptyString(t *testing.T) {
	result := services.MaskSensitive("")
	if result != "****" {
		t.Fatalf("expected **** for empty string, got %s", result)
	}
}

func TestMaskSensitive_ExactlyFourChars(t *testing.T) {
	result := services.MaskSensitive("abcd")
	if result != "****" {
		t.Fatalf("expected **** for 4-char string, got %s", result)
	}
}

func TestMaskSensitive_LongerString(t *testing.T) {
	result := services.MaskSensitive("ABCDEFGH")
	// First 2 + **** + last 2 = "AB****GH"
	if !strings.HasPrefix(result, "AB") {
		t.Fatalf("expected prefix AB, got %s", result)
	}
	if !strings.HasSuffix(result, "GH") {
		t.Fatalf("expected suffix GH, got %s", result)
	}
	if !strings.Contains(result, "****") {
		t.Fatalf("expected **** in middle, got %s", result)
	}
}

func TestMaskSensitive_DoesNotRevealFullValue(t *testing.T) {
	original := "supersecretvalue1234"
	masked := services.MaskSensitive(original)
	if masked == original {
		t.Fatal("masked value must not equal original")
	}
	if strings.Contains(masked, "supersecret") {
		t.Fatal("masked value must not contain full original content")
	}
}

// ---------------------------------------------------------------------------
// Pricing edge cases
// ---------------------------------------------------------------------------

func TestEstimateFare_ZeroMiles(t *testing.T) {
	// No mileage charge when odo didn't change (or moved < included miles)
	result := services.EstimateFare(services.DefaultPricingConfig(), services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 100,
		OdoEnd:   100, // no movement
		Deposit:  75,
	})
	if result.MileageAmount != 0 {
		t.Fatalf("expected 0 mileage charge for no movement, got %v", result.MileageAmount)
	}
}

func TestEstimateFare_WithinIncludedMiles(t *testing.T) {
	cfg := services.DefaultPricingConfig()
	// Default included miles = 2; drive 1 mile -> still within free miles
	result := services.EstimateFare(cfg, services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 100,
		OdoEnd:   101,
		Deposit:  75,
	})
	if result.MileageAmount != 0 {
		t.Fatalf("expected 0 mileage for 1 mile within 2-mile allowance, got %v", result.MileageAmount)
	}
}

func TestEstimateFare_ExceedsIncludedMiles(t *testing.T) {
	cfg := services.DefaultPricingConfig()
	// Drive 10 miles, 2 included → 8 billable
	result := services.EstimateFare(cfg, services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 100,
		OdoEnd:   110,
		Deposit:  75,
	})
	expected := 8 * cfg.PerMile
	if result.MileageAmount < expected-0.01 || result.MileageAmount > expected+0.01 {
		t.Fatalf("expected mileage ~%v, got %v", expected, result.MileageAmount)
	}
}

func TestEstimateFare_NegativeOdoReturnsZero(t *testing.T) {
	// OdoEnd < OdoStart should not produce negative mileage charge
	result := services.EstimateFare(services.DefaultPricingConfig(), services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 200,
		OdoEnd:   100, // reversed
		Deposit:  75,
	})
	if result.MileageAmount < 0 {
		t.Fatalf("expected non-negative mileage for reversed odo, got %v", result.MileageAmount)
	}
}

func TestEstimateFare_MinFareApplied(t *testing.T) {
	// Very short trip with no extra miles should hit minimum fare
	cfg := services.DefaultPricingConfig()
	result := services.EstimateFare(cfg, services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T10:01:00Z"), // 1 minute
		OdoStart: 100,
		OdoEnd:   100, // no movement
		Deposit:  75,
	})
	if result.Total < cfg.MinFare {
		t.Fatalf("expected total >= min fare %v, got %v", cfg.MinFare, result.Total)
	}
}

func TestEstimateFare_DaytimeNoNightSurcharge(t *testing.T) {
	result := services.EstimateFare(services.DefaultPricingConfig(), services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 100,
		OdoEnd:   110,
		Deposit:  75,
	})
	if result.NightSurcharge != 0 {
		t.Fatalf("expected no night surcharge for daytime trip, got %v", result.NightSurcharge)
	}
}

func TestEstimateFare_DefaultDepositUsed(t *testing.T) {
	cfg := services.DefaultPricingConfig()
	// Pass zero deposit — should use default
	result := services.EstimateFare(cfg, services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 100,
		OdoEnd:   105,
		Deposit:  0,
	})
	if result.Deposit != cfg.DepositDefault {
		t.Fatalf("expected deposit %v, got %v", cfg.DepositDefault, result.Deposit)
	}
}

func TestEstimateFare_CustomDepositHonoured(t *testing.T) {
	result := services.EstimateFare(services.DefaultPricingConfig(), services.EstimateInput{
		StartAt:  pt("2026-01-15T10:00:00Z"),
		EndAt:    pt("2026-01-15T11:00:00Z"),
		OdoStart: 100,
		OdoEnd:   105,
		Deposit:  150,
	})
	if result.Deposit != 150 {
		t.Fatalf("expected custom deposit 150, got %v", result.Deposit)
	}
}
