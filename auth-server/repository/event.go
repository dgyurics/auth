package repository

import (
	"context"
	"log"

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
	repo.initTable()
	return repo
}

func (r *eventDBRepo) CreateEvent(ctx context.Context, event *model.Event) error {
	_, err := r.c.connPool.ExecContext(ctx, "INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)", event.UUID, event.Type, event.Body)
	return err
}

func (r *eventDBRepo) initTable() {
	_, err := r.c.connPool.Exec("CREATE SCHEMA IF NOT EXISTS auth")
	if err != nil {
		log.Fatal(err)
	}
	_, err = r.c.connPool.Exec(`CREATE TABLE IF NOT EXISTS "auth"."event" (
		"id"         serial PRIMARY KEY not NULL,
		"uuid"       uuid NOT NULL,
		"type"       text NOT NULL,
		"body"       jsonb,
		"created_at" timestamp without time zone DEFAULT (now() at time zone 'utc')
	);`)
	if err != nil {
		log.Fatal(err)
	}
}
