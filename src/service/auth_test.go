package service

import (
	"auth/src/model"
	"context"
	"testing"
)

func TestCreate(t *testing.T) {
	// TODO
}

func TestLogin(t *testing.T) {
	// TODO
}

func TestLogout(t *testing.T) {
	// TODO
}

// TODO create fake UserRepository
// TODO move to separate file

type fakeUserRepository struct{}

func (f *fakeUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	return nil
}

func (f *fakeUserRepository) LoginSuccess(ctx context.Context, user *model.User) error {
	return nil
}

func (f *fakeUserRepository) GetUser(ctx context.Context, user *model.User) error {
	return nil
}

func (f *fakeUserRepository) RemoveUserByUsername(username string) error {
	return nil
}

func (f *fakeUserRepository) UpdateUser(user *model.User) (*model.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) Exists(ctx context.Context, username string) bool {
	return false
}

func (f *fakeUserRepository) LogoutUser(ctx context.Context, user *model.User) error {
	return nil
}
