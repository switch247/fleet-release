package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"crypto/tls"

	"github.com/labstack/echo/v4"
)

func TestRequireTLSOrAllowlistedCIDR_TLS(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := RequireTLSOrAllowlistedCIDR([]string{"127.0.0.1/32"})
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != 200 {
		t.Fatalf("expected ok")
	}
}

func TestRequireTLSOrAllowlistedCIDR_Forbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := RequireTLSOrAllowlistedCIDR([]string{"10.0.0.0/8"})
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}
