package handlers

import (
	"net"

	"github.com/labstack/echo/v4"
)

func requesterIP(c echo.Context) string {
	host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		host = c.Request().RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return ""
	}
	return ip.String()
}
