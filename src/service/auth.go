package service

import (
	"context"

	"github.com/dgyurics/auth/src/model"
	"github.com/dgyurics/auth/src/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, user *model.User) error
	Create(ctx context.Context, user *model.User) error
	Logout(ctx context.Context, user *model.User) error
	Remove(ctx context.Context, user *model.User) error
	Exists(ctx context.Context, user *model.User) bool // FIXME remove and combine into single transaction with Create
	Fetch(ctx context.Context, user *model.User) error
}

type authService struct {
	userRepository repository.UserRepository
}

func NewAuthService(userRepository repository.UserRepository) AuthService {
	return &authService{
		userRepository,
	}
}

func (s *authService) Exists(ctx context.Context, user *model.User) bool {
	return s.userRepository.Exists(ctx, user.Username)
}

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

func (s *authService) Login(ctx context.Context, user *model.User) error {
	userCpy := *user
	if err := s.userRepository.GetUser(ctx, &userCpy); err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userCpy.Password), []byte(user.Password)); err != nil {
		return err
	}
	user.Id = userCpy.Id
	return s.userRepository.LoginSuccess(ctx, &userCpy)
}

func (s *authService) Logout(ctx context.Context, user *model.User) error {
	return s.userRepository.LogoutUser(ctx, user)
}

func (s *authService) Remove(ctx context.Context, user *model.User) error {
	// TODO
	// verify session is valid
	// store event in db
	// invalidate session
	return nil
}

func (s *authService) Fetch(ctx context.Context, user *model.User) error {
	return s.userRepository.GetUser(ctx, user)
}
