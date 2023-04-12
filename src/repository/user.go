package repository

import (
	"auth/src/model"
	"context"
	"encoding/json"
)

type UserRepository interface {
	CreateUser(ctx context.Context, usr *model.User) error
	LoginSuccess(ctx context.Context, usr *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUser(ctx context.Context, usr *model.User) error
	RemoveUserByUsername(username string) error
	UpdateUser(usr *model.User) (*model.User, error)
	Exists(ctx context.Context, username string) bool
	LogoutUser(ctx context.Context, usr *model.User) error
}

type userRepository struct {
	c *DbClient
}

func NewUserRepository(c *DbClient) UserRepository {
	return &userRepository{c}
}

const USER_CREATE_TYPE = "user_create"
const USER_LOGIN_TYPE = "user_login"
const USER_LOGOUT_TYPE = "user_logout"
const USER_DELETE_TYPE = "user_delete"

func (r *userRepository) Exists(ctx context.Context, username string) bool {
	if _, err := r.GetUserByUsername(ctx, username); err != nil {
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

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (usr *model.User, err error) {
	tmp := model.User{}
	err = r.c.connPool.QueryRowContext(ctx, "SELECT id, username, password FROM auth.user WHERE username = $1", username).Scan(&tmp.Id, &tmp.Username, &tmp.Password)

	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (r *userRepository) RemoveUserByUsername(usrName string) error {
	// TODO create event
	_, err := r.c.connPool.Exec("DELETE FROM user WHERE auth.username = $1", usrName)
	return err
}

func (r *userRepository) UpdateUser(usr *model.User) (updatedUsr *model.User, err error) {
	// TODO create event
	_, err = r.c.connPool.Exec("UPDATE auth.user SET username = $1, password = $2 WHERE id = $3", usr.Username, usr.Password, usr.Id)
	return updatedUsr, err // FIXME
}

func (r *userRepository) LoginSuccess(ctx context.Context, usr *model.User) error {
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
	OmitPassword(usr)
	stringifyUsr, err := json.Marshal(usr)
	if err != nil {
		return err
	}

	if _, err = stmtEvents.Exec(usr.Id, USER_LOGIN_TYPE, stringifyUsr); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *userRepository) LogoutUser(ctx context.Context, usr *model.User) error {
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
	OmitPassword(usr)
	stringifyUsr, err := json.Marshal(usr)
	if err != nil {
		return err
	}

	if _, err = stmtEvents.Exec(usr.Id, USER_LOGOUT_TYPE, stringifyUsr); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// FIXME: wrap in single transaction
func (r *userRepository) CreateUser(ctx context.Context, usr *model.User) error {
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
	OmitPassword(usr)
	stringifyUsr, err := json.Marshal(usr)
	if err != nil {
		return err
	}

	_, err = stmtEvents.Exec(usr.Id, USER_CREATE_TYPE, stringifyUsr)
	if err != nil {
		return err
	}

	stmtUser, err := tx.Prepare("INSERT INTO auth.user (id, username, password) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmtUser.Close()
	if _, err = stmtUser.Exec(usr.Id, usr.Username, usr.Password); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// creates a copy of the user with the password field set to ""
func OmitPassword(usr *model.User) {
	usr.Password = ""
}
