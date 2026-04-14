package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
)

// ---------------------------------------------------------------------------
// Complaints
// ---------------------------------------------------------------------------

func TestSaveAndGetComplaint(t *testing.T) {
	st := newStore()
	c := models.Complaint{
		ID:        "cmp1",
		BookingID: "b1",
		OpenedBy:  "u1",
		Status:    "open",
		CreatedAt: time.Now().UTC(),
	}
	st.SaveComplaint(c)
	got, ok := st.GetComplaint("cmp1")
	if !ok {
		t.Fatal("expected complaint to be found")
	}
	if got.Status != "open" {
		t.Fatalf("expected status open, got %s", got.Status)
	}
}

func TestGetComplaint_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetComplaint("no-cmp")
	if ok {
		t.Fatal("expected false for missing complaint")
	}
}

func TestListComplaints(t *testing.T) {
	st := newStore()
	st.SaveComplaint(models.Complaint{ID: "c1", BookingID: "b1", Status: "open"})
	st.SaveComplaint(models.Complaint{ID: "c2", BookingID: "b2", Status: "resolved"})
	complaints := st.ListComplaints()
	if len(complaints) != 2 {
		t.Fatalf("expected 2 complaints, got %d", len(complaints))
	}
}

func TestComplaintStatusTransition(t *testing.T) {
	st := newStore()
	st.SaveComplaint(models.Complaint{ID: "c1", BookingID: "b1", Status: "open"})
	c, _ := st.GetComplaint("c1")
	c.Status = "resolved"
	c.Outcome = "refunded"
	st.SaveComplaint(c)
	got, _ := st.GetComplaint("c1")
	if got.Status != "resolved" {
		t.Fatalf("expected resolved, got %s", got.Status)
	}
	if got.Outcome != "refunded" {
		t.Fatalf("expected refunded outcome, got %s", got.Outcome)
	}
}

// ---------------------------------------------------------------------------
// Consultations: save, get, list by booking/thread/topic
// ---------------------------------------------------------------------------

func TestSaveAndGetConsultation(t *testing.T) {
	st := newStore()
	c := models.Consultation{
		ID:        "con1",
		BookingID: "b1",
		Topic:     "billing",
		KeyPoints: "overcharge on mileage",
		Version:   1,
		CreatedAt: time.Now().UTC(),
	}
	st.SaveConsultation(c)
	got, ok := st.GetConsultation("con1")
	if !ok {
		t.Fatal("expected consultation to be found")
	}
	if got.Topic != "billing" {
		t.Fatalf("expected topic billing, got %s", got.Topic)
	}
}

func TestGetConsultation_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetConsultation("no-con")
	if ok {
		t.Fatal("expected false for missing consultation")
	}
}

func TestListConsultationsByBooking(t *testing.T) {
	st := newStore()
	st.SaveConsultation(models.Consultation{ID: "con1", BookingID: "b1", Topic: "billing", Version: 1})
	st.SaveConsultation(models.Consultation{ID: "con2", BookingID: "b1", Topic: "damage", Version: 1})
	st.SaveConsultation(models.Consultation{ID: "con3", BookingID: "b2", Topic: "billing", Version: 1})
	b1cons := st.ListConsultationsByBooking("b1")
	if len(b1cons) != 2 {
		t.Fatalf("expected 2 consultations for b1, got %d", len(b1cons))
	}
}

func TestListConsultationsByThread(t *testing.T) {
	// Thread ID = bookingID + "::" + topic
	st := newStore()
	st.SaveConsultation(models.Consultation{ID: "con1", BookingID: "b1", Topic: "billing", Version: 1})
	st.SaveConsultation(models.Consultation{ID: "con2", BookingID: "b1", Topic: "billing", Version: 2}) // same thread
	st.SaveConsultation(models.Consultation{ID: "con3", BookingID: "b1", Topic: "damage", Version: 1}) // different thread
	thread := st.ListConsultationsByThread("b1::billing")
	if len(thread) != 2 {
		t.Fatalf("expected 2 in billing thread, got %d", len(thread))
	}
}

func TestListConsultationsByTopic(t *testing.T) {
	st := newStore()
	st.SaveConsultation(models.Consultation{ID: "con1", BookingID: "b1", Topic: "billing", Version: 1})
	st.SaveConsultation(models.Consultation{ID: "con2", BookingID: "b2", Topic: "billing", Version: 1})
	st.SaveConsultation(models.Consultation{ID: "con3", BookingID: "b1", Topic: "damage", Version: 1})
	billingCons := st.ListConsultationsByTopic("billing")
	if len(billingCons) != 2 {
		t.Fatalf("expected 2 billing consultations, got %d", len(billingCons))
	}
}

// ---------------------------------------------------------------------------
// Consultation attachments
// ---------------------------------------------------------------------------

func TestSaveAndListConsultationAttachments(t *testing.T) {
	st := newStore()
	a := models.ConsultationAttachment{
		ID:             "ca1",
		ConsultationID: "con1",
		AttachmentID:   "att1",
		CreatedAt:      time.Now().UTC(),
	}
	st.SaveConsultationAttachment(a)
	attachments := st.ListConsultationAttachments("con1")
	if len(attachments) != 1 {
		t.Fatalf("expected 1 consultation attachment, got %d", len(attachments))
	}
	if attachments[0].AttachmentID != "att1" {
		t.Fatalf("expected att1, got %s", attachments[0].AttachmentID)
	}
}

func TestListConsultationAttachments_Empty(t *testing.T) {
	st := newStore()
	attachments := st.ListConsultationAttachments("no-con")
	if len(attachments) != 0 {
		t.Fatalf("expected 0 attachments, got %d", len(attachments))
	}
}

// ---------------------------------------------------------------------------
// Ratings
// ---------------------------------------------------------------------------

func TestSaveAndListRatings(t *testing.T) {
	st := newStore()
	r := models.Rating{
		ID:         "r1",
		BookingID:  "b1",
		FromUserID: "u1",
		ToUserID:   "u2",
		Score:      5,
		Comment:    "excellent service",
		CreatedAt:  time.Now().UTC(),
	}
	st.SaveRating(r)
	ratings := st.ListRatings("b1")
	if len(ratings) != 1 {
		t.Fatalf("expected 1 rating, got %d", len(ratings))
	}
	if ratings[0].Score != 5 {
		t.Fatalf("expected score 5, got %d", ratings[0].Score)
	}
}

func TestListRatings_Empty(t *testing.T) {
	st := newStore()
	ratings := st.ListRatings("no-booking")
	if len(ratings) != 0 {
		t.Fatalf("expected 0 ratings, got %d", len(ratings))
	}
}

func TestListRatings_MultipleRatingsPerBooking(t *testing.T) {
	// Both customer->provider and provider->customer ratings for same booking
	st := newStore()
	st.SaveRating(models.Rating{ID: "r1", BookingID: "b1", FromUserID: "cust", ToUserID: "prov", Score: 4})
	st.SaveRating(models.Rating{ID: "r2", BookingID: "b1", FromUserID: "prov", ToUserID: "cust", Score: 5})
	ratings := st.ListRatings("b1")
	if len(ratings) != 2 {
		t.Fatalf("expected 2 ratings for b1, got %d", len(ratings))
	}
}

func TestListRatings_IsolatedByBooking(t *testing.T) {
	st := newStore()
	st.SaveRating(models.Rating{ID: "r1", BookingID: "b1", Score: 4})
	st.SaveRating(models.Rating{ID: "r2", BookingID: "b2", Score: 3})
	if len(st.ListRatings("b1")) != 1 {
		t.Fatal("expected 1 rating for b1")
	}
	if len(st.ListRatings("b2")) != 1 {
		t.Fatal("expected 1 rating for b2")
	}
}
