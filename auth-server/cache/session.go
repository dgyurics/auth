package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// SessionCache is an interface for interacting with Redis.
type SessionCache interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	SAdd(ctx context.Context, key string, value string) error
	SRem(ctx context.Context, key string, value string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SCard(ctx context.Context, key string) (int64, error)
	// ExpNotify(ctx context.Context, ch chan string)
}

// fixme refactor to use
type sessionCache struct {
	c *redis.Client
}

// NewSessionCache returns a new instance of SessionCache.
func NewSessionCache(c *redis.Client) SessionCache {
	return &sessionCache{
		c: c,
	}
}

func (s *sessionCache) Del(ctx context.Context, key string) error {
	return s.c.Del(ctx, key).Err()
}

func (s *sessionCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return s.c.Set(ctx, key, value, expiration).Err()
}

func (s *sessionCache) Get(ctx context.Context, key string) (string, error) {
	return s.c.Get(ctx, key).Result()
}

func (s *sessionCache) SAdd(ctx context.Context, key string, value string) error {
	return s.c.SAdd(ctx, key, value).Err()
}

func (s *sessionCache) SRem(ctx context.Context, key string, value string) error {
	return s.c.SRem(ctx, key, value).Err()
}

func (s *sessionCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return s.c.SMembers(ctx, key).Result()
}

func (s *sessionCache) SCard(ctx context.Context, key string) (int64, error) {
	return s.c.SCard(ctx, key).Result()
}

// func (s *sessionCache) ExpNotify(ctx context.Context, ch chan string) {
// 	// enable redis keyspace events. keyspace events are disabled by default
// 	_, err := s.c.Do(ctx, "CONFIG", "SET", "notify-keyspace-events", "KEA").Result()
// 	if err != nil {
// 		log.Fatalf("unable to set keyspace events %v", err.Error())
// 	}

// 	// subscribe to expired events
// 	pubsubExpired := s.c.PSubscribe(ctx, "__keyevent@0__:expired") // FIXME make database 0 configurable
// 	for {
// 		message, err := pubsubExpired.ReceiveMessage(ctx)
// 		if err != nil {
// 			log.Fatalf("fatal error while listening for keyspace events %v", err.Error())
// 			break
// 		}

// 		ch <- message.Payload
// 	}
// }
