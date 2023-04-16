package service

import (
	"auth/src/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	service := NewAuthService(&fakeUserRepository{
		users: []*model.User{},
	})

	user := model.User{
		Username: "test",
		Password: "test",
	}
	err := service.Create(context.Background(), &user)
	require.NoError(t, err)
	// verify user assigned id
	require.NotEmpty(t, user.Id)
	// verify user password was hashed
	require.NotEqual(t, user.Password, "test")
	// verify user.Id is not default uuid
	require.NotEqual(t, user.Id, "00000000-0000-0000-0000-000000000000")
	// using authService.Exists verify user was created
	require.True(t, service.Exists(context.Background(), &user))
}

func TestLogin(t *testing.T) {
	// TODO
}

func TestLogout(t *testing.T) {
	// TODO
}

// TODO move fake repository to own separate file

type fakeUserRepository struct {
	users []*model.User
}

func (f *fakeUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	f.users = append(f.users, user)
	return nil
}

func (f *fakeUserRepository) LoginSuccess(ctx context.Context, user *model.User) error {
	return nil
}

func (f *fakeUserRepository) GetUser(ctx context.Context, user *model.User) error {
	for _, u := range f.users {
		if u.Username == user.Username {
			user.Id = u.Id
			user.Password = u.Password
			return nil
		}
		if u.Id == user.Id {
			user.Username = u.Username
			user.Password = u.Password
			return nil
		}
	}
	return errors.New("user not found")
}

func (f *fakeUserRepository) Exists(ctx context.Context, username string) bool {
	for _, u := range f.users {
		if u.Username == username {
			return true
		}
	}
	return false
}

func (f *fakeUserRepository) LogoutUser(ctx context.Context, user *model.User) error {
	return nil
}
