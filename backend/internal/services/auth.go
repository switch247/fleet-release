package services

import (
	"errors"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/store"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	secret         []byte
	idleTimeout    time.Duration
	absoluteTimout time.Duration
	store          store.Repository
}

type Claims struct {
	UserID string        `json:"uid"`
	Roles  []models.Role `json:"roles"`
	SID    string        `json:"sid"`
	jwt.RegisteredClaims
}

func NewAuthService(secret string, idleTimeout, absoluteTimeout time.Duration, st store.Repository) *AuthService {
	return &AuthService{
		secret:         []byte(secret),
		idleTimeout:    idleTimeout,
		absoluteTimout: absoluteTimeout,
		store:          st,
	}
}

func (a *AuthService) IssueToken(user models.User) (string, models.Session, error) {
	now := time.Now().UTC()
	sid := uuid.NewString()
	session := models.Session{
		ID:          sid,
		UserID:      user.ID,
		IssuedAt:    now,
		LastSeenAt:  now,
		AbsoluteExp: now.Add(a.absoluteTimout),
	}
	a.store.SaveSession(session)

	claims := Claims{
		UserID: user.ID,
		Roles:  user.Roles,
		SID:    sid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.absoluteTimout)),
			ID:        sid,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", models.Session{}, err
	}
	return signed, session, nil
}

func (a *AuthService) ParseAndValidate(tokenString string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	if err != nil || !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return Claims{}, errors.New("invalid claims")
	}
	session, ok := a.store.GetSession(claims.SID)
	if !ok || session.Revoked {
		return Claims{}, errors.New("session revoked")
	}
	now := time.Now().UTC()
	if now.After(session.AbsoluteExp) {
		return Claims{}, errors.New("absolute timeout reached")
	}
	if now.Sub(session.LastSeenAt) > a.idleTimeout {
		return Claims{}, errors.New("idle timeout reached")
	}
	session.LastSeenAt = now
	a.store.SaveSession(session)
	return *claims, nil
}

func (a *AuthService) RevokeSession(sessionID string) {
	session, ok := a.store.GetSession(sessionID)
	if !ok {
		return
	}
	session.Revoked = true
	a.store.SaveSession(session)
}
