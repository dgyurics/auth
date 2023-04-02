package repository

import (
	"auth/src/model"
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
	// row := r.c.connPool.QueryRow("INSERT INTO auth.events (uuid, type, body) VALUES ($1, $2, $3)",
	// 	usr.Id, USER_CREATE_TYPE, usr) // FIXME remove password from event body
	// if err = row.Err(); err != nil {
	// 	return usr, err
	// }
	stmt, err := connPool.Prepare("INSERT INTO auth.user (id, username, password) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close() // https://go.dev/doc/database/prepared-statements
	_, err = stmt.Exec(usr.Id, usr.Username, usr.Password)
	return err
}
