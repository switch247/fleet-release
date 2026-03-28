package handlers

import "fleetlease/backend/internal/models"

func canAccessBooking(userID string, roles []models.Role, booking models.Booking) bool {
	if booking.CustomerID == userID {
		return true
	}
	if booking.ProviderID != "" && booking.ProviderID == userID {
		return true
	}
	if hasRole(roles, models.RoleAdmin) || hasRole(roles, models.RoleCSA) {
		return true
	}
	return false
}
