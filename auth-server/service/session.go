package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/google/uuid"
)

// SessionService is an interface for session/Redis related operations.
type SessionService interface {
	Create(ctx context.Context, userID string) (*http.Cookie, error)
	Extend(ctx context.Context, userID string, cookie *http.Cookie) (*http.Cookie, error)
	Fetch(ctx context.Context, sessionID string) (uuid.UUID, error)
	Remove(ctx context.Context, cookie *http.Cookie) (*http.Cookie, error)
}

type sessionService struct {
	sessionCache  cache.SessionCache
	sessionConfig config.Session
}

// NewSessionService creates a new SessionService with the given session cache.
func NewSessionService(sessionCache cache.SessionCache) SessionService {
	return &sessionService{
		sessionCache,
		config.New().Session,
	}
}

// Remove removes the session from shared cache and returns an expired cookie.
func (s *sessionService) Remove(ctx context.Context, cookie *http.Cookie) (*http.Cookie, error) {
	s.modifyCookie(cookie)
	cookie.MaxAge = 0
	cookie.Expires = time.Now() // workaround since MaxAge 0 not being respected by some tools/browsers
	return cookie, s.sessionCache.Del(ctx, cookie.Value)
}

func (s *sessionService) Fetch(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userID, err := s.sessionCache.Get(ctx, sessionID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(userID)
}

func (s *sessionService) Create(ctx context.Context, userID string) (*http.Cookie, error) {
	sessionID := generateSessionID()
	expiration := maxAgeToExpiration(s.sessionConfig.MaxAge)
	err := s.sessionCache.Set(ctx, sessionID, userID, expiration)
	return s.newCookie(sessionID), err
}

// Extend updates the expiration of the session in the session cache and
func (s *sessionService) Extend(ctx context.Context, userID string, cookie *http.Cookie) (*http.Cookie, error) {
	s.modifyCookie(cookie)
	return cookie, s.sessionCache.Set(ctx, cookie.Value, userID, maxAgeToExpiration(s.sessionConfig.MaxAge))
}

// base64 encoded 32 byte random string
func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// TODO Validate contents of cookie to ensure it has not been modified/tampered with.
// This can be done by adding a message authentication code (MAC) to the cookie,
// which can be used to verify the integrity of the cookie's contents.
func (s *sessionService) newCookie(value string) *http.Cookie {
	session := s.sessionConfig
	return &http.Cookie{
		Value:    value,
		Name:     session.Name,
		Domain:   session.Domain,
		Path:     session.Path,
		MaxAge:   session.MaxAge,
		Secure:   session.Secure,
		HttpOnly: session.HTTPOnly,
		SameSite: mapSameSite(session.SameSite),
	}
}

func (s *sessionService) modifyCookie(cookie *http.Cookie) {
	session := s.sessionConfig
	cookie.Name = session.Name
	cookie.Domain = session.Domain
	cookie.Path = session.Path
	cookie.MaxAge = session.MaxAge
	cookie.Secure = session.Secure
	cookie.HttpOnly = session.HTTPOnly
	cookie.SameSite = mapSameSite(session.SameSite)
}

// Convert Cookie MaxAge from seconds to time.Duration
func maxAgeToExpiration(maxAge int) time.Duration {
	return time.Duration(maxAge) * time.Second
}

func mapSameSite(value string) http.SameSite {
	switch value {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
