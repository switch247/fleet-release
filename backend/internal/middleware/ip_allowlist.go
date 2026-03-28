package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func AdminIPAllowlist(cidrs, trustedProxyCIDRs []string) echo.MiddlewareFunc {
	nets := parseCIDRs(cidrs)
	trusted := parseCIDRs(trustedProxyCIDRs)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := requestRemoteIP(c, trusted)
			if ip == nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "unable to determine client IP"})
			}
			for _, n := range nets {
				if n.Contains(ip) {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, map[string]string{"error": "ip not allowlisted"})
		}
	}
}

func parseCIDRs(cidrs []string) []*net.IPNet {
	nets := []*net.IPNet{}
	for _, c := range cidrs {
		_, ipnet, err := net.ParseCIDR(c)
		if err == nil {
			nets = append(nets, ipnet)
		}
	}
	return nets
}

func requestRemoteIP(c echo.Context, trustedProxies []*net.IPNet) net.IP {
	host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		host = c.Request().RemoteAddr
	}
	remote := net.ParseIP(host)
	if remote == nil {
		return nil
	}
	if !isFromTrustedProxy(remote, trustedProxies) {
		return remote
	}
	forwarded := c.Request().Header.Get("X-Forwarded-For")
	if forwarded == "" {
		return remote
	}
	firstHop := strings.TrimSpace(strings.Split(forwarded, ",")[0])
	ip := net.ParseIP(firstHop)
	if ip == nil {
		return remote
	}
	return ip
}

func isFromTrustedProxy(ip net.IP, trusted []*net.IPNet) bool {
	for _, n := range trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
