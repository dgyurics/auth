package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/model"
	repo "github.com/dgyurics/auth/auth-server/repository"
	"github.com/dgyurics/auth/auth-server/service"
	"github.com/stretchr/testify/require"
)

var env = config.New()

func TestHandlerSuite(t *testing.T) {
	suite := &HandlerTestSuite{}
	suite.Setup()

	t.Run("TestHealthCheck", suite.TestHealthCheck)
	t.Run("TestLogin", suite.TestRegistration)
	t.Run("TestLogin", suite.TestLogin)
	// t.Run("TestLogout", suite.TestLogout)
}

type HandlerTestSuite struct {
	userRepo       repo.UserRepository
	eventRepo      repo.EventRepository
	authService    service.AuthService
	sessionCache   cache.SessionCache
	sessionService service.SessionService
	handler        RequestHandler
}

func (suite *HandlerTestSuite) Setup() {
	suite.userRepo = &repo.MockUserRepository{
		Users: []*model.User{},
	}
	suite.eventRepo = &repo.MockEventRepository{
		Events: []*model.Event{},
	}
	suite.authService = service.NewAuthService(suite.userRepo, suite.eventRepo)
	suite.sessionCache = &cache.MockSessionCache{
		Sessions: make(map[string]string),
	}
	suite.sessionService = service.NewSessionService(suite.sessionCache)
	suite.handler = &HTTPHandler{
		authService:    suite.authService,
		sessionService: suite.sessionService,
	}
}

func (suite *HandlerTestSuite) TestHealthCheck(t *testing.T) {
	// Create a mock request with a GET method and nil body
	req := httptest.NewRequest("GET", "/health", nil)

	// Create a mock response recorder
	rr := httptest.NewRecorder()
	suite.handler.healthCheck(rr, req)

	// Check the status code of the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func (suite *HandlerTestSuite) TestRegistration(t *testing.T) {
	// Generate unique user
	_, userIO := generateUniqueUser(t)

	// Create a mock request with a POST method and user body
	req := httptest.NewRequest(http.MethodPost, "/registration", userIO)
	rr := httptest.NewRecorder()
	suite.handler.registration(rr, req)

	// Check status code of response
	require.Equal(t, http.StatusCreated, rr.Code)

	// Check response header for session cookie
	cookie := rr.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookie)
	verifycookie(t, cookie, false)
}

func (suite *HandlerTestSuite) TestLogin(t *testing.T) {
	// Generate unique user
	user, userIO := generateUniqueUser(t)

	// Create new user in database
	err := suite.authService.Create(context.Background(), user)
	require.NoError(t, err)

	// Create a mock request with a POST method and user body
	req := httptest.NewRequest(http.MethodPost, "/login", userIO)
	rr := httptest.NewRecorder()
	suite.handler.login(rr, req)

	// Check status code of response
	require.Equal(t, http.StatusOK, rr.Code)

	// Check response header for session cookie
	cookie := rr.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookie)
	verifycookie(t, cookie, false)
}

func (suite *HandlerTestSuite) TestLogout(t *testing.T) {
	// Generate unique user
	user, userIO := generateUniqueUser(t)

	// Create new user in database
	err := suite.authService.Create(context.Background(), user)
	require.NoError(t, err)
	// Create a session for the user

	// Create a mock request with a POST method and user body
	req := httptest.NewRequest(http.MethodPost, "/logout", userIO)
	// Add session cookie to request
	rr := httptest.NewRecorder()
	suite.handler.logout(rr, req)

	// Check status code of response
	require.Equal(t, http.StatusOK, rr.Code)

	// Check response header for session cookie
	cookie := rr.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookie)
	verifycookie(t, cookie, true)
}

func generateUniqueUser(t *testing.T) (*model.User, io.Reader) {
	user := model.User{
		Username: repo.GenerateUniqueUsername(),
		Password: "test",
	}

	// Encode the struct as a JSON string
	jsonUser, err := json.Marshal(user)
	require.NoError(t, err)

	// Convert the JSON string to an io.Reader
	return &user, bytes.NewReader(jsonUser)
}

// Verifycookie verifies the session cookie has the correct attributes and values
func verifycookie(t *testing.T, cookieStr string, expired bool) {
	name := fmt.Sprintf("%s=", env.Session.Name)
	require.Contains(t, cookieStr, name)

	domain := fmt.Sprintf("Domain=%s", env.Session.Domain)
	require.Contains(t, cookieStr, domain)

	path := fmt.Sprintf("Path=%s", env.Session.Path)
	require.Contains(t, cookieStr, path)

	if expired {
		maxAge := "Max-Age=0"
		require.Contains(t, cookieStr, maxAge)
	} else {
		maxAge := fmt.Sprintf("Max-Age=%d", env.Session.MaxAge)
		require.Contains(t, cookieStr, maxAge)
	}

	sameSite := fmt.Sprintf("SameSite=%s", env.Session.SameSite)
	require.Contains(t, cookieStr, sameSite)

	secure := "Secure"
	if env.Session.Secure {
		require.Contains(t, cookieStr, secure)
	} else {
		require.NotContains(t, cookieStr, secure)
	}

	httpOnly := "HttpOnly"
	if env.Session.HTTPOnly {
		require.Contains(t, cookieStr, httpOnly)
	} else {
		require.NotContains(t, cookieStr, httpOnly)
	}
}
