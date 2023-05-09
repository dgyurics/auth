package service

import (
	"context"
	"testing"

	"github.com/dgyurics/auth/auth-server/model"
	repo "github.com/dgyurics/auth/auth-server/repository"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceSuite(t *testing.T) {
	suite := &AuthServiceTestSuite{}
	suite.Setup()

	t.Run("TestCreate", suite.TestCreate)
	t.Run("TestCreateUserAlreadyExists", suite.TestCreateUserAlreadyExists)
	t.Run("Login", suite.TestLogin)
	t.Run("LoginUserNotExist", suite.TestLoginUserNotExist)
	t.Run("Logout", suite.TestLogout)
}

type AuthServiceTestSuite struct {
	userRepo  repo.UserRepository
	eventRepo repo.EventRepository
	service   AuthService
}

func (suite *AuthServiceTestSuite) Setup() {
	suite.userRepo = &repo.MockUserRepository{
		Users: []*model.User{},
	}
	suite.eventRepo = &repo.MockEventRepository{
		Events: []*model.Event{},
	}
	suite.service = NewAuthService(suite.userRepo, suite.eventRepo)
}

func (suite *AuthServiceTestSuite) TestCreate(t *testing.T) {
	user := model.User{
		Username: repo.GenerateUniqueUsername(),
		Password: "test",
	}
	err := suite.service.Create(context.Background(), &user)
	require.NoError(t, err)
	// verify user assigned id
	require.NotEmpty(t, user.ID)
	// verify user password was hashed
	require.NotEqual(t, user.Password, "test")
	// verify user.Id is not default uuid
	require.NotEqual(t, user.ID, "00000000-0000-0000-0000-000000000000")
	// using authService.Exists verify user was created
	require.True(t, suite.service.Exists(context.Background(), &user))
}

func (suite *AuthServiceTestSuite) TestCreateUserAlreadyExists(t *testing.T) {
	username := repo.GenerateUniqueUsername()
	password := "pw123"
	err := suite.service.Create(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	// create user with same username
	// should return error
	err = suite.service.Create(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.Error(t, err)
}

func (suite *AuthServiceTestSuite) TestLogin(t *testing.T) {
	username := repo.GenerateUniqueUsername()
	password := "pw1234"
	err := suite.service.Create(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	err = suite.service.Login(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)
}

func (suite *AuthServiceTestSuite) TestLoginUserNotExist(t *testing.T) {
	username := repo.GenerateUniqueUsername()
	password := "pw123"
	err := suite.service.Login(context.Background(), &model.User{
		Username: username,
		Password: password,
	})
	require.Error(t, err)
}

func (suite *AuthServiceTestSuite) TestLogout(t *testing.T) {
	user := model.User{
		Username: repo.GenerateUniqueUsername(),
		Password: "test",
	}
	err := suite.service.Create(context.Background(), &user)
	require.NoError(t, err)

	err = suite.service.Logout(context.Background(), &user)
	require.NoError(t, err)
}
