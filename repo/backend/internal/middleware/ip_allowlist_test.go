package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRequestRemoteIPIgnoresForwardedWhenProxyUntrusted(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:8080"
	req.Header.Set("X-Forwarded-For", "198.51.100.44")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ip := requestRemoteIP(c, parseCIDRs([]string{"192.0.2.0/24"}))
	if ip == nil || ip.String() != "203.0.113.10" {
		t.Fatalf("expected remote addr IP, got %v", ip)
	}
}

func TestRequestRemoteIPUsesForwardedForTrustedProxy(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.5:8080"
	req.Header.Set("X-Forwarded-For", "198.51.100.44, 192.0.2.5")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ip := requestRemoteIP(c, parseCIDRs([]string{"192.0.2.0/24"}))
	if ip == nil || ip.String() != "198.51.100.44" {
		t.Fatalf("expected forwarded client IP, got %v", ip)
	}
}
