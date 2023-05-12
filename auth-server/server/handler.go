package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/model"
	"github.com/dgyurics/auth/auth-server/repository"
	"github.com/dgyurics/auth/auth-server/service"
	"github.com/google/uuid"
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
	// Parse user from request body
	var user *model.User
	if err := s.parseRequestBody(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := s.authService.ValidateUserInput(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify username is unique
	if s.authService.Exists(r.Context(), user) {
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}

	// Create user and session
	if err := s.createUserAndSession(r.Context(), w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *HTTPHandler) login(w http.ResponseWriter, r *http.Request) {
	// Return bad request if user already has active session
	cookie, err := s.getSession(r)
	if err == nil && cookie.Value != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse user from request body
	var user *model.User
	if err := s.parseRequestBody(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := s.authService.ValidateUserInput(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Login user and create session
	if err := s.authenticateUserAndCreateSession(r.Context(), w, user); err != nil {
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
	userID, err := s.sessionService.Fetch(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.logoutUserAndInvalidateSession(r.Context(), w, userID, cookie); err != nil {
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

	// verify session is valid
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

	// fetch user from database
	user := &model.User{ID: userID}
	if err = s.authService.Fetch(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(model.OmitPassword(user)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (s *HTTPHandler) parseRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *HTTPHandler) getSession(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(s.sessionConfig.Name)
}

func (s *HTTPHandler) logoutUserAndInvalidateSession(ctx context.Context, w http.ResponseWriter, userID uuid.UUID, cookie *http.Cookie) error {
	// Generate logout event
	user := &model.User{ID: userID}
	if err := s.authService.Fetch(ctx, user); err != nil {
		return err
	}
	if err := s.authService.Logout(ctx, user); err != nil {
		return err
	}

	// Invalidate session
	cookie, err := s.sessionService.Remove(ctx, cookie)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s *HTTPHandler) authenticateUserAndCreateSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	// login user
	if err := s.authService.Login(ctx, user); err != nil {
		log.Printf("login failed: username: %s, err: %s", user.Username, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return err
	}

	// create session
	cookie, err := s.sessionService.Create(ctx, user.ID.String())
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return err
	}
	// set session cookie
	http.SetCookie(w, cookie)

	return nil
}

func (s *HTTPHandler) createUserAndSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	if err := s.authService.Create(ctx, user); err != nil {
		return err
	}
	cookie, err := s.sessionService.Create(ctx, user.ID.String())
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}
