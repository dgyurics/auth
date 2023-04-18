package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dgyurics/auth/src/model"
	"github.com/stretchr/testify/require"
)

// FIXME this can be refactored a lot...

func cleanup(repo *fakeUserRepository) {
	repo.reset()
}

func TestCreate(t *testing.T) {
	repository := &fakeUserRepository{
		users: []*model.User{},
	}
	service := NewAuthService(repository)
	defer cleanup(repository)

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

func TestCreateUserAlreadyExists(t *testing.T) {
	repository := &fakeUserRepository{
		users: []*model.User{},
	}
	service := NewAuthService(repository)
	defer cleanup(repository)

	username := "test"
	password := "test"
	err := service.Create(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	// create user with same username
	// should return error
	err = service.Create(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.Error(t, err)
}

func TestLogin(t *testing.T) {
	repository := &fakeUserRepository{
		users: []*model.User{},
	}
	service := NewAuthService(repository)
	defer cleanup(repository)

	username := "test"
	password := "test"
	err := service.Create(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	err = service.Login(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)
}

func TestLoginUserNotExist(t *testing.T) {
	repository := &fakeUserRepository{
		users: []*model.User{},
	}
	service := NewAuthService(repository)
	defer cleanup(repository)

	username := "test"
	password := "test"
	err := service.Login(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.Error(t, err)
}

func TestLogout(t *testing.T) {
	repository := &fakeUserRepository{
		users: []*model.User{},
	}
	service := NewAuthService(repository)
	defer cleanup(repository)

	user := model.User{
		Username: "test",
		Password: "test",
	}
	err := service.Create(context.Background(), &user)
	require.NoError(t, err)

	err = service.Logout(context.Background(), &user)
	require.NoError(t, err)
}

// TODO move fake repository to own separate file

type fakeUserRepository struct {
	users []*model.User
}

func (f *fakeUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := f.GetUser(ctx, user)
	if err == nil {
		return errors.New("user already exists")
	}
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
	if !f.Exists(ctx, user.Username) {
		return errors.New("user not found")
	}
	return nil
}

func (f *fakeUserRepository) reset() {
	f.users = nil
}
