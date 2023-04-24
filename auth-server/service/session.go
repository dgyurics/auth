package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// SessionService is an interface for session/Redis related operations.
type SessionService interface {
	Create(ctx context.Context, userID string) (string, error)
	Fetch(ctx context.Context, sessionID string) (uuid.UUID, error)
	Remove(ctx context.Context, sessionID string) error
}

type sessionService struct {
	sessionCache cache.SessionCache
}

// NewSessionService creates a new session service with the given redis client.
func NewSessionService(c *redis.Client) SessionService {
	return &sessionService{
		sessionCache: cache.NewSessionCache(c),
	}
}

func (s *sessionService) Remove(ctx context.Context, sessionID string) error {
	return s.sessionCache.Del(ctx, sessionID)
}

func (s *sessionService) Fetch(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userID, err := s.sessionCache.Get(ctx, sessionID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(userID)
}

func (s *sessionService) Create(ctx context.Context, userID string) (string, error) {
	// TODO verify likeliehood of collision
	// TODO prevent user from creating too many sessions
	sessionID := generateSessionID()
	return sessionID, s.sessionCache.Set(ctx, sessionID, userID)
}

// base64 encoded 32 byte random string
func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
