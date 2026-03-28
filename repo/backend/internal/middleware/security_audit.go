package middleware

import (
	"net/http"

	applogger "fleetlease/backend/internal/logger"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func SecurityAuditLog(logger *zap.Logger) echo.MiddlewareFunc {
	return SecurityAuditLogWithTrustedProxies(logger, nil)
}

func SecurityAuditLogWithTrustedProxies(logger *zap.Logger, trustedProxyCIDRs []string) echo.MiddlewareFunc {
	trusted := parseCIDRs(trustedProxyCIDRs)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			status := c.Response().Status
			if status == http.StatusUnauthorized || status == http.StatusForbidden {
				ip := ""
				parsed := requestRemoteIP(c, trusted)
				if parsed != nil {
					ip = parsed.String()
				}
				userID, _ := c.Get(CtxUserID).(string)
				resourceID := c.Param("id")
				if resourceID == "" {
					resourceID = c.Param("bookingID")
				}
				if resourceID == "" {
					resourceID = c.Param("userID")
				}
				logger.Warn("security_access_denied",
					zap.Int("status", status),
					zap.String("userId", applogger.Redact(userID)),
					zap.String("ip", ip),
					zap.String("method", c.Request().Method),
					zap.String("resource", c.Path()),
					zap.String("resourceId", applogger.Redact(resourceID)),
					zap.Object("authorization", applogger.RedactedField{Key: "header", Value: c.Request().Header.Get("Authorization")}),
				)
			}
			return err
		}
	}
}
