package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"fleetlease/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) SaveUser(u models.User) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO users (id, username, password_hash, government_id_enc, failed_attempts, locked_until, totp_secret, totp_enabled, email, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,COALESCE((SELECT created_at FROM users WHERE id=$1), NOW()))
ON CONFLICT (id) DO UPDATE SET
username=EXCLUDED.username,
password_hash=EXCLUDED.password_hash,
government_id_enc=EXCLUDED.government_id_enc,
failed_attempts=EXCLUDED.failed_attempts,
locked_until=EXCLUDED.locked_until,
totp_secret=EXCLUDED.totp_secret,
totp_enabled=EXCLUDED.totp_enabled,
email=EXCLUDED.email`,
		u.ID, u.Username, u.PasswordHash, u.GovernmentIDEnc, u.FailedAttempts, nullableTime(u.LockedUntil), u.TOTPSecret, u.TOTPEnabled, u.Email,
	)
	_, _ = s.pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, u.ID)
	for _, r := range u.Roles {
		_, _ = s.pool.Exec(ctx, `INSERT INTO user_roles (user_id, role) VALUES ($1,$2)`, u.ID, string(r))
	}
}

func (s *PostgresStore) loadRoles(ctx context.Context, userID string) []models.Role {
	rows, err := s.pool.Query(ctx, `SELECT role FROM user_roles WHERE user_id=$1`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	roles := make([]models.Role, 0)
	for rows.Next() {
		var r string
		if rows.Scan(&r) == nil {
			roles = append(roles, models.Role(r))
		}
	}
	return roles
}

func (s *PostgresStore) GetUserByUsername(username string) (models.User, bool) {
	ctx := context.Background()
	var u models.User
	var locked sql.NullTime
	err := s.pool.QueryRow(ctx, `SELECT id, username, COALESCE(email,''), password_hash, government_id_enc, failed_attempts, locked_until, COALESCE(totp_secret,''), COALESCE(totp_enabled,false) FROM users WHERE username=$1`, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GovernmentIDEnc, &u.FailedAttempts, &locked, &u.TOTPSecret, &u.TOTPEnabled)
	if err != nil {
		return models.User{}, false
	}
	if locked.Valid {
		u.LockedUntil = locked.Time
	}
	u.Roles = s.loadRoles(ctx, u.ID)
	return u, true
}

func (s *PostgresStore) GetUserByID(id string) (models.User, bool) {
	ctx := context.Background()
	var u models.User
	var locked sql.NullTime
	err := s.pool.QueryRow(ctx, `SELECT id, username, COALESCE(email,''), password_hash, government_id_enc, failed_attempts, locked_until, COALESCE(totp_secret,''), COALESCE(totp_enabled,false) FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GovernmentIDEnc, &u.FailedAttempts, &locked, &u.TOTPSecret, &u.TOTPEnabled)
	if err != nil {
		return models.User{}, false
	}
	if locked.Valid {
		u.LockedUntil = locked.Time
	}
	u.Roles = s.loadRoles(ctx, u.ID)
	return u, true
}

func (s *PostgresStore) ListUsers() []models.User {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id FROM users ORDER BY username`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	users := make([]models.User, 0)
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			if u, ok := s.GetUserByID(id); ok {
				users = append(users, u)
			}
		}
	}
	return users
}

func (s *PostgresStore) DeleteUser(id string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
}

func (s *PostgresStore) UsernameExists(username string) bool {
	ctx := context.Background()
	var count int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(1) FROM users WHERE username=$1`, username).Scan(&count)
	return count > 0
}

func (s *PostgresStore) HasAdminExcluding(excludeID string) bool {
	ctx := context.Background()
	var count int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(1) FROM user_roles WHERE role='admin' AND user_id <> $1`, excludeID).Scan(&count)
	return count > 0
}

func (s *PostgresStore) SaveSession(session models.Session) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO sessions (id, user_id, issued_at, last_seen_at, absolute_exp, revoked)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (id) DO UPDATE SET
last_seen_at=EXCLUDED.last_seen_at,
absolute_exp=EXCLUDED.absolute_exp,
revoked=EXCLUDED.revoked`,
		session.ID, session.UserID, session.IssuedAt, session.LastSeenAt, session.AbsoluteExp, session.Revoked,
	)
}

func (s *PostgresStore) GetSession(id string) (models.Session, bool) {
	ctx := context.Background()
	var out models.Session
	err := s.pool.QueryRow(ctx, `SELECT id, user_id, issued_at, last_seen_at, absolute_exp, revoked FROM sessions WHERE id=$1`, id).
		Scan(&out.ID, &out.UserID, &out.IssuedAt, &out.LastSeenAt, &out.AbsoluteExp, &out.Revoked)
	if err != nil {
		return models.Session{}, false
	}
	return out, true
}

func (s *PostgresStore) SaveAuthEvent(event models.AuthEvent) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO auth_events (id,user_id,username,ip,event_type,created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		event.ID, nullIfEmpty(event.UserID), event.Username, event.IP, event.EventType, event.CreatedAt,
	)
}

func (s *PostgresStore) ListAuthEventsByUser(userID string, limit int) []models.AuthEvent {
	ctx := context.Background()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `SELECT id,COALESCE(user_id::text,''),COALESCE(username,''),COALESCE(ip,''),event_type,created_at FROM auth_events WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.AuthEvent, 0)
	for rows.Next() {
		var e models.AuthEvent
		if rows.Scan(&e.ID, &e.UserID, &e.Username, &e.IP, &e.EventType, &e.CreatedAt) == nil {
			out = append(out, e)
		}
	}
	return out
}

func (s *PostgresStore) SaveCategory(c models.Category) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO categories (id,name) VALUES ($1,$2) ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name`, c.ID, c.Name)
}

func (s *PostgresStore) ListCategories() []models.Category {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,name FROM categories ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if rows.Scan(&c.ID, &c.Name) == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) GetCategory(id string) (models.Category, bool) {
	ctx := context.Background()
	var c models.Category
	err := s.pool.QueryRow(ctx, `SELECT id,name FROM categories WHERE id=$1`, id).Scan(&c.ID, &c.Name)
	if err != nil {
		return models.Category{}, false
	}
	return c, true
}

func (s *PostgresStore) DeleteCategory(id string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
}

func (s *PostgresStore) SaveListing(l models.Listing) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO listings (id,category_id,provider_id,spu,sku,name,included_miles,deposit,available)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (id) DO UPDATE SET
category_id=EXCLUDED.category_id,
provider_id=EXCLUDED.provider_id,
spu=EXCLUDED.spu,
sku=EXCLUDED.sku,
name=EXCLUDED.name,
included_miles=EXCLUDED.included_miles,
deposit=EXCLUDED.deposit,
available=EXCLUDED.available`,
		l.ID, l.CategoryID, nullableUUID(l.ProviderID), l.SPU, l.SKU, l.Name, l.IncludedMiles, l.Deposit, l.Available,
	)
}

func (s *PostgresStore) ListListings() []models.Listing {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,category_id,COALESCE(provider_id::text,''),spu,sku,name,included_miles,deposit,available FROM listings ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Listing, 0)
	for rows.Next() {
		var l models.Listing
		if rows.Scan(&l.ID, &l.CategoryID, &l.ProviderID, &l.SPU, &l.SKU, &l.Name, &l.IncludedMiles, &l.Deposit, &l.Available) == nil {
			out = append(out, l)
		}
	}
	return out
}

func (s *PostgresStore) GetListing(id string) (models.Listing, bool) {
	ctx := context.Background()
	var l models.Listing
	err := s.pool.QueryRow(ctx, `SELECT id,category_id,COALESCE(provider_id::text,''),spu,sku,name,included_miles,deposit,available FROM listings WHERE id=$1`, id).
		Scan(&l.ID, &l.CategoryID, &l.ProviderID, &l.SPU, &l.SKU, &l.Name, &l.IncludedMiles, &l.Deposit, &l.Available)
	if err != nil {
		return models.Listing{}, false
	}
	return l, true
}

func (s *PostgresStore) DeleteListing(id string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `DELETE FROM listings WHERE id=$1`, id)
}

func (s *PostgresStore) SaveBooking(b models.Booking) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO bookings (id,customer_id,provider_id,listing_id,status,estimated_amount,deposit_amount,start_at,end_at,odo_start,odo_end,coupon_code)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (id) DO UPDATE SET
status=EXCLUDED.status,
provider_id=EXCLUDED.provider_id,
estimated_amount=EXCLUDED.estimated_amount,
deposit_amount=EXCLUDED.deposit_amount,
start_at=EXCLUDED.start_at,
end_at=EXCLUDED.end_at,
odo_start=EXCLUDED.odo_start,
odo_end=EXCLUDED.odo_end,
coupon_code=EXCLUDED.coupon_code`,
		b.ID, b.CustomerID, nullableUUID(b.ProviderID), b.ListingID, b.Status, b.EstimatedAmount, b.DepositAmount, b.StartAt, b.EndAt, b.OdoStart, b.OdoEnd, b.CouponCode,
	)
}

func (s *PostgresStore) GetBooking(id string) (models.Booking, bool) {
	ctx := context.Background()
	var b models.Booking
	err := s.pool.QueryRow(ctx, `SELECT id,customer_id,COALESCE(provider_id::text,''),listing_id,COALESCE(coupon_code,''),start_at,end_at,odo_start,odo_end,status,estimated_amount,deposit_amount FROM bookings WHERE id=$1`, id).
		Scan(&b.ID, &b.CustomerID, &b.ProviderID, &b.ListingID, &b.CouponCode, &b.StartAt, &b.EndAt, &b.OdoStart, &b.OdoEnd, &b.Status, &b.EstimatedAmount, &b.DepositAmount)
	if err != nil {
		return models.Booking{}, false
	}
	return b, true
}

func (s *PostgresStore) ListBookings() []models.Booking {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,customer_id,COALESCE(provider_id::text,''),listing_id,COALESCE(coupon_code,''),start_at,end_at,odo_start,odo_end,status,estimated_amount,deposit_amount FROM bookings ORDER BY start_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Booking, 0)
	for rows.Next() {
		var b models.Booking
		if rows.Scan(&b.ID, &b.CustomerID, &b.ProviderID, &b.ListingID, &b.CouponCode, &b.StartAt, &b.EndAt, &b.OdoStart, &b.OdoEnd, &b.Status, &b.EstimatedAmount, &b.DepositAmount) == nil {
			out = append(out, b)
		}
	}
	return out
}

func (s *PostgresStore) SaveInspection(bookingID string, revision models.InspectionRevision) {
	ctx := context.Background()
	payload, _ := json.Marshal(revision)
	_, _ = s.pool.Exec(ctx, `INSERT INTO inspection_revisions (id,booking_id,stage,payload,prev_hash,hash,created_at,created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		revision.RevisionID, bookingID, revision.Stage, payload, revision.PrevHash, revision.Hash, revision.CreatedAt, revision.CreatedBy,
	)
}

func (s *PostgresStore) ListInspections(bookingID string) []models.InspectionRevision {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT payload FROM inspection_revisions WHERE booking_id=$1 ORDER BY created_at ASC`, bookingID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.InspectionRevision, 0)
	for rows.Next() {
		var payload []byte
		if rows.Scan(&payload) == nil {
			var rev models.InspectionRevision
			if json.Unmarshal(payload, &rev) == nil {
				out = append(out, rev)
			}
		}
	}
	return out
}

func (s *PostgresStore) SaveAttachment(a models.Attachment) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO attachments (id,booking_id,type,path,size_bytes,checksum,fingerprint)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (id) DO UPDATE SET
path=EXCLUDED.path,size_bytes=EXCLUDED.size_bytes,checksum=EXCLUDED.checksum,fingerprint=EXCLUDED.fingerprint`,
		a.ID, a.BookingID, a.Type, a.Path, a.SizeBytes, a.Checksum, a.Fingerprint,
	)
}

func (s *PostgresStore) FindAttachmentByFingerprint(fingerprint string) (models.Attachment, bool) {
	ctx := context.Background()
	var a models.Attachment
	err := s.pool.QueryRow(ctx, `SELECT id,booking_id,type,path,size_bytes,checksum,fingerprint FROM attachments WHERE fingerprint=$1`, fingerprint).
		Scan(&a.ID, &a.BookingID, &a.Type, &a.Path, &a.SizeBytes, &a.Checksum, &a.Fingerprint)
	if err != nil {
		return models.Attachment{}, false
	}
	return a, true
}

func (s *PostgresStore) GetAttachment(id string) (models.Attachment, bool) {
	ctx := context.Background()
	var a models.Attachment
	err := s.pool.QueryRow(ctx, `SELECT id,booking_id,type,path,size_bytes,checksum,fingerprint FROM attachments WHERE id=$1`, id).
		Scan(&a.ID, &a.BookingID, &a.Type, &a.Path, &a.SizeBytes, &a.Checksum, &a.Fingerprint)
	if err != nil {
		return models.Attachment{}, false
	}
	return a, true
}

func (s *PostgresStore) AppendLedger(bookingID string, e models.LedgerEntry) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO ledger_entries (id,booking_id,entry_type,amount,description,prev_hash,hash,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, bookingID, e.Type, e.Amount, e.Description, e.PrevHash, e.Hash, e.CreatedAt,
	)
}

func (s *PostgresStore) ListLedger(bookingID string) []models.LedgerEntry {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,booking_id,entry_type,amount,description,created_at,COALESCE(prev_hash,''),hash FROM ledger_entries WHERE booking_id=$1 ORDER BY created_at ASC`, bookingID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.LedgerEntry, 0)
	for rows.Next() {
		var l models.LedgerEntry
		if rows.Scan(&l.ID, &l.BookingID, &l.Type, &l.Amount, &l.Description, &l.CreatedAt, &l.PrevHash, &l.Hash) == nil {
			out = append(out, l)
		}
	}
	return out
}

func (s *PostgresStore) SaveComplaint(c models.Complaint) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO complaints (id,booking_id,opened_by,status,outcome,created_at)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status,outcome=EXCLUDED.outcome`,
		c.ID, c.BookingID, c.OpenedBy, c.Status, c.Outcome, c.CreatedAt,
	)
}

func (s *PostgresStore) GetComplaint(id string) (models.Complaint, bool) {
	ctx := context.Background()
	var c models.Complaint
	err := s.pool.QueryRow(ctx, `SELECT id,booking_id,opened_by,status,COALESCE(outcome,''),created_at FROM complaints WHERE id=$1`, id).
		Scan(&c.ID, &c.BookingID, &c.OpenedBy, &c.Status, &c.Outcome, &c.CreatedAt)
	if err != nil {
		return models.Complaint{}, false
	}
	return c, true
}

func (s *PostgresStore) ListComplaints() []models.Complaint {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,booking_id,opened_by,status,COALESCE(outcome,''),created_at FROM complaints ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Complaint, 0)
	for rows.Next() {
		var c models.Complaint
		if rows.Scan(&c.ID, &c.BookingID, &c.OpenedBy, &c.Status, &c.Outcome, &c.CreatedAt) == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) SaveConsultation(c models.Consultation) {
	ctx := context.Background()
	consultationKey := c.ID
	if c.Topic != "" {
		consultationKey = c.Topic
	}
	_, _ = s.pool.Exec(ctx, `INSERT INTO consultation_versions (id,consultation_key,booking_id,version,topic,key_points,recommendation,follow_up,visibility,created_by,created_at,change_reason) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, consultationKey, c.BookingID, c.Version, c.Topic, c.KeyPoints, c.Recommendation, c.FollowUp, c.Visibility, c.CreatedBy, c.CreatedAt, c.ChangeReason,
	)
}

func (s *PostgresStore) GetConsultation(id string) (models.Consultation, bool) {
	ctx := context.Background()
	var c models.Consultation
	err := s.pool.QueryRow(ctx, `SELECT id,COALESCE(booking_id::text,''),topic,key_points,recommendation,follow_up,COALESCE(visibility,'csa_admin'),COALESCE(change_reason,''),version,created_by,created_at FROM consultation_versions WHERE id=$1`, id).
		Scan(&c.ID, &c.BookingID, &c.Topic, &c.KeyPoints, &c.Recommendation, &c.FollowUp, &c.Visibility, &c.ChangeReason, &c.Version, &c.CreatedBy, &c.CreatedAt)
	if err != nil {
		return models.Consultation{}, false
	}
	return c, true
}

func (s *PostgresStore) ListConsultationsByBooking(bookingID string) []models.Consultation {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,COALESCE(booking_id::text,''),topic,key_points,recommendation,follow_up,COALESCE(visibility,'csa_admin'),COALESCE(change_reason,''),version,created_by,created_at FROM consultation_versions WHERE booking_id=$1 ORDER BY created_at ASC`, bookingID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Consultation, 0)
	for rows.Next() {
		var c models.Consultation
		if rows.Scan(&c.ID, &c.BookingID, &c.Topic, &c.KeyPoints, &c.Recommendation, &c.FollowUp, &c.Visibility, &c.ChangeReason, &c.Version, &c.CreatedBy, &c.CreatedAt) == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) ListConsultationsByTopic(topic string) []models.Consultation {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,COALESCE(booking_id::text,''),topic,key_points,recommendation,follow_up,COALESCE(visibility,'csa_admin'),COALESCE(change_reason,''),version,created_by,created_at FROM consultation_versions WHERE topic=$1 ORDER BY version ASC`, topic)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Consultation, 0)
	for rows.Next() {
		var c models.Consultation
		if rows.Scan(&c.ID, &c.BookingID, &c.Topic, &c.KeyPoints, &c.Recommendation, &c.FollowUp, &c.Visibility, &c.ChangeReason, &c.Version, &c.CreatedBy, &c.CreatedAt) == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *PostgresStore) SaveConsultationAttachment(a models.ConsultationAttachment) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO consultation_attachments (id,consultation_id,attachment_id,created_by,created_at) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO UPDATE SET attachment_id=EXCLUDED.attachment_id`,
		a.ID, a.ConsultationID, a.AttachmentID, a.CreatedBy, a.CreatedAt,
	)
}

func (s *PostgresStore) ListConsultationAttachments(consultationID string) []models.ConsultationAttachment {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,consultation_id,attachment_id,created_by,created_at FROM consultation_attachments WHERE consultation_id=$1 ORDER BY created_at ASC`, consultationID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.ConsultationAttachment, 0)
	for rows.Next() {
		var a models.ConsultationAttachment
		if rows.Scan(&a.ID, &a.ConsultationID, &a.AttachmentID, &a.CreatedBy, &a.CreatedAt) == nil {
			out = append(out, a)
		}
	}
	return out
}

func (s *PostgresStore) SaveNotification(n models.Notification) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO notifications (id,user_id,template_id,title,body,status,attempts,fingerprint,delivered_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (user_id,fingerprint) DO UPDATE SET status=EXCLUDED.status, attempts=EXCLUDED.attempts, delivered_at=EXCLUDED.delivered_at`,
		n.ID, n.UserID, nullableUUID(n.TemplateID), n.Title, n.Body, n.Status, n.Attempts, n.Fingerprint, nullableTime(n.DeliveredAt),
	)
}

func (s *PostgresStore) ListNotifications(userID string) []models.Notification {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,user_id,COALESCE(template_id::text,''),title,body,COALESCE(status,'queued'),COALESCE(attempts,0),fingerprint,COALESCE(delivered_at, NOW()) FROM notifications WHERE user_id=$1 ORDER BY delivered_at DESC`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Notification, 0)
	for rows.Next() {
		var n models.Notification
		if rows.Scan(&n.ID, &n.UserID, &n.TemplateID, &n.Title, &n.Body, &n.Status, &n.Attempts, &n.Fingerprint, &n.DeliveredAt) == nil {
			out = append(out, n)
		}
	}
	return out
}

func (s *PostgresStore) ListAllNotifications() []models.Notification {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,user_id,COALESCE(template_id::text,''),title,body,COALESCE(status,'queued'),COALESCE(attempts,0),fingerprint,COALESCE(delivered_at, NOW()) FROM notifications ORDER BY delivered_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Notification, 0)
	for rows.Next() {
		var n models.Notification
		if rows.Scan(&n.ID, &n.UserID, &n.TemplateID, &n.Title, &n.Body, &n.Status, &n.Attempts, &n.Fingerprint, &n.DeliveredAt) == nil {
			out = append(out, n)
		}
	}
	return out
}

func (s *PostgresStore) SaveNotificationTemplate(t models.NotificationTemplate) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `
INSERT INTO notification_templates (id,name,title,body,channel,enabled,created_by,modified_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (id) DO UPDATE SET
name=EXCLUDED.name,title=EXCLUDED.title,body=EXCLUDED.body,channel=EXCLUDED.channel,enabled=EXCLUDED.enabled,modified_at=EXCLUDED.modified_at`,
		t.ID, t.Name, t.Title, t.Body, t.Channel, t.Enabled, t.CreatedBy, t.ModifiedAt,
	)
}

func (s *PostgresStore) ListNotificationTemplates() []models.NotificationTemplate {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,name,title,body,channel,enabled,created_by,modified_at FROM notification_templates ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.NotificationTemplate, 0)
	for rows.Next() {
		var t models.NotificationTemplate
		if rows.Scan(&t.ID, &t.Name, &t.Title, &t.Body, &t.Channel, &t.Enabled, &t.CreatedBy, &t.ModifiedAt) == nil {
			out = append(out, t)
		}
	}
	return out
}

func (s *PostgresStore) GetNotificationTemplate(id string) (models.NotificationTemplate, bool) {
	ctx := context.Background()
	var t models.NotificationTemplate
	err := s.pool.QueryRow(ctx, `SELECT id,name,title,body,channel,enabled,created_by,modified_at FROM notification_templates WHERE id=$1`, id).
		Scan(&t.ID, &t.Name, &t.Title, &t.Body, &t.Channel, &t.Enabled, &t.CreatedBy, &t.ModifiedAt)
	if err != nil {
		return models.NotificationTemplate{}, false
	}
	return t, true
}

func (s *PostgresStore) SaveRating(r models.Rating) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO ratings (id,booking_id,from_user_id,to_user_id,score,comment,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (id) DO UPDATE SET score=EXCLUDED.score, comment=EXCLUDED.comment`,
		r.ID, r.BookingID, r.FromUserID, r.ToUserID, r.Score, r.Comment, r.CreatedAt,
	)
}

func (s *PostgresStore) ListRatings(bookingID string) []models.Rating {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,booking_id,from_user_id,to_user_id,score,comment,created_at FROM ratings WHERE booking_id=$1 ORDER BY created_at ASC`, bookingID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Rating, 0)
	for rows.Next() {
		var r models.Rating
		if rows.Scan(&r.ID, &r.BookingID, &r.FromUserID, &r.ToUserID, &r.Score, &r.Comment, &r.CreatedAt) == nil {
			out = append(out, r)
		}
	}
	return out
}

func (s *PostgresStore) SaveBackupJob(job models.BackupJob) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO backup_jobs (id,job_type,status,artifact,requested_by,created_at,finished_at,error_message) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status,artifact=EXCLUDED.artifact,finished_at=EXCLUDED.finished_at,error_message=EXCLUDED.error_message`,
		job.ID, job.Type, job.Status, job.Artifact, nullableUUID(job.RequestedBy), job.CreatedAt, nullableTime(job.FinishedAt), nullIfEmpty(job.Error),
	)
}

func (s *PostgresStore) ListBackupJobs() []models.BackupJob {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id,job_type,status,COALESCE(artifact,''),COALESCE(requested_by::text,''),created_at,COALESCE(finished_at, '0001-01-01T00:00:00Z'::timestamptz),COALESCE(error_message,'') FROM backup_jobs ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.BackupJob, 0)
	for rows.Next() {
		var b models.BackupJob
		if rows.Scan(&b.ID, &b.Type, &b.Status, &b.Artifact, &b.RequestedBy, &b.CreatedAt, &b.FinishedAt, &b.Error) == nil {
			out = append(out, b)
		}
	}
	return out
}

func (s *PostgresStore) SavePasswordResetEvidence(e models.PasswordResetEvidence) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO password_reset_evidence (id,target_user_id,checked_by,method,evidence_ref,reason,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.ID, e.TargetUserID, e.CheckedBy, e.Method, e.EvidenceRef, e.Reason, e.CreatedAt,
	)
}

func (s *PostgresStore) MarkCouponUsed(code, bookingID string) bool {
	ctx := context.Background()
	fingerprint := code + ":" + bookingID
	res, err := s.pool.Exec(ctx, `INSERT INTO coupon_redemptions (id,code,booking_id,status,fingerprint,created_at) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (fingerprint) DO NOTHING`, uuid.NewString(), code, bookingID, "provisional", fingerprint, time.Now().UTC())
	if err != nil {
		return false
	}
	return res.RowsAffected() == 1
}

func nullableTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

func nullableUUID(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func nullIfEmpty(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
