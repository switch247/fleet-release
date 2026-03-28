package store

import (
	"sort"
	"time"

	"fleetlease/backend/internal/models"
)

func (s *MemoryStore) SaveUser(u models.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Users[u.ID] = u
	s.UsersByName[u.Username] = u.ID
}

func (s *MemoryStore) GetUserByUsername(username string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.UsersByName[username]
	if !ok {
		return models.User{}, false
	}
	u, ok := s.Users[id]
	return u, ok
}

func (s *MemoryStore) GetUserByID(id string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.Users[id]
	return u, ok
}

func (s *MemoryStore) SaveSession(session models.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sessions[session.ID] = session
}

func (s *MemoryStore) ListUsers() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.User, 0, len(s.Users))
	for _, u := range s.Users {
		out = append(out, u)
	}
	return out
}

func (s *MemoryStore) DeleteUser(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.Users[id]
	if !ok {
		return
	}
	delete(s.Users, id)
	delete(s.UsersByName, u.Username)
}

func (s *MemoryStore) UsernameExists(username string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.UsersByName[username]
	return ok
}

func (s *MemoryStore) HasAdminExcluding(excludeID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, user := range s.Users {
		if excludeID != "" && id == excludeID {
			continue
		}
		for _, role := range user.Roles {
			if role == models.RoleAdmin {
				return true
			}
		}
	}
	return false
}

func (s *MemoryStore) GetSession(id string) (models.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Sessions[id]
	return v, ok
}

func (s *MemoryStore) SaveAuthEvent(event models.AuthEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.UserID == "" {
		s.AuthEvents["anonymous"] = append(s.AuthEvents["anonymous"], event)
		return
	}
	s.AuthEvents[event.UserID] = append(s.AuthEvents[event.UserID], event)
}

func (s *MemoryStore) ListAuthEventsByUser(userID string, limit int) []models.AuthEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := append([]models.AuthEvent{}, s.AuthEvents[userID]...)
	if limit <= 0 || len(events) <= limit {
		return events
	}
	return events[len(events)-limit:]
}

func (s *MemoryStore) SaveCategory(c models.Category) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Categories[c.ID] = c
}

func (s *MemoryStore) ListCategories() []models.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Category, 0, len(s.Categories))
	for _, v := range s.Categories {
		out = append(out, v)
	}
	return out
}

func (s *MemoryStore) GetCategory(id string) (models.Category, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Categories[id]
	return v, ok
}

func (s *MemoryStore) DeleteCategory(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Categories, id)
}

func (s *MemoryStore) SaveListing(l models.Listing) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Listings[l.ID] = l
}

func (s *MemoryStore) ListListings() []models.Listing {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Listing, 0, len(s.Listings))
	for _, v := range s.Listings {
		out = append(out, v)
	}
	return out
}

func (s *MemoryStore) GetListing(id string) (models.Listing, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Listings[id]
	return v, ok
}

func (s *MemoryStore) DeleteListing(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Listings, id)
}

func (s *MemoryStore) SaveBooking(b models.Booking) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Bookings[b.ID] = b
}

func (s *MemoryStore) GetBooking(id string) (models.Booking, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Bookings[id]
	return v, ok
}

func (s *MemoryStore) ListBookings() []models.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Booking, 0, len(s.Bookings))
	for _, v := range s.Bookings {
		out = append(out, v)
	}
	return out
}

func (s *MemoryStore) SaveInspection(bookingID string, revision models.InspectionRevision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Inspections[bookingID] = append(s.Inspections[bookingID], revision)
}

func (s *MemoryStore) ListInspections(bookingID string) []models.InspectionRevision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.InspectionRevision{}, s.Inspections[bookingID]...)
}

func (s *MemoryStore) SaveAttachment(a models.Attachment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	s.Attachments[a.ID] = a
}

func (s *MemoryStore) FindAttachmentByFingerprint(fingerprint string) (models.Attachment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.Attachments {
		if a.Fingerprint == fingerprint {
			return a, true
		}
	}
	return models.Attachment{}, false
}

func (s *MemoryStore) GetAttachment(id string) (models.Attachment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Attachments[id]
	return v, ok
}

func (s *MemoryStore) PurgeAttachmentsOlderThan(cutoff time.Time) []models.Attachment {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := make([]models.Attachment, 0)
	for id, attachment := range s.Attachments {
		if attachment.CreatedAt.IsZero() {
			continue
		}
		if attachment.CreatedAt.Before(cutoff) {
			removed = append(removed, attachment)
			delete(s.Attachments, id)
		}
	}
	return removed
}

func (s *MemoryStore) AppendLedger(bookingID string, e models.LedgerEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Ledgers[bookingID] = append(s.Ledgers[bookingID], e)
}

func (s *MemoryStore) ListLedger(bookingID string) []models.LedgerEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.LedgerEntry{}, s.Ledgers[bookingID]...)
}

func (s *MemoryStore) PurgeLedgerOlderThan(cutoff time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for bookingID, entries := range s.Ledgers {
		kept := make([]models.LedgerEntry, 0, len(entries))
		for _, entry := range entries {
			if entry.CreatedAt.Before(cutoff) {
				removed++
				continue
			}
			kept = append(kept, entry)
		}
		s.Ledgers[bookingID] = kept
	}
	return removed
}

func (s *MemoryStore) SaveComplaint(c models.Complaint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Complaints[c.ID] = c
}

func (s *MemoryStore) GetComplaint(id string) (models.Complaint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Complaints[id]
	return v, ok
}

func (s *MemoryStore) ListComplaints() []models.Complaint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Complaint, 0, len(s.Complaints))
	for _, v := range s.Complaints {
		out = append(out, v)
	}
	return out
}

func (s *MemoryStore) SaveConsultation(c models.Consultation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Consultations[c.Topic] = append(s.Consultations[c.Topic], c)
	s.ConsultByID[c.ID] = c
}

func (s *MemoryStore) GetConsultation(id string) (models.Consultation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.ConsultByID[id]
	return v, ok
}

func (s *MemoryStore) ListConsultationsByBooking(bookingID string) []models.Consultation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Consultation, 0)
	for _, versions := range s.Consultations {
		for _, v := range versions {
			if v.BookingID == bookingID {
				out = append(out, v)
			}
		}
	}
	return out
}

func (s *MemoryStore) ListConsultationsByTopic(topic string) []models.Consultation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Consultation, 0)
	for _, versions := range s.Consultations {
		for _, v := range versions {
			if v.Topic == topic {
				out = append(out, v)
			}
		}
	}
	return out
}

func (s *MemoryStore) SaveConsultationAttachment(a models.ConsultationAttachment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ConsultAttach[a.ConsultationID] = append(s.ConsultAttach[a.ConsultationID], a)
}

func (s *MemoryStore) ListConsultationAttachments(consultationID string) []models.ConsultationAttachment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.ConsultationAttachment{}, s.ConsultAttach[consultationID]...)
}

func (s *MemoryStore) SaveNotification(n models.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.Notifications[n.UserID]
	for i := range items {
		if items[i].Fingerprint == n.Fingerprint {
			items[i] = n
			s.Notifications[n.UserID] = items
			return
		}
	}
	s.Notifications[n.UserID] = append(items, n)
}

func (s *MemoryStore) ListNotifications(userID string) []models.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Notification{}, s.Notifications[userID]...)
}

func (s *MemoryStore) ListAllNotifications() []models.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Notification, 0)
	for _, items := range s.Notifications {
		out = append(out, items...)
	}
	return out
}

func (s *MemoryStore) SaveNotificationTemplate(t models.NotificationTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NotifTemplate[t.ID] = t
}

func (s *MemoryStore) ListNotificationTemplates() []models.NotificationTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.NotificationTemplate, 0, len(s.NotifTemplate))
	for _, v := range s.NotifTemplate {
		out = append(out, v)
	}
	return out
}

func (s *MemoryStore) GetNotificationTemplate(id string) (models.NotificationTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.NotifTemplate[id]
	return v, ok
}

func (s *MemoryStore) SaveRating(r models.Rating) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Ratings[r.BookingID] = append(s.Ratings[r.BookingID], r)
}

func (s *MemoryStore) ListRatings(bookingID string) []models.Rating {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Rating{}, s.Ratings[bookingID]...)
}

func (s *MemoryStore) SaveBackupJob(job models.BackupJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BackupJobs[job.ID] = job
}

func (s *MemoryStore) ListBackupJobs() []models.BackupJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.BackupJob, 0, len(s.BackupJobs))
	for _, v := range s.BackupJobs {
		out = append(out, v)
	}
	return out
}

func (s *MemoryStore) SaveRetentionReport(report models.RetentionReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RetentionReports[report.ID] = report
}

func (s *MemoryStore) ListRetentionReports(limit int) []models.RetentionReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.RetentionReport, 0, len(s.RetentionReports))
	for _, report := range s.RetentionReports {
		out = append(out, report)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *MemoryStore) SavePasswordResetEvidence(e models.PasswordResetEvidence) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResetEvidence[e.ID] = e
}

func (s *MemoryStore) MarkCouponUsed(code, bookingID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.CouponsUsed[code]; exists {
		return false
	}
	s.CouponsUsed[code] = bookingID
	return true
}
