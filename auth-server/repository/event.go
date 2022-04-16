package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/dgyurics/auth/auth-server/model"
)

// EventRepository is an interface for interacting with the event table
type EventRepository interface {
	CreateEvent(ctx context.Context, event *model.Event) error
	Close() error
}

type eventRepository struct {
	*DbClient
	stmtInsertEvent *sql.Stmt // Prepared statement for inserting into auth.event
}

// NewEventRepository creates a new event repository
func NewEventRepository(c *DbClient) EventRepository {
	repo := &eventRepository{
		DbClient: c,
	}
	repo.prepareStatements()
	return repo
}

func (r *eventRepository) CreateEvent(ctx context.Context, event *model.Event) error {
	_, err := r.stmtInsertEvent.ExecContext(ctx, event.UUID, event.Type, event.Body)
	return err
}

func (r *eventRepository) prepareStatements() {
	var err error
	r.stmtInsertEvent, err = r.connPool.Prepare(`
		INSERT INTO auth.event (uuid, type, body)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *eventRepository) Close() error {
	return r.stmtInsertEvent.Close()
}
