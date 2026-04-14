package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/pkg/public"
)

func loginToken(t *testing.T, e http.Handler, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid login response: %v", err)
	}
	return payload.Token
}

func TestSettlementHashChainDetectsTampering(t *testing.T) {
	h := public.BuildHarnessForTests()
	token := loginToken(t, h.Router, "customer", "Customer1234!")

	closeReq := httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/"+h.BookingID, nil)
	closeReq.Header.Set("Authorization", "Bearer "+token)
	closeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(closeRec, closeReq)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("close settlement expected 200 got %d body=%s", closeRec.Code, closeRec.Body.String())
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/"+h.BookingID+"/verify", nil)
	verifyReq.Header.Set("Authorization", "Bearer "+token)
	verifyRec := httptest.NewRecorder()
	h.Router.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify expected 200 got %d body=%s", verifyRec.Code, verifyRec.Body.String())
	}
	var before struct {
		Valid bool `json:"valid"`
	}
	_ = json.Unmarshal(verifyRec.Body.Bytes(), &before)
	if !before.Valid {
		t.Fatalf("expected chain to be valid before tamper")
	}

	h.TamperLedger(h.BookingID)

	afterReq := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/"+h.BookingID+"/verify", nil)
	afterReq.Header.Set("Authorization", "Bearer "+token)
	afterRec := httptest.NewRecorder()
	h.Router.ServeHTTP(afterRec, afterReq)
	if afterRec.Code != http.StatusOK {
		t.Fatalf("verify after tamper expected 200 got %d body=%s", afterRec.Code, afterRec.Body.String())
	}
	var after struct {
		Valid bool `json:"valid"`
	}
	_ = json.Unmarshal(afterRec.Body.Bytes(), &after)
	if after.Valid {
		t.Fatalf("expected chain to be invalid after tampering")
	}
}

// TestSettlementIncludesWearDeductions verifies that when an inspection records
// damage deduction amounts, closing settlement produces a wear_deduction ledger
// entry and the deposit refund/deduction is adjusted accordingly.
func TestSettlementIncludesWearDeductions(t *testing.T) {
	h := public.BuildHarnessForTests()

	// Seed an inspection with damage items totalling $100 in deductions.
	h.SeedInspection([]models.InspectionItem{
		{Name: "Exterior", Condition: "minor", DamageDeductionAmount: 20.0},
		{Name: "Interior", Condition: "major", DamageDeductionAmount: 80.0},
	})

	token := loginToken(t, h.Router, "customer", "Customer1234!")

	closeReq := httptest.NewRequest(http.MethodPost, "/api/v1/settlements/close/"+h.BookingID, nil)
	closeReq.Header.Set("Authorization", "Bearer "+token)
	closeRec := httptest.NewRecorder()
	h.Router.ServeHTTP(closeRec, closeReq)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("close settlement expected 200 got %d body=%s", closeRec.Code, closeRec.Body.String())
	}

	var body struct {
		Ledger []struct {
			Type   string  `json:"type"`
			Amount float64 `json:"amount"`
		} `json:"ledger"`
	}
	if err := json.Unmarshal(closeRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse settlement response: %v", err)
	}

	foundDeduction := false
	var deductionAmount float64
	for _, entry := range body.Ledger {
		if entry.Type == "wear_deduction" {
			foundDeduction = true
			deductionAmount = entry.Amount
		}
	}
	if !foundDeduction {
		t.Fatal("expected wear_deduction entry in settlement ledger")
	}
	if deductionAmount != 100.0 {
		t.Fatalf("expected wear_deduction amount=100.00 got %.2f", deductionAmount)
	}

	// The booking has EstimatedAmount=25, DepositAmount=75, Deductions=100.
	// net deposit refund = 75 - 25 - 100 = -50  →  deposit_deduction
	for _, entry := range body.Ledger {
		if entry.Type == "deposit_deduction" {
			if entry.Amount != -50.0 {
				t.Fatalf("expected deposit_deduction=-50.00 got %.2f", entry.Amount)
			}
			return
		}
	}
	t.Fatal("expected deposit_deduction entry after wear deductions exceeded deposit")
}
