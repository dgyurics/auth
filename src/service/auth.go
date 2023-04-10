package service

import (
	"auth/src/model"
	repo "auth/src/repository"
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, usr *model.User) error
	Create(ctx context.Context, usr *model.User) error
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
func (s *authService) Create(ctx context.Context, user *model.User) error {
	// TODO: make cost configurable, should be 12+ in prod env
	// https://stackoverflow.com/a/6833165/714618
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return err
	}
	user.Id = uuid.New()
	user.Password = string(hashedPass)
	if err := s.userRepository.CreateUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *authService) Login(ctx context.Context, usr *model.User) error {
	usrRec, err := s.userRepository.GetUserByUsername(ctx, usr.Username)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(usrRec.Password), []byte(usr.Password)); err != nil {
		return err
	}
	return nil
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
