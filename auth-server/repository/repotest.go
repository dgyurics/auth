package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgyurics/auth/auth-server/model"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	Users []*model.User
}

// CreateUser creates a new user
func (r *MockUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := r.GetUser(ctx, user)
	if err == nil {
		return errors.New("user already exists")
	}
	r.Users = append(r.Users, user)
	return nil
}

// ExistsUser checks if a user exists
func (r *MockUserRepository) ExistsUser(_ context.Context, username string) bool {
	for _, u := range r.Users {
		if u.Username == username {
			return true
		}
	}
	return false
}

// GetUser gets a user by username or ID
func (r *MockUserRepository) GetUser(_ context.Context, user *model.User) error {
	for _, u := range r.Users {
		if u.Username == user.Username {
			user.ID = u.ID
			user.Password = u.Password
			return nil
		}
		if u.ID == user.ID {
			user.Username = u.Username
			user.Password = u.Password
			return nil
		}
	}
	return errors.New("user not found")
}

// Close closes the repository prepared statements
func (r *MockUserRepository) Close() error {
	return nil
}

// MockEventRepository is a mock implementation of the EventRepository interface
type MockEventRepository struct {
	Events []*model.Event
}

// CreateEvent creates a new event
func (r *MockEventRepository) CreateEvent(_ context.Context, event *model.Event) error {
	r.Events = append(r.Events, event)
	return nil
}

// GenerateUniqueUsername generates a unique username for testing
func GenerateUniqueUsername() string {
	rand.Seed(time.Now().UnixNano()) // nolint:staticcheck
	randomSuffix := rand.Intn(100000)
	return fmt.Sprintf("testuser%d", randomSuffix)
}

// Close closes the repository prepared statements
func (r *MockEventRepository) Close() error {
	return nil
}
