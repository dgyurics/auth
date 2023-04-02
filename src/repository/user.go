package repository

import (
	"auth/src/model"
	"context"
	"encoding/json"
)

type UserRepository interface {
	CreateUser(usr *model.User) error
	GetUserByUsername(username string) (*model.User, error)
	RemoveUserByUsername(username string) error
	UpdateUser(usr *model.User) (*model.User, error)
	Exists(username string) bool
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

func (r *userRepository) Exists(username string) bool {
	if _, err := r.GetUserByUsername(username); err != nil {
		return false
	}
	return true
}

func (r *userRepository) GetUserByUsername(username string) (usr *model.User, err error) {
	tmp := model.User{}
	err = r.c.connPool.QueryRow("SELECT id, username, password FROM auth.user WHERE username = $1", username).Scan(&tmp.Id, &tmp.Username, &tmp.Password)
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

// FIXME: wrap in single transaction
func (r *userRepository) CreateUser(usr *model.User) error {
	connPool := r.c.connPool
	ctx := context.TODO() // FIXME: placeholder while I wire up request context
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
	stringifyUsr, err := json.Marshal(omitPassword(usr))
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
func omitPassword(usr *model.User) model.User {
	usrCopy := *usr
	usrCopy.Password = ""
	return usrCopy
}
