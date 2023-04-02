package service

import (
	"auth/src/model"
	repo "auth/src/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(username string, password string) (*model.User, error)
	Create(ctx context.Context, username string, password []byte) (*model.User, error)
	Logout(username string) error
	Remove(username string) error
	Exists(ctx context.Context, username string) bool
}

type authService struct {
	userRepository repo.UserRepository
}

func NewAuthService(c *repo.DbClient) AuthService {
	return &authService{
		userRepository: repo.NewUserRepository(c),
	}
}

func (s *authService) Exists(ctx context.Context, username string) bool {
	return s.userRepository.Exists(ctx, username)
}

// Assumes username is not taken
func (s *authService) Create(ctx context.Context, username string, password []byte) (*model.User, error) {
	if len(password) > 72 {
		return nil, errors.New("password too long")
	}
	// TODO: make cost configurable, should be 12+ in prod env
	// https://stackoverflow.com/a/6833165/714618
	hashedPass, err := bcrypt.GenerateFromPassword(password, 10)
	if err != nil {
		return nil, err
	}
	newUsr := &model.User{
		Id:       uuid.New(),
		Username: username,
		Password: string(hashedPass),
	}
	if err := s.userRepository.CreateUser(ctx, newUsr); err != nil {
		return nil, err
	}
	return newUsr, nil
}

func (s *authService) Login(username string, password string) (usr *model.User, err error) {
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
