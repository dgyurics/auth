package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/google/uuid"
)

// SessionService is an interface for session/Redis related operations.
type SessionService interface {
	Create(ctx context.Context, userID uuid.UUID) (*http.Cookie, error)
	Extend(ctx context.Context, userID string, cookie *http.Cookie) (*http.Cookie, error)
	Fetch(ctx context.Context, sessionID string) (uuid.UUID, error)
	FetchAll(ctx context.Context, sessionID string) ([]string, error)
	Remove(ctx context.Context, cookie *http.Cookie) (*http.Cookie, error)
	RemoveAll(ctx context.Context, cookie *http.Cookie) (*http.Cookie, error)
}

type sessionService struct {
	sessionCache  cache.SessionCache
	sessionConfig config.Session
}

// FIXME how do we invalidate sessions in SQL when they expire in Redis aka sessionCache?
// Possible solution: trigger event on Redis expiration and invalidate session in SQL
// https://medium.com/nerd-for-tech/redis-getting-notified-when-a-key-is-expired-or-changed-ca3e1f1c7f0a

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

	userID, err := s.sessionCache.Get(ctx, cookie.Value)
	if err != nil {
		return nil, err
	}

	if err := s.sessionCache.Del(ctx, cookie.Value); err != nil {
		return nil, err
	}

	if err := s.sessionCache.SRem(ctx, userID, cookie.Value); err != nil {
		return nil, err
	}
	return cookie, nil
}

// RemoveAll removes all sessions for the user from shared cache and returns an expired cookie.
func (s *sessionService) RemoveAll(ctx context.Context, cookie *http.Cookie) (*http.Cookie, error) {
	s.modifyCookie(cookie)
	cookie.MaxAge = 0
	cookie.Expires = time.Now() // workaround since MaxAge 0 not being respected by some tools/browsers

	userID, err := s.sessionCache.Get(ctx, cookie.Value)
	if err != nil {
		return nil, err
	}

	sessions, err := s.sessionCache.SMembers(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if err := s.sessionCache.Del(ctx, session); err != nil {
			return nil, err
		}
		if err := s.sessionCache.SRem(ctx, userID, session); err != nil {
			return nil, err
		}
	}

	return cookie, nil
}

func (s *sessionService) Fetch(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userID, err := s.sessionCache.Get(ctx, sessionID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(userID)
}

func (s *sessionService) FetchAll(ctx context.Context, sessionID string) ([]string, error) {
	userID, err := s.Fetch(ctx, sessionID)
	if err != nil {
		log.Printf("sessionID %s not found in cache", sessionID)
		return nil, err
	}
	// iterate SMembers and validate session is still valid
	sessionIDs, err := s.sessionCache.SMembers(ctx, userID.String())
	if err != nil {
		log.Printf("error fetching sessions for user %s: %v", userID.String(), err)
		return nil, err
	}

	// Index to keep track of the current position in the slice
	validCount := 0

	// remove invalid sessions
	for _, id := range sessionIDs {
		if _, err := s.sessionCache.Get(ctx, id); err != nil {
			// remove invalid session from set
			s.sessionCache.SRem(ctx, userID.String(), id)
		} else {
			// Overwrite the current position in the slice with the valid session
			sessionIDs[validCount] = id
			validCount++
		}
	}

	// Trim the slice to remove any remaining elements after validCount
	sessionIDs = sessionIDs[:validCount]

	return sessionIDs, nil
}

func (s *sessionService) Create(ctx context.Context, userID uuid.UUID) (*http.Cookie, error) {
	sessionID := generateSessionID()
	expiration := maxAgeToExpiration(s.sessionConfig.MaxAge)
	if err := s.sessionCache.Set(ctx, sessionID, userID.String(), expiration); err != nil {
		return nil, err
	}

	if err := s.sessionCache.SAdd(ctx, userID.String(), sessionID); err != nil {
		return nil, err
	}

	return s.newCookie(sessionID), nil
}

// Extend updates the expiration of the session in the session cache and
func (s *sessionService) Extend(ctx context.Context, userID string, cookie *http.Cookie) (*http.Cookie, error) {
	s.modifyCookie(cookie)
	return cookie, s.sessionCache.Set(ctx, cookie.Value, userID, maxAgeToExpiration(s.sessionConfig.MaxAge))
}

// base64 encoded 32 byte random string
// Note: base64 converts binary data into a string of characters from a set of 64 characters.
// Each character in the string represents 6 bits of data. Since 32 bytes is equivalent to 256 bits,
// the base64 encoded string will be 256/6 = 42.67 characters long. In base64 encoding, padding is used
// to ensure that the encoded output contains a multiple of 4 characters. Thus the length of the encoded
// string will be 44 characters.
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
