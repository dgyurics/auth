package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/dgyurics/auth/auth-server/model"
	"github.com/google/uuid"
)

// SessionRepository is an interface for interacting with the session table
type SessionRepository interface {
	CreateSession(ctx context.Context, session *model.Session) error
	GetSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error)
	RemoveSession(ctx context.Context, sessionID string) error
	RemoveSessions(ctx context.Context, userID uuid.UUID) error
	Close() error
}

type sessionRepository struct {
	*DbClient
	stmtInsertSession  *sql.Stmt // Prepared statement for inserting into auth.session
	stmtDeleteSession  *sql.Stmt // Prepared statement for deleting from auth.session
	stmtDeleteSessions *sql.Stmt // Prepared statement for deleting from auth.session
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(c *DbClient) SessionRepository {
	repo := &sessionRepository{
		DbClient: c,
	}
	repo.prepareStatements()
	return repo
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *model.Session) error {
	_, err := r.stmtInsertSession.ExecContext(ctx, session.ID, session.UserID)
	return err
}

func (r *sessionRepository) RemoveSession(ctx context.Context, sessionID string) error {
	_, err := r.stmtDeleteSession.ExecContext(ctx, sessionID)
	return err
}

func (r *sessionRepository) RemoveSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.stmtDeleteSessions.ExecContext(ctx, userID)
	return err
}

func (r *sessionRepository) GetSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	rows, err := r.connPool.QueryContext(ctx, `
		SELECT id, user_id, created_at
		FROM auth.session
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := rows.Close(); e != nil {
			log.Fatal(e)
		}
	}()
	var sessions []*model.Session
	for rows.Next() {
		var session model.Session
		if err := rows.Scan(&session.ID, &session.UserID, &session.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

func (r *sessionRepository) prepareStatements() {
	var err error
	r.stmtInsertSession, err = r.connPool.Prepare(`
		INSERT INTO auth.session (id, user_id)
		VALUES ($1, $2)
	`)
	if err != nil {
		log.Fatal(err)
	}
	r.stmtDeleteSession, err = r.connPool.Prepare(`
		DELETE FROM auth.session
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	r.stmtDeleteSessions, err = r.connPool.Prepare(`
		DELETE FROM auth.session
		WHERE user_id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// https://go.dev/doc/database/prepared-statements
func (r *sessionRepository) Close() error {
	var err error
	if e := r.stmtInsertSession.Close(); e != nil {
		err = e
	}
	if e := r.stmtDeleteSession.Close(); e != nil {
		err = e
	}
	if e := r.stmtDeleteSessions.Close(); e != nil {
		err = e
	}
	return err
}
