package middleware

import (
	"net/http"

	"fleetlease/backend/internal/models"

	"github.com/labstack/echo/v4"
)

func RequireRoles(allowed ...models.Role) echo.MiddlewareFunc {
	allowedSet := map[models.Role]struct{}{}
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := c.Get(CtxRoles).([]models.Role)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "roles missing"})
			}
			for _, r := range roles {
				if _, exists := allowedSet[r]; exists {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient role"})
		}
	}
}
