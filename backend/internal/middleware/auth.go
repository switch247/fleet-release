package middleware

import (
	"net/http"
	"strings"

	"fleetlease/backend/internal/services"

	"github.com/labstack/echo/v4"
)

const (
	CtxUserID = "userID"
	CtxRoles  = "roles"
	CtxSID    = "sid"
)

func JWTAuth(auth *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authz := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			}
			token := strings.TrimPrefix(authz, "Bearer ")
			claims, err := auth.ParseAndValidate(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}
			c.Set(CtxUserID, claims.UserID)
			c.Set(CtxRoles, claims.Roles)
			c.Set(CtxSID, claims.SID)
			return next(c)
		}
	}
}
