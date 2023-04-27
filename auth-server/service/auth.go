package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/dgyurics/auth/auth-server/model"
	"github.com/dgyurics/auth/auth-server/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService is an interface for authentication related operations.
type AuthService interface {
	Login(ctx context.Context, user *model.User) error
	Create(ctx context.Context, user *model.User) error
	Logout(ctx context.Context, user *model.User) error
	Exists(ctx context.Context, user *model.User) bool // FIXME remove and combine into single transaction with Create
	Fetch(ctx context.Context, user *model.User) error
	ValidateUserInput(user *model.User) error
}

type authService struct {
	userRepository  repository.UserRepository
	eventRepository repository.EventRepository
}

// NewAuthService creates a new AuthService with the given user + event repositories.
func NewAuthService(
	userRepository repository.UserRepository,
	eventRepository repository.EventRepository,
) AuthService {
	return &authService{
		userRepository,
		eventRepository,
	}
}

func (s *authService) Exists(ctx context.Context, user *model.User) bool {
	return s.userRepository.ExistsUser(ctx, user.Username)
}

// Create creates a new user with a unique UUID and a bcrypt-hashed password.
// The new user is stored in the underlying user repository, and the user's ID and password
// are updated with the new values.
//
// Returns an error if there is an issue generating the password hash or creating the user in the repository.
func (s *authService) Create(ctx context.Context, user *model.User) error {
	// TODO: make cost configurable, should be 12+ in prod env
	// https://stackoverflow.com/a/6833165/714618
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return err
	}
	user.ID = uuid.New()
	user.Password = string(hashedPass)
	return s.userRepository.CreateUser(ctx, user)
}

// Login attempts to authenticate the given user by retrieving their stored password hash and comparing it
// to the provided password hash using the bcrypt algorithm. If the hashes match, the user's ID is set and
// the LoginSuccess method is called on the underlying user repository.
//
// Returns an error if the user cannot be retrieved or the password hashes do not match.
func (s *authService) Login(ctx context.Context, user *model.User) error {
	userCpy := *user
	if err := s.userRepository.GetUser(ctx, &userCpy); err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userCpy.Password), []byte(user.Password)); err != nil {
		return err
	}
	user.ID = userCpy.ID

	// stringify user for event body
	userEncoded, err := json.Marshal(model.OmitPassword(user))
	if err != nil {
		return err
	}
	return s.eventRepository.CreateEvent(ctx, &model.Event{
		UUID: user.ID,
		Type: model.LoggedIn,
		Body: userEncoded,
	})
}

func (s *authService) Logout(ctx context.Context, user *model.User) error {
	// stringify user for event body
	userEncoded, err := json.Marshal(model.OmitPassword(user))
	if err != nil {
		return err
	}
	return s.eventRepository.CreateEvent(ctx, &model.Event{
		UUID: user.ID,
		Type: model.LoggedOut,
		Body: userEncoded,
	})
}

func (s *authService) Fetch(ctx context.Context, user *model.User) error {
	return s.userRepository.GetUser(ctx, user)
}

func (s *authService) ValidateUserInput(user *model.User) error {
	if user.Username == "" {
		return errors.New("username cannot be empty")
	}
	// Strings are UTF-8 encoded, this means each charcter aka rune can be of 1 to 4 bytes long
	if len(user.Username) > 50 {
		return errors.New("username length cannot exceed 50 characters")
	}
	if len(user.Password) < 1 || len(user.Password) > 72 {
		return errors.New("password length must be between 1 and 72 characters")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(user.Username) {
		return errors.New("username must be alphanumeric")
	}
	return nil
}
