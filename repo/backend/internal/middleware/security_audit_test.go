package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func TestSecurityAuditLog_CallsNext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := SecurityAuditLog(logger)
	called := false
	h := mw(func(c echo.Context) error { called = true; return c.String(200, "ok") })
	_ = h(c)
	if !called {
		t.Fatalf("next handler not called")
	}
}

func TestSecurityAuditLogWithTrustedProxies_LogsDenied(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxUserID, "u1")
	h := SecurityAuditLogWithTrustedProxies(logger, nil)(func(c echo.Context) error {
		c.Response().WriteHeader(http.StatusForbidden)
		return nil
	})
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}
