package cache

import (
	"context"
	"time"
)

// MockSessionCache is a mock implementation of SessionCache.
type MockSessionCache struct {
	Sessions map[string]string
}

// Del deletes a session from the cache.
func (s *MockSessionCache) Del(_ context.Context, key string) error {
	delete(s.Sessions, key)
	return nil
}

// Set sets a session in the cache.
func (s *MockSessionCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	s.Sessions[key] = value
	return nil
}

// Get gets a session from the cache.
func (s *MockSessionCache) Get(_ context.Context, key string) (string, error) {
	value, ok := s.Sessions[key]
	if !ok {
		return "", nil
	}
	return value, nil
}
