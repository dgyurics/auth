package repository

import (
	"context"

	"github.com/dgyurics/auth/auth-server/model"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *model.Event) error
}

type eventDBRepo struct {
	c *DbClient
}

func NewEventRepository(c *DbClient) EventRepository {
	return &eventDBRepo{c}
}

func (r *eventDBRepo) CreateEvent(ctx context.Context, event *model.Event) error {
	_, err := r.c.connPool.ExecContext(ctx, "INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)", event.UUID, event.Type, event.Body)
	return err
}
