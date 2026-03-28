package services

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/store"
)

func TestAuthIssueValidateAndRevoke(t *testing.T) {
	st := store.NewMemoryStore()
	auth := NewAuthService("secret", time.Minute, time.Hour, st)
	user := models.User{ID: "u1", Roles: []models.Role{models.RoleCustomer}}
	token, session, err := auth.IssueToken(user)
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}
	if session.ID == "" {
		t.Fatalf("expected session id")
	}
	claims, err := auth.ParseAndValidate(token)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("unexpected user id")
	}

	auth.RevokeSession(session.ID)
	if _, err = auth.ParseAndValidate(token); err == nil {
		t.Fatalf("expected revoked session error")
	}
}

func TestAuthIdleTimeout(t *testing.T) {
	st := store.NewMemoryStore()
	auth := NewAuthService("secret", 2*time.Millisecond, time.Hour, st)
	user := models.User{ID: "u2", Roles: []models.Role{models.RoleCustomer}}
	token, _, err := auth.IssueToken(user)
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err = auth.ParseAndValidate(token); err == nil {
		t.Fatalf("expected idle timeout error")
	}
}

func TestAuthAbsoluteTimeout(t *testing.T) {
	st := store.NewMemoryStore()
	auth := NewAuthService("secret", time.Minute, 2*time.Millisecond, st)
	user := models.User{ID: "u3", Roles: []models.Role{models.RoleCustomer}}
	token, _, err := auth.IssueToken(user)
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err = auth.ParseAndValidate(token); err == nil {
		t.Fatalf("expected absolute timeout error")
	}
}

func TestAuthInvalidTokenAndUnknownRevoke(t *testing.T) {
	st := store.NewMemoryStore()
	auth := NewAuthService("secret", time.Minute, time.Hour, st)
	if _, err := auth.ParseAndValidate("not-a-jwt"); err == nil {
		t.Fatalf("expected invalid token error")
	}
	auth.RevokeSession("missing-session-id")
}
