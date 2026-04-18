package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fleetlease/backend/internal/models"

	"github.com/labstack/echo/v4"
)

// fakeRepo implements store.Repository with only GetUserByID and no-ops for others
type fakeRepo struct {
	user models.User
	ok   bool
}

func (f *fakeRepo) GetUserByID(id string) (models.User, bool) { return f.user, f.ok }
func (f *fakeRepo) SaveUser(u models.User)                    {}
func (f *fakeRepo) GetUserByUsername(username string) (models.User, bool) {
	return models.User{}, false
}
func (f *fakeRepo) ListUsers() []models.User                                            { return nil }
func (f *fakeRepo) DeleteUser(id string)                                                {}
func (f *fakeRepo) UsernameExists(username string) bool                                 { return false }
func (f *fakeRepo) HasAdminExcluding(excludeID string) bool                             { return false }
func (f *fakeRepo) SaveSession(session models.Session)                                  {}
func (f *fakeRepo) GetSession(id string) (models.Session, bool)                         { return models.Session{}, false }
func (f *fakeRepo) SaveAuthEvent(event models.AuthEvent)                                {}
func (f *fakeRepo) ListAuthEventsByUser(userID string, limit int) []models.AuthEvent    { return nil }
func (f *fakeRepo) SaveCategory(c models.Category)                                      {}
func (f *fakeRepo) ListCategories() []models.Category                                   { return nil }
func (f *fakeRepo) GetCategory(id string) (models.Category, bool)                       { return models.Category{}, false }
func (f *fakeRepo) DeleteCategory(id string)                                            {}
func (f *fakeRepo) SaveListing(l models.Listing)                                        {}
func (f *fakeRepo) ListListings() []models.Listing                                      { return nil }
func (f *fakeRepo) GetListing(id string) (models.Listing, bool)                         { return models.Listing{}, false }
func (f *fakeRepo) DeleteListing(id string)                                             {}
func (f *fakeRepo) SaveBooking(b models.Booking)                                        {}
func (f *fakeRepo) GetBooking(id string) (models.Booking, bool)                         { return models.Booking{}, false }
func (f *fakeRepo) ListBookings() []models.Booking                                      { return nil }
func (f *fakeRepo) SaveInspection(bookingID string, revision models.InspectionRevision) {}
func (f *fakeRepo) ListInspections(bookingID string) []models.InspectionRevision        { return nil }
func (f *fakeRepo) SaveAttachment(a models.Attachment)                                  {}
func (f *fakeRepo) FindAttachmentByFingerprint(fingerprint string) (models.Attachment, bool) {
	return models.Attachment{}, false
}
func (f *fakeRepo) GetAttachment(id string) (models.Attachment, bool) {
	return models.Attachment{}, false
}
func (f *fakeRepo) PurgeAttachmentsOlderThan(cutoff time.Time) []models.Attachment { return nil }
func (f *fakeRepo) AppendLedger(bookingID string, e models.LedgerEntry)            {}
func (f *fakeRepo) ListLedger(bookingID string) []models.LedgerEntry               { return nil }
func (f *fakeRepo) PurgeLedgerOlderThan(cutoff time.Time) int                      { return 0 }
func (f *fakeRepo) SaveComplaint(c models.Complaint)                               {}
func (f *fakeRepo) GetComplaint(id string) (models.Complaint, bool)                { return models.Complaint{}, false }
func (f *fakeRepo) ListComplaints() []models.Complaint                             { return nil }
func (f *fakeRepo) SaveConsultation(c models.Consultation)                         {}
func (f *fakeRepo) GetConsultation(id string) (models.Consultation, bool) {
	return models.Consultation{}, false
}
func (f *fakeRepo) ListConsultationsByBooking(bookingID string) []models.Consultation { return nil }
func (f *fakeRepo) ListConsultationsByThread(threadID string) []models.Consultation   { return nil }
func (f *fakeRepo) ListConsultationsByTopic(topic string) []models.Consultation       { return nil }
func (f *fakeRepo) SaveConsultationAttachment(a models.ConsultationAttachment)        {}
func (f *fakeRepo) ListConsultationAttachments(consultationID string) []models.ConsultationAttachment {
	return nil
}
func (f *fakeRepo) SaveNotification(n models.Notification)                   {}
func (f *fakeRepo) ListNotifications(userID string) []models.Notification    { return nil }
func (f *fakeRepo) ListAllNotifications() []models.Notification              { return nil }
func (f *fakeRepo) SaveNotificationTemplate(t models.NotificationTemplate)   {}
func (f *fakeRepo) ListNotificationTemplates() []models.NotificationTemplate { return nil }
func (f *fakeRepo) GetNotificationTemplate(id string) (models.NotificationTemplate, bool) {
	return models.NotificationTemplate{}, false
}
func (f *fakeRepo) SaveRating(r models.Rating)                               {}
func (f *fakeRepo) ListRatings(bookingID string) []models.Rating             { return nil }
func (f *fakeRepo) SaveBackupJob(job models.BackupJob)                       {}
func (f *fakeRepo) ListBackupJobs() []models.BackupJob                       { return nil }
func (f *fakeRepo) SavePasswordResetEvidence(e models.PasswordResetEvidence) {}
func (f *fakeRepo) SaveRetentionReport(r models.RetentionReport)             {}
func (f *fakeRepo) ListRetentionReports(limit int) []models.RetentionReport  { return nil }
func (f *fakeRepo) MarkCouponUsed(code, bookingID string) bool               { return false }
func (f *fakeRepo) GetCouponDiscount(code string) (float64, bool)            { return 0, false }
func (f *fakeRepo) MarkReconcileApplied(userID, key string) bool             { return false }

func TestRequireMFAForRoles_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxUserID, "u1")
	c.Set(CtxRoles, []models.Role{"admin"})
	st := &fakeRepo{user: models.User{TOTPEnabled: true}, ok: true}
	mw := RequireMFAForRoles(st, "admin")
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != 200 {
		t.Fatalf("expected ok")
	}
}

func TestRequireMFAForRoles_MissingRoles(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := RequireMFAForRoles(&fakeRepo{}, "admin")
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}

func TestRequireMFAForRoles_UserNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxUserID, "u1")
	c.Set(CtxRoles, []models.Role{"admin"})
	st := &fakeRepo{ok: false}
	mw := RequireMFAForRoles(st, "admin")
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized")
	}
}

func TestRequireMFAForRoles_MFARequired(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxUserID, "u1")
	c.Set(CtxRoles, []models.Role{"admin"})
	st := &fakeRepo{user: models.User{TOTPEnabled: false}, ok: true}
	mw := RequireMFAForRoles(st, "admin")
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}
