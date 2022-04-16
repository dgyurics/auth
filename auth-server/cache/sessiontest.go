package cache

import (
	"context"
	"time"
)

// TODO add unimplemented methods

// MockSessionCache is a mock implementation of SessionCache.
type MockSessionCache struct {
	Sessions    map[string]string
	SessionsSet map[string]map[string]struct{}
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

// Del deletes a session from the cache.
func (s *MockSessionCache) Del(_ context.Context, key string) error {
	delete(s.Sessions, key)
	return nil
}

func (s *MockSessionCache) SAdd(_ context.Context, key string, value string) error {
	if _, ok := s.SessionsSet[key]; !ok {
		s.SessionsSet[key] = make(map[string]struct{})
	}
	s.SessionsSet[key][value] = struct{}{}
	return nil
}

func (s *MockSessionCache) SRem(_ context.Context, key string, value string) error {
	delete(s.SessionsSet[key], value)
	return nil
}

func (s *MockSessionCache) SMembers(_ context.Context, key string) ([]string, error) {
	var members []string
	for member := range s.SessionsSet[key] {
		members = append(members, member)
	}
	return members, nil
}

func (s *MockSessionCache) SCard(_ context.Context, key string) (int64, error) {
	return int64(len(s.SessionsSet[key])), nil
}
