package middleware

import (
	"net/http"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/store"

	"github.com/labstack/echo/v4"
)

func RequireMFAForRoles(st store.Repository, protectedRoles ...models.Role) echo.MiddlewareFunc {
	protectedSet := map[models.Role]struct{}{}
	for _, role := range protectedRoles {
		protectedSet[role] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := c.Get(CtxRoles).([]models.Role)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "roles missing"})
			}
			enforce := false
			for _, role := range roles {
				if _, exists := protectedSet[role]; exists {
					enforce = true
					break
				}
			}
			if !enforce {
				return next(c)
			}

			userID, _ := c.Get(CtxUserID).(string)
			user, ok := st.GetUserByID(userID)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
			}
			if !user.TOTPEnabled {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "MFA required for admin-sensitive actions"})
			}
			return next(c)
		}
	}
}
