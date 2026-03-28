package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

type DisputePDFData struct {
	ComplaintID    string
	BookingID      string
	Status         string
	Outcome        string
	OpenedBy       string
	GeneratedAt    time.Time
	InspectionRows []string
	LedgerRows     []string
}

func GenerateDisputePDF(data DisputePDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "FleetLease Dispute Evidence Export", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Generated: %s", data.GeneratedAt.UTC().Format(time.RFC3339)), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Complaint ID: %s", data.ComplaintID), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Booking ID: %s", data.BookingID), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Opened By: %s", data.OpenedBy), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Status: %s", data.Status), "", 1, "L", false, 0, "")
	pdf.MultiCell(0, 7, fmt.Sprintf("Outcome: %s", data.Outcome), "", "L", false)

	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Inspection Hash Chain", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	if len(data.InspectionRows) == 0 {
		pdf.CellFormat(0, 6, "No inspection revisions found.", "", 1, "L", false, 0, "")
	} else {
		for _, row := range data.InspectionRows {
			pdf.MultiCell(0, 6, row, "", "L", false)
		}
	}

	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Settlement Ledger Hash Chain", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	if len(data.LedgerRows) == 0 {
		pdf.CellFormat(0, 6, "No ledger entries found.", "", 1, "L", false, 0, "")
	} else {
		for _, row := range data.LedgerRows {
			pdf.MultiCell(0, 6, row, "", "L", false)
		}
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
