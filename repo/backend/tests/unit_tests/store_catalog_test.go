package unit_tests

import (
	"testing"

	"fleetlease/backend/internal/models"
)

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

func TestSaveAndGetCategory(t *testing.T) {
	st := newStore()
	cat := models.Category{ID: "cat1", Name: "Trucks"}
	st.SaveCategory(cat)
	got, ok := st.GetCategory("cat1")
	if !ok {
		t.Fatal("expected category to be found")
	}
	if got.Name != "Trucks" {
		t.Fatalf("expected Trucks, got %s", got.Name)
	}
}

func TestGetCategory_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetCategory("no-cat")
	if ok {
		t.Fatal("expected false for missing category")
	}
}

func TestListCategories(t *testing.T) {
	st := newStore()
	st.SaveCategory(models.Category{ID: "c1", Name: "Cars"})
	st.SaveCategory(models.Category{ID: "c2", Name: "Vans"})
	cats := st.ListCategories()
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
}

func TestDeleteCategory(t *testing.T) {
	st := newStore()
	st.SaveCategory(models.Category{ID: "c1", Name: "Cars"})
	st.DeleteCategory("c1")
	_, ok := st.GetCategory("c1")
	if ok {
		t.Fatal("expected category deleted")
	}
}

func TestDeleteCategory_Idempotent(t *testing.T) {
	st := newStore()
	st.DeleteCategory("nonexistent")
}

func TestCategoryHierarchy(t *testing.T) {
	st := newStore()
	parent := models.Category{ID: "p1", Name: "Vehicles"}
	child := models.Category{ID: "c1", Name: "SUVs", ParentID: "p1"}
	st.SaveCategory(parent)
	st.SaveCategory(child)
	got, ok := st.GetCategory("c1")
	if !ok {
		t.Fatal("expected child category")
	}
	if got.ParentID != "p1" {
		t.Fatalf("expected parentID p1, got %s", got.ParentID)
	}
}

// ---------------------------------------------------------------------------
// Listings
// ---------------------------------------------------------------------------

func TestSaveAndGetListing(t *testing.T) {
	st := newStore()
	lst := models.Listing{ID: "l1", CategoryID: "c1", ProviderID: "p1", SPU: "SUV-SPU", SKU: "SUV-A", Name: "City SUV", IncludedMiles: 3, Deposit: 100, Available: true}
	st.SaveListing(lst)
	got, ok := st.GetListing("l1")
	if !ok {
		t.Fatal("expected listing to be found")
	}
	if got.Name != "City SUV" {
		t.Fatalf("expected City SUV, got %s", got.Name)
	}
}

func TestGetListing_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetListing("no-listing")
	if ok {
		t.Fatal("expected false for missing listing")
	}
}

func TestListListings(t *testing.T) {
	st := newStore()
	st.SaveListing(models.Listing{ID: "l1", Name: "Sedan A"})
	st.SaveListing(models.Listing{ID: "l2", Name: "Sedan B"})
	listings := st.ListListings()
	if len(listings) != 2 {
		t.Fatalf("expected 2 listings, got %d", len(listings))
	}
}

func TestDeleteListing(t *testing.T) {
	st := newStore()
	st.SaveListing(models.Listing{ID: "l1", Name: "Sedan A"})
	st.DeleteListing("l1")
	_, ok := st.GetListing("l1")
	if ok {
		t.Fatal("expected listing deleted")
	}
}

func TestDeleteListing_Idempotent(t *testing.T) {
	st := newStore()
	st.DeleteListing("nonexistent")
}

func TestListingAvailabilityToggle(t *testing.T) {
	st := newStore()
	st.SaveListing(models.Listing{ID: "l1", Name: "Van", Available: true})
	got, _ := st.GetListing("l1")
	if !got.Available {
		t.Fatal("expected available=true")
	}
	got.Available = false
	st.SaveListing(got)
	got2, _ := st.GetListing("l1")
	if got2.Available {
		t.Fatal("expected available=false after update")
	}
}

// ---------------------------------------------------------------------------
// Bookings
// ---------------------------------------------------------------------------

func TestSaveAndGetBooking(t *testing.T) {
	st := newStore()
	b := models.Booking{ID: "b1", CustomerID: "c1", ProviderID: "p1", ListingID: "l1", Status: "booked"}
	st.SaveBooking(b)
	got, ok := st.GetBooking("b1")
	if !ok {
		t.Fatal("expected booking to be found")
	}
	if got.Status != "booked" {
		t.Fatalf("expected status booked, got %s", got.Status)
	}
}

func TestGetBooking_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetBooking("no-booking")
	if ok {
		t.Fatal("expected false for missing booking")
	}
}

func TestListBookings(t *testing.T) {
	st := newStore()
	st.SaveBooking(models.Booking{ID: "b1", Status: "booked"})
	st.SaveBooking(models.Booking{ID: "b2", Status: "settled"})
	bookings := st.ListBookings()
	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}
}

func TestBookingStatusTransition(t *testing.T) {
	st := newStore()
	st.SaveBooking(models.Booking{ID: "b1", Status: "booked"})
	b, _ := st.GetBooking("b1")
	b.Status = "active"
	st.SaveBooking(b)
	got, _ := st.GetBooking("b1")
	if got.Status != "active" {
		t.Fatalf("expected status active after transition, got %s", got.Status)
	}
}
