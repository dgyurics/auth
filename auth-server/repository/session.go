package repository

import (
	"context"

	"github.com/google/uuid"
)

// SessionRepository is an interface for interacting with the session table
type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, sessionID string) error
	RemoveSession(ctx context.Context, sessionID string) error
	RemoveSessions(ctx context.Context, userID string) error
}

type sessionRepository struct {
	c *DbClient
}

// Placeholder
// Not implemented yet
