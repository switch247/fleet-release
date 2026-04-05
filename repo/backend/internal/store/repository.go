package store

import (
	"time"

	"fleetlease/backend/internal/models"
)

type Repository interface {
	SaveUser(u models.User)
	GetUserByUsername(username string) (models.User, bool)
	GetUserByID(id string) (models.User, bool)
	ListUsers() []models.User
	DeleteUser(id string)
	UsernameExists(username string) bool
	HasAdminExcluding(excludeID string) bool

	SaveSession(session models.Session)
	GetSession(id string) (models.Session, bool)
	SaveAuthEvent(event models.AuthEvent)
	ListAuthEventsByUser(userID string, limit int) []models.AuthEvent

	SaveCategory(c models.Category)
	ListCategories() []models.Category
	GetCategory(id string) (models.Category, bool)
	DeleteCategory(id string)
	SaveListing(l models.Listing)
	ListListings() []models.Listing
	GetListing(id string) (models.Listing, bool)
	DeleteListing(id string)

	SaveBooking(b models.Booking)
	GetBooking(id string) (models.Booking, bool)
	ListBookings() []models.Booking

	SaveInspection(bookingID string, revision models.InspectionRevision)
	ListInspections(bookingID string) []models.InspectionRevision

	SaveAttachment(a models.Attachment)
	FindAttachmentByFingerprint(fingerprint string) (models.Attachment, bool)
	GetAttachment(id string) (models.Attachment, bool)
	PurgeAttachmentsOlderThan(cutoff time.Time) []models.Attachment

	AppendLedger(bookingID string, e models.LedgerEntry)
	ListLedger(bookingID string) []models.LedgerEntry
	PurgeLedgerOlderThan(cutoff time.Time) int

	SaveComplaint(c models.Complaint)
	GetComplaint(id string) (models.Complaint, bool)
	ListComplaints() []models.Complaint

	SaveConsultation(c models.Consultation)
	GetConsultation(id string) (models.Consultation, bool)
	ListConsultationsByBooking(bookingID string) []models.Consultation
	ListConsultationsByThread(threadID string) []models.Consultation
	ListConsultationsByTopic(topic string) []models.Consultation
	SaveConsultationAttachment(a models.ConsultationAttachment)
	ListConsultationAttachments(consultationID string) []models.ConsultationAttachment

	SaveNotification(n models.Notification)
	ListNotifications(userID string) []models.Notification
	ListAllNotifications() []models.Notification
	SaveNotificationTemplate(t models.NotificationTemplate)
	ListNotificationTemplates() []models.NotificationTemplate
	GetNotificationTemplate(id string) (models.NotificationTemplate, bool)

	SaveRating(r models.Rating)
	ListRatings(bookingID string) []models.Rating

	SaveBackupJob(job models.BackupJob)
	ListBackupJobs() []models.BackupJob
	SavePasswordResetEvidence(e models.PasswordResetEvidence)
	SaveRetentionReport(r models.RetentionReport)
	ListRetentionReports(limit int) []models.RetentionReport

	MarkCouponUsed(code, bookingID string) bool
}
