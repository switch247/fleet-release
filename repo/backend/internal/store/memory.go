package store

import (
	"sync"

	"fleetlease/backend/internal/models"
)

type MemoryStore struct {
	mu sync.RWMutex

	Users            map[string]models.User
	UsersByName      map[string]string
	Sessions         map[string]models.Session
	AuthEvents       map[string][]models.AuthEvent
	Categories       map[string]models.Category
	Listings         map[string]models.Listing
	Bookings         map[string]models.Booking
	Inspections      map[string][]models.InspectionRevision
	Attachments      map[string]models.Attachment
	Ledgers          map[string][]models.LedgerEntry
	Complaints       map[string]models.Complaint
	Consultations    map[string][]models.Consultation
	ConsultByID      map[string]models.Consultation
	ConsultAttach    map[string][]models.ConsultationAttachment
	Notifications    map[string][]models.Notification
	NotifTemplate    map[string]models.NotificationTemplate
	Ratings          map[string][]models.Rating
	BackupJobs       map[string]models.BackupJob
	ResetEvidence    map[string]models.PasswordResetEvidence
	CouponsUsed      map[string]string
	CouponCatalog    map[string]float64
	RetentionReports map[string]models.RetentionReport
	ReconcileKeys    map[string]struct{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Users:            map[string]models.User{},
		UsersByName:      map[string]string{},
		Sessions:         map[string]models.Session{},
		AuthEvents:       map[string][]models.AuthEvent{},
		Categories:       map[string]models.Category{},
		Listings:         map[string]models.Listing{},
		Bookings:         map[string]models.Booking{},
		Inspections:      map[string][]models.InspectionRevision{},
		Attachments:      map[string]models.Attachment{},
		Ledgers:          map[string][]models.LedgerEntry{},
		Complaints:       map[string]models.Complaint{},
		Consultations:    map[string][]models.Consultation{},
		ConsultByID:      map[string]models.Consultation{},
		ConsultAttach:    map[string][]models.ConsultationAttachment{},
		Notifications:    map[string][]models.Notification{},
		NotifTemplate:    map[string]models.NotificationTemplate{},
		Ratings:          map[string][]models.Rating{},
		BackupJobs:       map[string]models.BackupJob{},
		ResetEvidence:    map[string]models.PasswordResetEvidence{},
		CouponsUsed:      map[string]string{},
		CouponCatalog: map[string]float64{
			"DEMO10": 0.10,
			"SAVE20": 0.20,
		},
		RetentionReports: map[string]models.RetentionReport{},
		ReconcileKeys:    map[string]struct{}{},
	}
}

func (s *MemoryStore) MarkReconcileApplied(userID, key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	composite := userID + ":" + key
	if _, exists := s.ReconcileKeys[composite]; exists {
		return true
	}
	s.ReconcileKeys[composite] = struct{}{}
	return false
}

