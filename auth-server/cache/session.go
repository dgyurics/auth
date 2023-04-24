package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// SessionCache is an interface for interacting with Redis.
type SessionCache interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type sessionCache struct {
	c *redis.Client
}

// NewSessionCache returns a new instance of SessionCache.
func NewSessionCache(c *redis.Client) SessionCache {
	return &sessionCache{c: c}
}

func (s *sessionCache) Del(ctx context.Context, key string) error {
	return s.c.Del(ctx, key).Err()
}

func (s *sessionCache) Set(ctx context.Context, key string, value string) error {
	// TODO: set expiration
	// expiration time.Duration
	// obtain from config
	return s.c.Set(ctx, key, value, 0).Err()
}

func (s *sessionCache) Get(ctx context.Context, key string) (string, error) {
	return s.c.Get(ctx, key).Result()
}
