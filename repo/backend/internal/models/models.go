package models

import "time"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleProvider Role = "provider"
	RoleCSA      Role = "csa"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	Roles           []Role    `json:"roles"`
	FailedAttempts  int       `json:"-"`
	LockedUntil     time.Time `json:"-"`
	TOTPSecret      string    `json:"-"`
	TOTPEnabled     bool      `json:"totpEnabled"`
	GovernmentIDEnc    string `json:"-"`
	PaymentReferenceEnc string `json:"-"`
	AddressEnc          string `json:"-"`
}

type Session struct {
	ID          string
	UserID      string
	IssuedAt    time.Time
	LastSeenAt  time.Time
	AbsoluteExp time.Time
	Revoked     bool
}

type AuthEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	EventType string    `json:"eventType"`
	CreatedAt time.Time `json:"createdAt"`
}

type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parentId"`
}

type Listing struct {
	ID            string  `json:"id"`
	CategoryID    string  `json:"categoryId"`
	ProviderID    string  `json:"providerId"`
	SPU           string  `json:"spu"`
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	IncludedMiles float64 `json:"includedMiles"`
	Deposit       float64 `json:"deposit"`
	Available     bool    `json:"available"`
}

type Booking struct {
	ID              string    `json:"id"`
	CustomerID      string    `json:"customerId"`
	ProviderID      string    `json:"providerId"`
	ListingID       string    `json:"listingId"`
	CouponCode      string    `json:"couponCode"`
	StartAt         time.Time `json:"startAt"`
	EndAt           time.Time `json:"endAt"`
	OdoStart        float64   `json:"odoStart"`
	OdoEnd          float64   `json:"odoEnd"`
	Status          string    `json:"status"`
	EstimatedAmount float64   `json:"estimatedAmount"`
	DepositAmount   float64   `json:"depositAmount"`
}

type StatsSummary struct {
	ActiveBookings int     `json:"activeBookings"`
	SettledTrips   int     `json:"settledTrips"`
	InspectionsDue int     `json:"inspectionsDue"`
	HeldDeposits   float64 `json:"heldDeposits"`
}

type InspectionItem struct {
	Name                  string   `json:"name"`
	Condition             string   `json:"condition"`
	EvidenceIDs           []string `json:"evidenceIds"`
	DamageDeductionAmount float64  `json:"damageDeductionAmount,omitempty"`
}

type InspectionRevision struct {
	RevisionID string           `json:"revisionId"`
	BookingID  string           `json:"bookingId"`
	Stage      string           `json:"stage"`
	Items      []InspectionItem `json:"items"`
	Notes      string           `json:"notes"`
	CreatedBy  string           `json:"createdBy"`
	CreatedAt  time.Time        `json:"createdAt"`
	PrevHash   string           `json:"prevHash"`
	Hash       string           `json:"hash"`
}

type Attachment struct {
	ID          string `json:"id"`
	BookingID   string `json:"bookingId"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	SizeBytes   int64  `json:"sizeBytes"`
	Checksum    string `json:"checksum"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   time.Time `json:"createdAt"`
}

type LedgerEntry struct {
	ID          string    `json:"id"`
	BookingID   string    `json:"bookingId"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	PrevHash    string    `json:"prevHash"`
	Hash        string    `json:"hash"`
}

type Complaint struct {
	ID        string    `json:"id"`
	BookingID string    `json:"bookingId"`
	OpenedBy  string    `json:"openedBy"`
	Status    string    `json:"status"`
	Outcome   string    `json:"outcome"`
	CreatedAt time.Time `json:"createdAt"`
}

type Consultation struct {
	ID             string    `json:"id"`
	BookingID      string    `json:"bookingId"`
	Topic          string    `json:"topic"`
	KeyPoints      string    `json:"keyPoints"`
	Recommendation string    `json:"recommendation"`
	FollowUp       string    `json:"followUp"`
	Visibility     string    `json:"visibility"`
	ChangeReason   string    `json:"changeReason"`
	Version        int       `json:"version"`
	CreatedBy      string    `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ConsultationAttachment struct {
	ID             string    `json:"id"`
	ConsultationID string    `json:"consultationId"`
	AttachmentID   string    `json:"attachmentId"`
	CreatedBy      string    `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Notification struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	TemplateID  string    `json:"templateId"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Status      string    `json:"status"`
	Attempts    int       `json:"attempts"`
	Fingerprint string    `json:"fingerprint"`
	DeliveredAt time.Time `json:"deliveredAt"`
}

type NotificationTemplate struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Channel    string    `json:"channel"`
	Enabled    bool      `json:"enabled"`
	CreatedBy  string    `json:"createdBy"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type Rating struct {
	ID         string    `json:"id"`
	BookingID  string    `json:"bookingId"`
	FromUserID string    `json:"fromUserId"`
	ToUserID   string    `json:"toUserId"`
	Score      int       `json:"score"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"createdAt"`
}

type BackupJob struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Artifact    string    `json:"artifact"`
	RequestedBy string    `json:"requestedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	FinishedAt  time.Time `json:"finishedAt"`
	Error       string    `json:"error"`
}

type PasswordResetEvidence struct {
	ID           string    `json:"id"`
	TargetUserID string    `json:"targetUserId"`
	CheckedBy    string    `json:"checkedBy"`
	Method       string    `json:"method"`
	EvidenceRef  string    `json:"evidenceRef"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

type RetentionReport struct {
	ID                 string    `json:"id"`
	AttachmentsDeleted int       `json:"attachmentsDeleted"`
	LedgerDeleted      int       `json:"ledgerDeleted"`
	FileDeleteErrors   int       `json:"fileDeleteErrors"`
	CreatedAt          time.Time `json:"createdAt"`
}
