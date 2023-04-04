package service

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

type SessionService interface {
	Create(userId string) string
	FetchUserId(sessionId string) error
	Invalidate(sessionId string) error
}

type sessionService struct{}

// FIXME: will need client for in-memory db
func NewSessionService() SessionService {
	return &sessionService{}
}

func (s *sessionService) Invalidate(sessionId string) error {
	// TODO
	return nil
}

func (s *sessionService) FetchUserId(sessionId string) error {
	// TODO
	return nil
}

// session id is a base64 encoded 32 byte random string
func (s *sessionService) Create(userId string) string {
	// TOOD check for collisions
	// TODO: store session id + userId in in-memory db
	return generateSessionId()
}

func generateSessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
