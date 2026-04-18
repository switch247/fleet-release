package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/services"
	"fleetlease/backend/internal/store"

	"github.com/labstack/echo/v4"
)

type fakeAuth struct {
	parseErr error
	claims   *services.Claims
}

func (f *fakeAuth) ParseAndValidate(token string) (claims *services.Claims, err error) {
	if f.parseErr != nil {
		return nil, f.parseErr
	}
	return f.claims, nil
}

func TestJWTAuth_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	st := store.NewMemoryStore()
	auth := services.NewAuthService("secret", time.Minute, time.Hour, st)
	user := models.User{ID: "u1", Roles: []models.Role{"admin"}}
	token, _, _ := auth.IssueToken(user)
	req.Header.Set("Authorization", "Bearer "+token)
	mw := JWTAuth(auth)
	h := mw(func(c echo.Context) error {
		if c.Get(CtxUserID) != "u1" || c.Get(CtxSID) == nil {
			t.Fatalf("claims not set")
		}
		return c.String(200, "ok")
	})
	_ = h(c)
}

func TestJWTAuth_MissingBearer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	st := store.NewMemoryStore()
	auth := services.NewAuthService("secret", time.Minute, time.Hour, st)
	mw := JWTAuth(auth)
	h := mw(func(c echo.Context) error { return nil })
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized")
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	st := store.NewMemoryStore()
	auth := services.NewAuthService("secret", time.Minute, time.Hour, st)
	mw := JWTAuth(auth)
	h := mw(func(c echo.Context) error { return nil })
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized")
	}
}

func TestRequireRoles(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxRoles, []models.Role{"admin"})
	mw := RequireRoles("admin")
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != 200 {
		t.Fatalf("expected ok")
	}
}

func TestRequireRoles_Forbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(CtxRoles, []models.Role{"user"})
	mw := RequireRoles("admin")
	h := mw(func(c echo.Context) error { return c.String(200, "ok") })
	_ = h(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden")
	}
}
