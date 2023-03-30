package service

import (
	"auth/src/model"
	repo "auth/src/repository"
)

type AuthService interface {
	Login(username string, password string) (model.User, error)
	Create(username string, password string) (model.User, error)
	Logout(username string) error
	Remove(username string) error
	Exists(username string) bool
}

type authService struct {
	userRepository repo.UserRepository
}

func NewAuthService(c *repo.DbClient) AuthService {
	return &authService{
		userRepository: repo.NewUserRepository(c),
	}
}

func (s *authService) Exists(username string) bool {
	// TODO
	// verify username is unique
	return false
}

func (s *authService) Create(username string, password string) (usr model.User, err error) {
	// TODO
	// verify username is unique
	// generate Id
	// salt + encrypt password
	// store event in db
	return usr, nil
}

func (s *authService) Login(username string, password string) (usr model.User, err error) {
	// TODO
	// verify username and password match
	// store event in db
	// create session
	return usr, nil
}

func (s *authService) Logout(username string) error {
	// TODO
	// verify session is valid
	// store event in db
	// invalidate session
	return nil
}

func (s *authService) Remove(username string) error {
	// TODO
	// verify session is valid
	// store event in db
	// invalidate session
	return nil
}
