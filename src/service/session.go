package service

import (
	"auth/src/cache"
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/go-redis/redis/v8"
)

type SessionService interface {
	Create(ctx context.Context, userId string) string
	FetchUserId(sessionId string) error
	Invalidate(ctx context.Context, sessionId string) error
}

type sessionService struct {
	sessionCache cache.SessionCache
}

// FIXME: will need client for in-memory db
func NewSessionService(c *redis.Client) SessionService {
	return &sessionService{
		sessionCache: cache.NewSessionCache(c),
	}
}

func (s *sessionService) Invalidate(ctx context.Context, sessionId string) error {
	return s.sessionCache.Del(ctx, sessionId)
}

func (s *sessionService) FetchUserId(sessionId string) error {
	// TODO
	return nil
}

func (s *sessionService) Create(ctx context.Context, userId string) string {
	// TOOD verify likeliehood of collision
	// TODO prevent user from creating too many sessions
	sessionId := generateSessionId()
	s.sessionCache.Set(ctx, sessionId, userId)
	return sessionId
}

// base64 encoded 32 byte random string
func generateSessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
