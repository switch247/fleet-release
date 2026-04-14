package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
)

// ---------------------------------------------------------------------------
// BackupJobs
// ---------------------------------------------------------------------------

func TestSaveAndListBackupJobs(t *testing.T) {
	st := newStore()
	job := models.BackupJob{
		ID:          "job1",
		Type:        "manual",
		Status:      "running",
		RequestedBy: "admin1",
		CreatedAt:   time.Now().UTC(),
	}
	st.SaveBackupJob(job)
	jobs := st.ListBackupJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Type != "manual" {
		t.Fatalf("expected type manual, got %s", jobs[0].Type)
	}
}

func TestListBackupJobs_Empty(t *testing.T) {
	st := newStore()
	jobs := st.ListBackupJobs()
	if len(jobs) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestBackupJobStatusTransition(t *testing.T) {
	st := newStore()
	st.SaveBackupJob(models.BackupJob{ID: "job1", Status: "running"})
	j, _ := func() (models.BackupJob, bool) {
		for _, j := range st.ListBackupJobs() {
			if j.ID == "job1" {
				return j, true
			}
		}
		return models.BackupJob{}, false
	}()
	j.Status = "completed"
	j.Artifact = "/backups/2026-01-15.sql.gz"
	j.FinishedAt = time.Now().UTC()
	st.SaveBackupJob(j)
	// verify update
	jobs := st.ListBackupJobs()
	var found models.BackupJob
	for _, bj := range jobs {
		if bj.ID == "job1" {
			found = bj
		}
	}
	if found.Status != "completed" {
		t.Fatalf("expected completed, got %s", found.Status)
	}
	if found.Artifact != "/backups/2026-01-15.sql.gz" {
		t.Fatalf("expected artifact path, got %s", found.Artifact)
	}
}

// ---------------------------------------------------------------------------
// PasswordResetEvidence
// ---------------------------------------------------------------------------

func TestSavePasswordResetEvidence(t *testing.T) {
	st := newStore()
	e := models.PasswordResetEvidence{
		ID:           "ev1",
		TargetUserID: "u1",
		CheckedBy:    "admin1",
		Method:       "government_id",
		EvidenceRef:  "file-001",
		Reason:       "user lost access",
		CreatedAt:    time.Now().UTC(),
	}
	st.SavePasswordResetEvidence(e)
	// No getter in repository interface; just verify no panic
}

// ---------------------------------------------------------------------------
// RetentionReports
// ---------------------------------------------------------------------------

func TestSaveAndListRetentionReports(t *testing.T) {
	st := newStore()
	r := models.RetentionReport{
		ID:                 "rr1",
		AttachmentsDeleted: 5,
		LedgerDeleted:      12,
		FileDeleteErrors:   0,
		CreatedAt:          time.Now().UTC(),
	}
	st.SaveRetentionReport(r)
	reports := st.ListRetentionReports(10)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].AttachmentsDeleted != 5 {
		t.Fatalf("expected 5 attachments deleted, got %d", reports[0].AttachmentsDeleted)
	}
}

func TestListRetentionReports_LimitApplied(t *testing.T) {
	st := newStore()
	base := time.Now().UTC()
	for i := 0; i < 5; i++ {
		st.SaveRetentionReport(models.RetentionReport{
			ID:        string(rune('a' + i)),
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}
	reports := st.ListRetentionReports(3)
	if len(reports) != 3 {
		t.Fatalf("expected 3 with limit, got %d", len(reports))
	}
}

func TestListRetentionReports_OrderedNewestFirst(t *testing.T) {
	st := newStore()
	base := time.Now().UTC()
	st.SaveRetentionReport(models.RetentionReport{ID: "old", AttachmentsDeleted: 1, CreatedAt: base.Add(-2 * time.Hour)})
	st.SaveRetentionReport(models.RetentionReport{ID: "new", AttachmentsDeleted: 99, CreatedAt: base})
	reports := st.ListRetentionReports(10)
	if reports[0].ID != "new" {
		t.Fatalf("expected newest report first, got %s", reports[0].ID)
	}
}

// ---------------------------------------------------------------------------
// Coupon lifecycle
// ---------------------------------------------------------------------------

func TestMarkCouponUsed_FirstUseSucceeds(t *testing.T) {
	st := newStore()
	ok := st.MarkCouponUsed("SAVE10", "b1")
	if !ok {
		t.Fatal("expected first coupon use to succeed")
	}
}

func TestMarkCouponUsed_SecondUseFails(t *testing.T) {
	st := newStore()
	st.MarkCouponUsed("SAVE10", "b1")
	ok := st.MarkCouponUsed("SAVE10", "b2")
	if ok {
		t.Fatal("expected second coupon use to fail (already used)")
	}
}

func TestMarkCouponUsed_DifferentCodesIndependent(t *testing.T) {
	st := newStore()
	ok1 := st.MarkCouponUsed("CODE-A", "b1")
	ok2 := st.MarkCouponUsed("CODE-B", "b2")
	if !ok1 || !ok2 {
		t.Fatal("expected different coupon codes to be independent")
	}
}

func TestMarkCouponUsed_SameBookingDifferentCoupons(t *testing.T) {
	st := newStore()
	ok1 := st.MarkCouponUsed("COUPON-X", "b1")
	ok2 := st.MarkCouponUsed("COUPON-Y", "b1")
	if !ok1 || !ok2 {
		t.Fatal("expected same booking to be able to use different coupons")
	}
}
