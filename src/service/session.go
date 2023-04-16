package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/dgyurics/auth/src/cache"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type SessionService interface {
	Create(ctx context.Context, userId string) (string, error)
	Fetch(ctx context.Context, sessionId string) (uuid.UUID, error)
	Remove(ctx context.Context, sessionId string) error
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

func (s *sessionService) Remove(ctx context.Context, sessionId string) error {
	return s.sessionCache.Del(ctx, sessionId)
}

func (s *sessionService) Fetch(ctx context.Context, sessionId string) (uuid.UUID, error) {
	userId, err := s.sessionCache.Get(ctx, sessionId)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(userId)
}

func (s *sessionService) Create(ctx context.Context, userId string) (string, error) {
	// TODO verify likeliehood of collision
	// TODO prevent user from creating too many sessions
	sessionId := generateSessionId()
	return sessionId, s.sessionCache.Set(ctx, sessionId, userId)
}

// base64 encoded 32 byte random string
func generateSessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
