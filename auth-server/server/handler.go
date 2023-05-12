package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/model"
	"github.com/dgyurics/auth/auth-server/repository"
	"github.com/dgyurics/auth/auth-server/service"
)

// TODO Create /logout-all which invalidates all sessions for the user
//
// TODO Create /authorized endpoint which returns 200 OK or 401 Unauthorized
// Will be used by api-gateway to verify user is authenticated.
// Currently api-gateway is using /user endpoint which is not performant.
//
// TODO Prevent user from creating too many sessions

// RequestHandler defines the methods necessary to handle HTTP requests.
type RequestHandler interface {
	healthCheck(w http.ResponseWriter, r *http.Request)
	registration(w http.ResponseWriter, r *http.Request)
	login(w http.ResponseWriter, r *http.Request)
	logout(w http.ResponseWriter, r *http.Request)
	user(w http.ResponseWriter, r *http.Request)
}

// HTTPHandler contains necessary dependents to handle HTTP requests.
type HTTPHandler struct {
	sessionConfig  config.Session
	authService    service.AuthService
	sessionService service.SessionService
}

// NewHTTPHandler returns an instance of HTTPHandler
func NewHTTPHandler(config config.Config) RequestHandler {
	// create session service
	redisClient := cache.NewClient(config.Redis)
	sessionCache := cache.NewSessionCache(redisClient)
	sessionService := service.NewSessionService(sessionCache)

	// create auth service
	sqlClient := repository.NewDBClient()
	sqlClient.Connect(config.PostgreSQL)
	userRepo := repository.NewUserRepository(sqlClient)
	eventRepo := repository.NewEventRepository(sqlClient)
	authService := service.NewAuthService(userRepo, eventRepo)

	// create HTTPHandler
	sessionConfig := config.Session
	return &HTTPHandler{
		sessionConfig,
		authService,
		sessionService,
	}
}

func (s *HTTPHandler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *HTTPHandler) registration(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var user *model.User
	if err := s.parseRequestBody(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := validateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify username unique
	if s.authService.Exists(r.Context(), user) {
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}

	// Create user
	if err := s.authService.Create(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session
	if err := s.createSession(r.Context(), w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *HTTPHandler) login(w http.ResponseWriter, r *http.Request) {
	// Return bad request if user has valid session cookie
	cookie, err := s.getSession(r)
	if err == nil && cookie.Value != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse request body
	var user *model.User
	if err := s.parseRequestBody(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := validateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Authenticate user
	if err := s.authService.Authenticate(r.Context(), user); err != nil {
		log.Printf("login failed: username: %s, err: %s", user.Username, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session
	if err := s.createSession(r.Context(), w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HTTPHandler) logout(w http.ResponseWriter, r *http.Request) {
	// Return error if user has no session
	cookie, err := s.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusBadRequest)
		return
	}

	// Generate logout event (requires userID)
	if err := s.logoutUser(r.Context(), cookie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Invalidate session
	if err := s.invalidateSession(r.Context(), w, cookie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HTTPHandler) user(w http.ResponseWriter, r *http.Request) {
	cookie, err := s.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return
	}

	// verify session valid
	userID, err := s.sessionService.Fetch(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("invalid session: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// extend session in cache and update cookie max age
	cookie, err = s.sessionService.Extend(r.Context(), userID.String(), cookie)
	if err != nil {
		log.Printf("failed to extend session: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, cookie)

	// fetch user from database
	user := &model.User{ID: userID}
	if err = s.authService.Fetch(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// encode user as json and write to response
	if err := json.NewEncoder(w).Encode(model.OmitPassword(user)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HTTPHandler) parseRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *HTTPHandler) getSession(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(s.sessionConfig.Name)
}

func (s *HTTPHandler) logoutUser(ctx context.Context, cookie *http.Cookie) error {
	// fetch session from cache
	userID, err := s.sessionService.Fetch(ctx, cookie.Value)
	if err != nil {
		return err
	}
	// fetch user from database
	user := &model.User{ID: userID}
	if err := s.authService.Fetch(ctx, user); err != nil {
		return err
	}
	// generate logout event
	return s.authService.Logout(ctx, user)
}

func (s *HTTPHandler) invalidateSession(ctx context.Context, w http.ResponseWriter, cookie *http.Cookie) error {
	cookie, err := s.sessionService.Remove(ctx, cookie)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s *HTTPHandler) createSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	cookie, err := s.sessionService.Create(ctx, user.ID.String())
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func validateUser(user *model.User) error {
	if user.Username == "" {
		return errors.New("username cannot be empty")
	}
	// Strings are UTF-8 encoded, this means each charcter aka rune can be 1 to 4 bytes
	if len(user.Username) > 50 {
		return errors.New("username cannot exceed 50 characters")
	}
	if len(user.Password) < 1 || len(user.Password) > 72 {
		return errors.New("password must be between 1 and 72 characters")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(user.Username) {
		return errors.New("username must be alphanumeric")
	}
	return nil
}
