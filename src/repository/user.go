package repository

import "auth/src/model"

type UserRepository interface {
	CreateUser(usr *model.User) (model.User, error)
	GetUserByUsername(username string) (model.User, error)
	RemoveUserByUsername(username string) error
	UpdateUser(usr *model.User) (model.User, error)
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

func (r *userRepository) GetUserByUsername(username string) (usr model.User, err error) {
	err = r.c.connPool.QueryRow("SELECT id, username, password FROM user WHERE username = $1", username).Scan(&usr.Id, &usr.Username, &usr.Password)
	if err != nil {
		return usr, err
	}
	return usr, nil
}

func (r *userRepository) RemoveUserByUsername(usrName string) error {
	// TODO create event
	_, err := r.c.connPool.Exec("DELETE FROM user WHERE username = $1", usrName)
	return err
}

func (r *userRepository) UpdateUser(usr *model.User) (updatedUsr model.User, err error) {
	// TODO create event
	_, err = r.c.connPool.Exec("UPDATE user SET username = $1, password = $2 WHERE id = $3", usr.Username, usr.Password, usr.Id)
	return updatedUsr, err // FIXME
}

// FIXME: wrap in single transaction
func (r *userRepository) CreateUser(usr *model.User) (newUsr model.User, err error) {
	row := r.c.connPool.QueryRow("INSERT INTO events (uuid, type, body) VALUES ($1, $2, $3)",
		usr.Id, USER_CREATE_TYPE, usr) // FIXME remove password from event body
	if err = row.Err(); err != nil {
		return *usr, err
	}
	row = r.c.connPool.QueryRow("INSERT INTO user (id, username, password) VALUES ($1, $2, $3)", usr.Id, usr.Username, usr.Password)
	if err = row.Err(); err != nil {
		return *usr, err
	}
	return newUsr, nil // FIXME return user from db
}
