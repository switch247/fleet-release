package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
)

// ---------------------------------------------------------------------------
// Attachments: CRUD, fingerprint dedup, purge
// ---------------------------------------------------------------------------

func TestSaveAndGetAttachment(t *testing.T) {
	st := newStore()
	a := models.Attachment{
		ID:          "att1",
		BookingID:   "b1",
		Type:        "photo",
		Path:        "/data/att1.jpg",
		SizeBytes:   1024,
		Checksum:    "abc123",
		Fingerprint: "fp-abc",
		CreatedAt:   time.Now().UTC(),
	}
	st.SaveAttachment(a)
	got, ok := st.GetAttachment("att1")
	if !ok {
		t.Fatal("expected attachment to be found")
	}
	if got.Checksum != "abc123" {
		t.Fatalf("expected checksum abc123, got %s", got.Checksum)
	}
}

func TestGetAttachment_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetAttachment("no-att")
	if ok {
		t.Fatal("expected false for missing attachment")
	}
}

func TestFindAttachmentByFingerprint(t *testing.T) {
	st := newStore()
	st.SaveAttachment(models.Attachment{
		ID:          "att1",
		BookingID:   "b1",
		Fingerprint: "unique-fp-001",
		CreatedAt:   time.Now().UTC(),
	})
	got, ok := st.FindAttachmentByFingerprint("unique-fp-001")
	if !ok {
		t.Fatal("expected to find attachment by fingerprint")
	}
	if got.ID != "att1" {
		t.Fatalf("expected att1, got %s", got.ID)
	}
}

func TestFindAttachmentByFingerprint_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.FindAttachmentByFingerprint("no-such-fp")
	if ok {
		t.Fatal("expected false for unknown fingerprint")
	}
}

func TestFindAttachmentByFingerprint_DedupScenario(t *testing.T) {
	// Simulate: first attach gets saved; second upload finds existing by fingerprint (dedup)
	st := newStore()
	st.SaveAttachment(models.Attachment{
		ID:          "att1",
		BookingID:   "b1",
		Fingerprint: "sha256:aaaa",
		CreatedAt:   time.Now().UTC(),
	})
	// Second upload with same content fingerprint should find the existing one
	existing, ok := st.FindAttachmentByFingerprint("sha256:aaaa")
	if !ok {
		t.Fatal("expected dedup to find existing attachment")
	}
	if existing.ID != "att1" {
		t.Fatalf("expected att1, got %s", existing.ID)
	}
}

func TestPurgeAttachmentsOlderThan(t *testing.T) {
	st := newStore()
	old := time.Now().UTC().Add(-72 * time.Hour)
	recent := time.Now().UTC()
	st.SaveAttachment(models.Attachment{ID: "old1", BookingID: "b1", Fingerprint: "fp1", CreatedAt: old})
	st.SaveAttachment(models.Attachment{ID: "new1", BookingID: "b1", Fingerprint: "fp2", CreatedAt: recent})
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	removed := st.PurgeAttachmentsOlderThan(cutoff)
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed attachment, got %d", len(removed))
	}
	if removed[0].ID != "old1" {
		t.Fatalf("expected old1 removed, got %s", removed[0].ID)
	}
	_, ok := st.GetAttachment("new1")
	if !ok {
		t.Fatal("expected new attachment to be retained")
	}
}

func TestPurgeAttachments_SkipsZeroCreatedAt(t *testing.T) {
	st := newStore()
	// Attachment with zero time should be skipped by purge
	st.SaveAttachment(models.Attachment{ID: "att-zero", BookingID: "b1", Fingerprint: "fp-zero"})
	// zero CreatedAt gets set to now by SaveAttachment, so it won't be purged with a past cutoff
	cutoff := time.Now().UTC().Add(-1 * time.Hour)
	removed := st.PurgeAttachmentsOlderThan(cutoff)
	if len(removed) != 0 {
		t.Fatalf("expected 0 removed for recent attachment, got %d", len(removed))
	}
}

func TestPurgeAttachments_NoneToRemove(t *testing.T) {
	st := newStore()
	recent := time.Now().UTC()
	st.SaveAttachment(models.Attachment{ID: "new1", BookingID: "b1", Fingerprint: "fp1", CreatedAt: recent})
	cutoff := recent.Add(-1 * time.Hour)
	removed := st.PurgeAttachmentsOlderThan(cutoff)
	if len(removed) != 0 {
		t.Fatalf("expected 0 removed, got %d", len(removed))
	}
}

// ---------------------------------------------------------------------------
// Inspections
// ---------------------------------------------------------------------------

func TestSaveAndListInspections(t *testing.T) {
	st := newStore()
	rev := models.InspectionRevision{
		RevisionID: "rev1",
		BookingID:  "b1",
		Stage:      "initial",
		Notes:      "All good",
		Items: []models.InspectionItem{
			{Name: "exterior", Condition: "ok", EvidenceIDs: []string{"att1"}},
		},
	}
	st.SaveInspection("b1", rev)
	revisions := st.ListInspections("b1")
	if len(revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(revisions))
	}
	if revisions[0].Stage != "initial" {
		t.Fatalf("expected stage initial, got %s", revisions[0].Stage)
	}
}

func TestListInspections_MultipleRevisions(t *testing.T) {
	st := newStore()
	st.SaveInspection("b1", models.InspectionRevision{RevisionID: "r1", Stage: "initial"})
	st.SaveInspection("b1", models.InspectionRevision{RevisionID: "r2", Stage: "final"})
	revisions := st.ListInspections("b1")
	if len(revisions) != 2 {
		t.Fatalf("expected 2 revisions, got %d", len(revisions))
	}
}

func TestListInspections_Empty(t *testing.T) {
	st := newStore()
	revisions := st.ListInspections("no-booking")
	if len(revisions) != 0 {
		t.Fatalf("expected 0 revisions, got %d", len(revisions))
	}
}

func TestListInspections_IsolatedByBooking(t *testing.T) {
	st := newStore()
	st.SaveInspection("b1", models.InspectionRevision{RevisionID: "r1", Stage: "initial"})
	st.SaveInspection("b2", models.InspectionRevision{RevisionID: "r2", Stage: "initial"})
	b1revs := st.ListInspections("b1")
	b2revs := st.ListInspections("b2")
	if len(b1revs) != 1 || len(b2revs) != 1 {
		t.Fatalf("expected inspections isolated by booking, b1=%d b2=%d", len(b1revs), len(b2revs))
	}
}
