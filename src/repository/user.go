package repository

import (
	"context"
	"encoding/json"

	"github.com/dgyurics/auth/src/config"
	"github.com/dgyurics/auth/src/model"
)

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

func NewUserRepository() UserRepository {
	c := NewDBClient()
	c.Connect(config.New().PostgreSql)
	return &userRepository{c}
}

const USER_CREATE_TYPE = "user_create"
const USER_LOGIN_TYPE = "user_login"
const USER_LOGOUT_TYPE = "user_logout"
const USER_DELETE_TYPE = "user_delete"

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
		arg = user.Id.String()
	}

	if err := r.c.connPool.QueryRowContext(ctx, query, arg).Scan(&user.Id, &user.Username, &user.Password); err != nil {
		return err
	}
	return nil
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
	defer stmtEvents.Close() // https://go.dev/doc/database/prepared-statements

	// stringify user for event body
	stringifyuser, err := json.Marshal(OmitPassword(user))
	if err != nil {
		return err
	}

	if _, err = stmtEvents.Exec(user.Id, USER_LOGIN_TYPE, stringifyuser); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
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
	defer stmtEvents.Close() // https://go.dev/doc/database/prepared-statements

	if _, err = stmtEvents.Exec(user.Id, USER_LOGOUT_TYPE); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// FIXME: wrap in single transaction
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	connPool := r.c.connPool
	tx, err := connPool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	stmtEvents, err := tx.Prepare("INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmtEvents.Close() // https://go.dev/doc/database/prepared-statements

	// stringify user for event body
	stringifyuser, err := json.Marshal(OmitPassword(user))
	if err != nil {
		return err
	}

	_, err = stmtEvents.Exec(user.Id, USER_CREATE_TYPE, stringifyuser)
	if err != nil {
		return err
	}

	stmtUser, err := tx.Prepare("INSERT INTO auth.user (id, username, password) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmtUser.Close()
	if _, err = stmtUser.Exec(user.Id, user.Username, user.Password); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// creates a copy of the user with the password field set to ""
func OmitPassword(user *model.User) *model.User {
	return &model.User{
		Id:       user.Id,
		Username: user.Username,
		Password: "",
	}
}
