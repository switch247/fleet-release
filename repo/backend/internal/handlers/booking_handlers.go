package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Bookings(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	all := h.Store.ListBookings()
	out := make([]models.Booking, 0, len(all))
	for _, b := range all {
		if canAccessBooking(userID, roles, b) {
			out = append(out, b)
		}
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateBooking(c echo.Context) error {
	var req struct {
		ListingID  string  `json:"listingId"`
		CouponCode string  `json:"couponCode"`
		StartAt    string  `json:"startAt"`
		EndAt      string  `json:"endAt"`
		OdoStart   float64 `json:"odoStart"`
		OdoEnd     float64 `json:"odoEnd"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	listing, ok := h.Store.GetListing(req.ListingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "listing not found"})
	}
	if !listing.Available {
		return c.JSON(http.StatusConflict, map[string]string{"error": "listing is not available for booking"})
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid startAt timestamp"})
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid endAt timestamp"})
	}
	if !endAt.After(startAt) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "endAt must be after startAt"})
	}
	pricingCfg := h.Pricing
	pricingCfg.IncludedMiles = listing.IncludedMiles
	couponPct := 0.0
	if req.CouponCode != "" {
		if pct, found := h.Store.GetCouponDiscount(req.CouponCode); found {
			couponPct = pct
		}
	}
	estimate := services.EstimateFare(pricingCfg, services.EstimateInput{StartAt: startAt, EndAt: endAt, OdoStart: req.OdoStart, OdoEnd: req.OdoEnd, Deposit: listing.Deposit, CouponDiscountPct: couponPct})
	customerID, _ := c.Get("userID").(string)
	booking := models.Booking{ID: uuid.NewString(), CustomerID: customerID, ProviderID: listing.ProviderID, ListingID: listing.ID, CouponCode: req.CouponCode, CouponDiscountAmount: estimate.CouponDiscountAmount, StartAt: startAt, EndAt: endAt, OdoStart: req.OdoStart, OdoEnd: req.OdoEnd, Status: "booked", EstimatedAmount: estimate.Total, DepositAmount: estimate.Deposit}
	h.Store.SaveBooking(booking)
	h.Logger.Info("booking_created", "bookingID", booking.ID, "customerID", booking.CustomerID)
	// Return only the booking object in the top-level 'booking' field, as expected by test helpers.
	return c.JSON(http.StatusCreated, map[string]interface{}{"booking": booking})
}

func (h *Handler) EstimateBooking(c echo.Context) error {
	var req struct {
		ListingID  string  `json:"listingId"`
		CouponCode string  `json:"couponCode"`
		StartAt    string  `json:"startAt"`
		EndAt      string  `json:"endAt"`
		OdoStart   float64 `json:"odoStart"`
		OdoEnd     float64 `json:"odoEnd"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	listing, ok := h.Store.GetListing(req.ListingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "listing not found"})
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid startAt timestamp"})
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid endAt timestamp"})
	}
	if !endAt.After(startAt) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "endAt must be after startAt"})
	}
	couponPct := 0.0
	if req.CouponCode != "" {
		if pct, found := h.Store.GetCouponDiscount(req.CouponCode); found {
			couponPct = pct
		}
	}
	pricingCfg := h.Pricing
	pricingCfg.IncludedMiles = listing.IncludedMiles
	estimate := services.EstimateFare(pricingCfg, services.EstimateInput{
		StartAt:           startAt,
		EndAt:             endAt,
		OdoStart:          req.OdoStart,
		OdoEnd:            req.OdoEnd,
		Deposit:           listing.Deposit,
		CouponDiscountPct: couponPct,
	})
	customerID, _ := c.Get("userID").(string)
	h.Logger.Info("booking_estimate", "listingId", listing.ID, "customerId", customerID, "total", estimate.Total, "deposit", estimate.Deposit)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"listingId": listing.ID,
		"estimate":  estimate,
	})
}

func (h *Handler) RedeemCoupon(c echo.Context) error {
	var req struct {
		Code      string `json:"code"`
		BookingID string `json:"bookingId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if req.Code == "" || req.BookingID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "code and bookingId are required"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	booking, ok := h.Store.GetBooking(req.BookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if ok := h.Store.MarkCouponUsed(req.Code, req.BookingID); !ok {
		return c.JSON(http.StatusConflict, map[string]string{"error": "coupon already redeemed"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "provisional accepted"})
}

func (h *Handler) CloseSettlement(c echo.Context) error {
	bookingID := c.Param("bookingID")
	booking, ok := h.Store.GetBooking(bookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if booking.Status == "settled" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "already settled"})
	}
	prev := h.Store.ListLedger(bookingID)
	prevHash := ""
	if len(prev) > 0 {
		prevHash = prev[len(prev)-1].Hash
	}
	// Truncate to microsecond precision so the timestamp stored in postgres
	// (which has µs resolution) matches the value used to compute the hash.
	now := time.Now().UTC().Truncate(time.Microsecond)
	chargeHash := services.ChainHash(prevHash, "trip_charge|"+formatAmount(booking.EstimatedAmount)+"|Trip fare settlement", now)
	charge := models.LedgerEntry{ID: uuid.NewString(), BookingID: bookingID, Type: "trip_charge", Amount: booking.EstimatedAmount, Description: "Trip fare settlement", CreatedAt: now, PrevHash: prevHash, Hash: chargeHash}
	h.Store.AppendLedger(bookingID, charge)
	refundPrev := charge.Hash

	// Apply coupon discount ledger entry if a discount was recorded at booking time.
	if booking.CouponDiscountAmount > 0 {
		couponTs := now.Add(time.Millisecond)
		couponHash := services.ChainHash(refundPrev, "coupon_discount|"+formatAmount(booking.CouponDiscountAmount)+"|Coupon code discount: "+booking.CouponCode, couponTs)
		couponEntry := models.LedgerEntry{
			ID: uuid.NewString(), BookingID: bookingID, Type: "coupon_discount",
			Amount:      booking.CouponDiscountAmount,
			Description: "Coupon code discount: " + booking.CouponCode,
			CreatedAt:   couponTs, PrevHash: refundPrev, Hash: couponHash,
		}
		h.Store.AppendLedger(bookingID, couponEntry)
		refundPrev = couponEntry.Hash
	}

	// Sum wear-and-tear deduction amounts recorded across all inspection revisions.
	var totalDeductions float64
	for _, rev := range h.Store.ListInspections(bookingID) {
		for _, item := range rev.Items {
			totalDeductions += item.DamageDeductionAmount
		}
	}
	nextTs := now.Add(2 * time.Millisecond)
	if totalDeductions > 0 {
		wearHash := services.ChainHash(refundPrev, "wear_deduction|"+formatAmount(totalDeductions)+"|Inspection wear-and-tear deduction", nextTs)
		wearEntry := models.LedgerEntry{
			ID: uuid.NewString(), BookingID: bookingID, Type: "wear_deduction",
			Amount: totalDeductions, Description: "Inspection wear-and-tear deduction",
			CreatedAt: nextTs, PrevHash: refundPrev, Hash: wearHash,
		}
		h.Store.AppendLedger(bookingID, wearEntry)
		refundPrev = wearEntry.Hash
		nextTs = nextTs.Add(time.Millisecond)
	}

	refundAmount := booking.DepositAmount - booking.EstimatedAmount - totalDeductions
	refundType := "deposit_refund"
	if refundAmount < 0 {
		refundType = "deposit_deduction"
	}
	refundAt := nextTs
	refundHash := services.ChainHash(refundPrev, refundType+"|"+formatAmount(refundAmount)+"|Auto settlement of deposit", refundAt)
	refund := models.LedgerEntry{ID: uuid.NewString(), BookingID: bookingID, Type: refundType, Amount: refundAmount, Description: "Auto settlement of deposit", CreatedAt: refundAt, PrevHash: refundPrev, Hash: refundHash}
	h.Store.AppendLedger(bookingID, refund)
	booking.Status = "settled"
	h.Store.SaveBooking(booking)
	h.Logger.Info("booking_settled", "bookingId", bookingID, "customerId", booking.CustomerID, "providerId", booking.ProviderID, "refundType", refundType, "totalCharged", booking.EstimatedAmount)

	// RULES: no 3rd-party integrations in offline mode.
	// Payment processor call intentionally stubbed for offline-first deployment.

	return c.JSON(http.StatusOK, map[string]interface{}{"booking": booking, "ledger": h.Store.ListLedger(bookingID)})
}

func (h *Handler) Ledger(c echo.Context) error {
	bookingID := c.Param("bookingID")
	booking, ok := h.Store.GetBooking(bookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	return c.JSON(http.StatusOK, h.Store.ListLedger(bookingID))
}

func (h *Handler) VerifyLedger(c echo.Context) error {
	bookingID := c.Param("bookingID")
	booking, ok := h.Store.GetBooking(bookingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "booking not found"})
	}
	actor, _ := c.Get("userID").(string)
	roles, _ := c.Get("roles").([]models.Role)
	if !canAccessBooking(actor, roles, booking) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	entries := h.Store.ListLedger(bookingID)
	prev := ""
	valid := true
	for _, e := range entries {
		expected := services.ChainHash(prev, e.Type+"|"+formatAmount(e.Amount)+"|"+e.Description, e.CreatedAt)
		if e.PrevHash != prev || !strings.EqualFold(e.Hash, expected) {
			valid = false
			break
		}
		prev = e.Hash
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"bookingId": bookingID, "valid": valid, "entries": entries})
}

func formatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
