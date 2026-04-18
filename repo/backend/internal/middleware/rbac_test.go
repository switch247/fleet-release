package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/internal/models"

	"github.com/labstack/echo/v4"
)

func TestRequireRoles_MissingRoles(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := RequireRoles(models.RoleAdmin)
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}

func TestRequireRoles_Allowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxRoles, []models.Role{models.RoleAdmin})
	mw := RequireRoles(models.RoleAdmin)
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != 200 {
		t.Fatalf("expected ok")
	}
}

func TestRequireRoles_InsufficientRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxRoles, []models.Role{"user"})
	mw := RequireRoles(models.RoleAdmin)
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}
