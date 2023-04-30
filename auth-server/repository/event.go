package repository

import (
	"context"

	"github.com/dgyurics/auth/auth-server/model"
)

// EventRepository is an interface for interacting with the event table
type EventRepository interface {
	CreateEvent(ctx context.Context, event *model.Event) error
}

type eventDBRepo struct {
	c *DbClient
}

// NewEventRepository creates a new event repository
func NewEventRepository(c *DbClient) EventRepository {
	repo := &eventDBRepo{c}
	return repo
}

func (r *eventDBRepo) CreateEvent(ctx context.Context, event *model.Event) error {
	_, err := r.c.connPool.ExecContext(ctx, "INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)", event.UUID, event.Type, event.Body)
	return err
}
