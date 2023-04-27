package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/dgyurics/auth/auth-server/model"
)

// UserRepository is an interface for interacting with the user table
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	ExistsUser(ctx context.Context, username string) bool
	GetUser(ctx context.Context, user *model.User) error
}

type userRepository struct {
	c *DbClient
}

// NewUserRepository creates a new user repository
func NewUserRepository(c *DbClient) UserRepository {
	return &userRepository{c}
}

func (r *userRepository) ExistsUser(ctx context.Context, username string) bool {
	if err := r.GetUser(ctx, &model.User{Username: username}); err != nil {
		return false
	}
	return true
}

func (r *userRepository) GetUser(ctx context.Context, user *model.User) error {
	var arg, query string
	if user.Username != "" {
		query = "SELECT id, username, password FROM auth.user WHERE username = $1"
		arg = user.Username
	} else {
		query = "SELECT id, username, password FROM auth.user WHERE id = $1"
		arg = user.ID.String()
	}

	return r.c.connPool.QueryRowContext(ctx, query, arg).Scan(&user.ID, &user.Username, &user.Password)
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	connPool := r.c.connPool
	tx, err := connPool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmtEvents, err := tx.Prepare("INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)")
	if err != nil {
		rollback(tx)
		return err
	}
	defer closeStmt(stmtEvents)

	// stringify user for event body
	userEncoded, err := json.Marshal(model.OmitPassword(user))
	if err != nil {
		rollback(tx)
		return err
	}

	_, err = stmtEvents.Exec(user.ID, model.AccountCreated, userEncoded)
	if err != nil {
		rollback(tx)
		return err
	}

	stmtUser, err := tx.Prepare("INSERT INTO auth.user (id, username, password) VALUES ($1, $2, $3)")
	if err != nil {
		rollback(tx)
		return err
	}
	defer closeStmt(stmtUser)

	if _, err = stmtUser.Exec(user.ID, user.Username, user.Password); err != nil {
		rollback(tx)
		return err
	}

	return tx.Commit()
}

func rollback(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		log.Println(err)
	}
}

// https://go.dev/doc/database/prepared-statements
func closeStmt(stmt *sql.Stmt) {
	if err := stmt.Close(); err != nil {
		log.Println(err)
	}
}
