package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RequireTLSOrAllowlistedCIDR(cidrs []string) echo.MiddlewareFunc {
	return RequireTLSOrAllowlistedCIDRWithTrustedProxies(cidrs, nil)
}

func RequireTLSOrAllowlistedCIDRWithTrustedProxies(cidrs, trustedProxyCIDRs []string) echo.MiddlewareFunc {
	allowed := parseCIDRs(cidrs)
	trusted := parseCIDRs(trustedProxyCIDRs)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			if r.TLS != nil {
				return next(c)
			}
			ip := requestRemoteIP(c, trusted)
			if ip == nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "unable to determine client IP"})
			}
			for _, n := range allowed {
				if n.Contains(ip) {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, map[string]string{"error": "HTTPS required from non-whitelisted IP"})
		}
	}
}
