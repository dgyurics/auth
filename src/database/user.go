package database

import (
	srv "auth/src/service"
)

const USER_CREATE_TYPE = "user_create"
const USER_LOGIN_TYPE = "user_login"
const USER_LOGOUT_TYPE = "user_logout"
const USER_DELETE_TYPE = "user_delete"

func (c *dbClient) GetUser(username string) (user *srv.User, err error) {
	err = c.connPool.QueryRow("SELECT id, username, password FROM user WHERE username = $1", username).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *dbClient) RemoveUser(username string) error {
	// TODO create event
	_, err := c.connPool.Exec("DELETE FROM user WHERE username = $1", username)
	return err
}

func (c *dbClient) UpdateUser(usr *srv.User) error {
	// TODO create event
	_, err := c.connPool.Exec("UPDATE user SET username = $1, password = $2 WHERE id = $3", usr.Username, usr.Password, usr.Id)
	return err
}

// FIXME: wrap in single transaction
func (c *dbClient) CreateUser(usr *srv.User) error {
	row := c.connPool.QueryRow("INSERT INTO events (uuid, type, body) VALUES ($1, $2, $3)",
		usr.Id, USER_CREATE_TYPE, usr) // FIXME remove password from event body
	if err := row.Err(); err != nil {
		return err
	}
	row = c.connPool.QueryRow("INSERT INTO user (id, username, password) VALUES ($1, $2, $3)", usr.Id, usr.Username, usr.Password)
	if err := row.Err(); err != nil {
		return err
	}
	return nil
}
