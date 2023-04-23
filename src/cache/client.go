package cache

import (
	"github.com/dgyurics/auth/src/config"
	"github.com/go-redis/redis/v8"
)

// NewClient creates a new redis client
func NewClient(config config.Redis) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})
}
