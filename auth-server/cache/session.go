package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// SessionCache is an interface for interacting with Redis.
type SessionCache interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	KeyspaceNotifications(ctx context.Context)
}

// fixme refactor to use
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

func (s *sessionCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return s.c.Set(ctx, key, value, expiration).Err()
}

func (s *sessionCache) Get(ctx context.Context, key string) (string, error) {
	return s.c.Get(ctx, key).Result()
}

func (s *sessionCache) KeyspaceNotifications(ctx context.Context) {
	// this is telling redis to publish events since it's off by default.
	_, err := s.c.Do(context.Background(), "CONFIG", "SET", "notify-keyspace-events", "KEA").Result()
	if err != nil {
		log.Fatalf("unable to set keyspace events %v", err.Error())
	}

	// this is telling redis to subscribe to events published in the keyevent channel, specifically for expired events
	// TODO 0 should be replaced with the database number
	pubsub := s.c.PSubscribe(context.Background(), "__keyevent@0__:expired")

	// this is just to show publishing events and catching the expired events in the same codebase
	// FIXME can I remove this?
	wg := &sync.WaitGroup{}
	wg.Add(2) // two goroutines are spawned

	go func(redis.PubSub) {
		exitLoopCounter := 0
		for { // infinite loop
			// this listens in the background for messages.
			message, err := pubsub.ReceiveMessage(context.Background())
			exitLoopCounter++
			if err != nil {
				log.Fatalf("fatal error while listening for keyspace events %v", err.Error())
				break
			}
			fmt.Printf("Keyspace event recieved %v  \n", message.String())
			// TODO remove session from postgres

			// if exitLoopCounter >= 10 {
			// 	wg.Done()
			// }
		}
	}(*pubsub)

	// FIXME can I remove this?
	wg.Wait()
	fmt.Println("exiting KeyspaceNotifications")
}
