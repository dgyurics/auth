package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/dgyurics/auth/src/config"
	"github.com/dgyurics/auth/src/model"
)

// TODO Create a transaction manager abstraction to encapsulate the transaction logic.
// reference: https://dev.to/techschoolguru/a-clean-way-to-implement-database-transaction-in-golang-2ba

// UserRepository is an interface for interacting with the user table
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	Exists(ctx context.Context, username string) bool
	GetUser(ctx context.Context, user *model.User) error
	LoginSuccess(ctx context.Context, user *model.User) error
	LogoutUser(ctx context.Context, user *model.User) error
}

type userRepository struct {
	c *DbClient
}

// NewUserRepository creates a new user repository
func NewUserRepository() UserRepository {
	c := NewDBClient()
	c.Connect(config.New().PostgreSQL)
	return &userRepository{c}
}

const (
	userCreateType = "user_create"
	userLoginType  = "user_login"
	userLogoutType = "user_logout"
)

func (r *userRepository) Exists(ctx context.Context, username string) bool {
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

func (r *userRepository) LoginSuccess(ctx context.Context, user *model.User) error {
	connPool := r.c.connPool
	tx, err := connPool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmtEvents, err := tx.Prepare("INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer closeStmt(stmtEvents)

	// stringify user for event body
	stringifyuser, err := json.Marshal(OmitPassword(user))
	if err != nil {
		return err
	}

	if _, err = stmtEvents.Exec(user.ID, userLoginType, stringifyuser); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *userRepository) LogoutUser(ctx context.Context, user *model.User) error {
	connPool := r.c.connPool
	tx, err := connPool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmtEvents, err := tx.Prepare("INSERT INTO auth.event (uuid, type) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer closeStmt(stmtEvents)

	if _, err = stmtEvents.Exec(user.ID, userLogoutType); err != nil {
		return err
	}
	return tx.Commit()
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
	stringifyuser, err := json.Marshal(OmitPassword(user))
	if err != nil {
		rollback(tx)
		return err
	}

	_, err = stmtEvents.Exec(user.ID, userCreateType, stringifyuser)
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

// OmitPassword creates a copy of the user with the password field set to ""
func OmitPassword(user *model.User) *model.User {
	return &model.User{
		ID:       user.ID,
		Username: user.Username,
		Password: "",
	}
}
