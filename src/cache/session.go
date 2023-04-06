package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type SessionCache interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
}

type sessionCache struct {
	c *redis.Client
}

func NewSessionCache(c *redis.Client) *sessionCache {
	return &sessionCache{c: c}
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
