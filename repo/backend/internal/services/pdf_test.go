package services

import (
	"bytes"
	"testing"
	"time"
)

func TestGenerateDisputePDF(t *testing.T) {
	pdfBytes, err := GenerateDisputePDF(DisputePDFData{
		ComplaintID:    "c1",
		BookingID:      "b1",
		Status:         "open",
		Outcome:        "pending review",
		OpenedBy:       "u1",
		GeneratedAt:    time.Now().UTC(),
		InspectionRows: []string{"rev-1 hash=abc"},
		LedgerRows:     []string{"entry-1 hash=def"},
	})
	if err != nil {
		t.Fatalf("pdf generation failed: %v", err)
	}
	if len(pdfBytes) == 0 || !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatalf("expected valid pdf output")
	}
}
