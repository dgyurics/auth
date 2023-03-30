package service

import (
	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
}

type UserService interface {
	Login(username string, password string) error
	Create(username string, password string) error
	Logout(username string) error
	Remove(username string) error
}

func (usr *User) Create() error {
	// TODO
	// verify username is unique
	// generate Id
	// salt + encrypt password
	// store event in db
	return nil
}

func (usr *User) Login() error {
	// TODO
	// verify username and password match
	// store event in db
	// create session
	return nil
}

func (usr *User) Logout() error {
	// TODO
	// verify session is valid
	// store event in db
	// invalidate session
	return nil
}

func (usr *User) Remove() error {
	// TODO
	// verify session is valid
	// store event in db
	// invalidate session
	return nil
}
