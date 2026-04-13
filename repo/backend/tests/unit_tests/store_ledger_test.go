package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"
)

// ---------------------------------------------------------------------------
// Ledger append / list / purge
// ---------------------------------------------------------------------------

func TestAppendAndListLedger(t *testing.T) {
	st := newStore()
	e := models.LedgerEntry{ID: "e1", BookingID: "b1", Type: "deposit", Amount: 75, CreatedAt: time.Now().UTC()}
	st.AppendLedger("b1", e)
	entries := st.ListLedger("b1")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Amount != 75 {
		t.Fatalf("expected amount 75, got %v", entries[0].Amount)
	}
}

func TestListLedger_Empty(t *testing.T) {
	st := newStore()
	entries := st.ListLedger("nonexistent-booking")
	if entries == nil {
		t.Fatal("expected empty slice, not nil")
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestListLedger_MultipleBookings(t *testing.T) {
	st := newStore()
	st.AppendLedger("b1", models.LedgerEntry{ID: "e1", BookingID: "b1", Amount: 10, CreatedAt: time.Now().UTC()})
	st.AppendLedger("b2", models.LedgerEntry{ID: "e2", BookingID: "b2", Amount: 20, CreatedAt: time.Now().UTC()})
	if len(st.ListLedger("b1")) != 1 {
		t.Fatal("expected 1 entry for b1")
	}
	if len(st.ListLedger("b2")) != 1 {
		t.Fatal("expected 1 entry for b2")
	}
}

func TestPurgeLedgerOlderThan(t *testing.T) {
	st := newStore()
	old := time.Now().UTC().Add(-48 * time.Hour)
	recent := time.Now().UTC()
	st.AppendLedger("b1", models.LedgerEntry{ID: "e1", BookingID: "b1", Type: "deposit", Amount: 50, CreatedAt: old})
	st.AppendLedger("b1", models.LedgerEntry{ID: "e2", BookingID: "b1", Type: "charge", Amount: 30, CreatedAt: recent})
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	removed := st.PurgeLedgerOlderThan(cutoff)
	if removed != 1 {
		t.Fatalf("expected 1 purged entry, got %d", removed)
	}
	remaining := st.ListLedger("b1")
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining entry, got %d", len(remaining))
	}
}

func TestPurgeLedger_NothingToRemove(t *testing.T) {
	st := newStore()
	recent := time.Now().UTC()
	st.AppendLedger("b1", models.LedgerEntry{ID: "e1", BookingID: "b1", Amount: 10, CreatedAt: recent})
	removed := st.PurgeLedgerOlderThan(recent.Add(-1 * time.Hour))
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}
}

// ---------------------------------------------------------------------------
// Hash chain integrity
// ---------------------------------------------------------------------------

func TestChainHashDeterministic(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	h1 := services.ChainHash("genesis", "payload-a", ts)
	h2 := services.ChainHash("genesis", "payload-a", ts)
	if h1 != h2 {
		t.Fatal("expected same hash for identical inputs")
	}
}

func TestChainHashChangesWithPrevHash(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	h1 := services.ChainHash("hash-a", "payload", ts)
	h2 := services.ChainHash("hash-b", "payload", ts)
	if h1 == h2 {
		t.Fatal("expected different hashes for different prevHash values")
	}
}

func TestChainHashChangesWithPayload(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	h1 := services.ChainHash("genesis", "payload-x", ts)
	h2 := services.ChainHash("genesis", "payload-y", ts)
	if h1 == h2 {
		t.Fatal("expected different hashes for different payloads")
	}
}

func TestChainHashChangesWithTimestamp(t *testing.T) {
	ts1 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 1, 15, 10, 0, 0, 1, time.UTC) // 1 nanosecond later
	h1 := services.ChainHash("genesis", "payload", ts1)
	h2 := services.ChainHash("genesis", "payload", ts2)
	if h1 == h2 {
		t.Fatal("expected different hashes for different timestamps")
	}
}

func TestChainHashGenesisEmpty(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	h := services.ChainHash("", "first-entry", ts)
	if h == "" {
		t.Fatal("expected non-empty hash for genesis entry")
	}
	if len(h) != 64 {
		t.Fatalf("expected 64-char SHA256 hex, got len %d", len(h))
	}
}

func TestLedgerChainIntegrity(t *testing.T) {
	// Build a 3-entry chain and verify each hash links to the previous
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	h0 := ""
	h1 := services.ChainHash(h0, "deposit:75", ts)
	h2 := services.ChainHash(h1, "charge:30", ts.Add(time.Minute))
	h3 := services.ChainHash(h2, "refund:10", ts.Add(2*time.Minute))

	// Verify chain by recomputing
	if services.ChainHash(h0, "deposit:75", ts) != h1 {
		t.Fatal("h1 recompute failed")
	}
	if services.ChainHash(h1, "charge:30", ts.Add(time.Minute)) != h2 {
		t.Fatal("h2 recompute failed")
	}
	if services.ChainHash(h2, "refund:10", ts.Add(2*time.Minute)) != h3 {
		t.Fatal("h3 recompute failed")
	}
}

func TestTamperedLedgerDetected(t *testing.T) {
	// Simulates tamper detection by showing that changing prevHash breaks recompute
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	h1 := services.ChainHash("", "deposit:75", ts)
	h2 := services.ChainHash(h1, "charge:30", ts.Add(time.Minute))

	// Tamper: change h1 and see if h2 can be reproduced
	fakeh1 := "tampered-hash"
	fakeH2 := services.ChainHash(fakeh1, "charge:30", ts.Add(time.Minute))
	if fakeH2 == h2 {
		t.Fatal("tampered chain should produce different hash")
	}
}
