package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/model"
	"github.com/dgyurics/auth/auth-server/repository"
	"github.com/dgyurics/auth/auth-server/service"
	"github.com/gorilla/websocket"
)

// TODO Prevent user from creating too many sessions

// RequestHandler contains necessary dependents to handle HTTP requests.
type RequestHandler struct {
	sessionConfig   config.Session
	authService     service.AuthService
	sessionService  service.SessionService
	userRepository  repository.UserRepository
	eventRepository repository.EventRepository
	upgrader        websocket.Upgrader
}

// NewHTTPHandler returns an instance of HTTPHandler
func NewHTTPHandler(config config.Config) *RequestHandler {
	// create SQL client
	sqlClient := repository.NewDBClient()
	sqlClient.Connect(config.PostgreSQL)

	// create session service
	redisClient := cache.NewClient(config.Redis)
	sessionCache := cache.NewSessionCache(redisClient)
	sessionService := service.NewSessionService(sessionCache)

	// create auth service
	userRepo := repository.NewUserRepository(sqlClient)
	eventRepo := repository.NewEventRepository(sqlClient)
	authService := service.NewAuthService(userRepo, eventRepo)

	// create websocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// FIXME limit to same origin or config.Cors.AllowOrigin
			return true
		},
	}

	// create HTTPHandler
	sessionConfig := config.Session
	return &RequestHandler{
		sessionConfig,
		authService,
		sessionService,
		userRepo,
		eventRepo,
		upgrader,
	}
}

func (s *RequestHandler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *RequestHandler) registration(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var user *model.User
	if err := parseRequestBody(r, &user); err != nil {
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

func (s *RequestHandler) login(w http.ResponseWriter, r *http.Request) {
	// TODO return existing session if exists

	// Parse request body
	var user *model.User
	if err := parseRequestBody(r, &user); err != nil {
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

func (s *RequestHandler) logout(w http.ResponseWriter, r *http.Request) {
	// Return error if user has no session
	cookie, err := s.extractSession(r)
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

func (s *RequestHandler) logoutAll(w http.ResponseWriter, r *http.Request) {
	// Return error if user has no session
	cookie, err := s.extractSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusBadRequest)
		return
	}

	// Generate logout all event (requires userID)
	if err := s.logoutUsers(r.Context(), cookie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO Invalidate all sessions
	if err := s.invalidateSessions(r.Context(), w, cookie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *RequestHandler) user(w http.ResponseWriter, r *http.Request) {
	cookie, err := s.extractSession(r)
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

func (s *RequestHandler) sessions(w http.ResponseWriter, r *http.Request) {
	cookie, err := s.extractSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return
	}

	// verify session valid
	sessionIDs, err := s.sessionService.FetchAll(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("invalid session: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := json.NewEncoder(w).Encode(sessionIDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// get user info/id
// when active sessions changes, send updated list to client
func (s *RequestHandler) websocket(w http.ResponseWriter, r *http.Request) {
	// verify session valid
	cookie, err := s.extractSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return
	}
	if _, err := s.sessionService.Fetch(r.Context(), cookie.Value); err != nil {
		log.Printf("invalid session: %s", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// upgrade connection to websocket
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("failed to close websocket connection: %s", err)
		}
	}()

	// initialize a variable to store last delivered payload
	var lastSessionVersion string

	for {
		// FIXME Handle disconnect. Currently blocks, so not working...
		// go func() {
		// 	c.ReadMessage()
		// 	fmt.Println("received message from client")
		// }()

		// fetch all sessions for user
		sessions, err := s.sessionService.FetchAll(r.Context(), cookie.Value)
		if err != nil {
			log.Printf("failed to fetch sessions: %s", err)
			if err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "failed to fetch sessions")); err != nil {
				log.Printf("failed to write close message: %s", err)
			}
			break
		}

		// serialize the array to JSON
		jsonData, err := json.Marshal(sessions)
		if err != nil {
			log.Printf("failed to marshal sessions: %s", err)
			if err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "failed to marshal sessions")); err != nil {
				log.Printf("failed to write close message: %s", err)
			}
			return
		}

		// when data changes, send new data to client
		if lastSessionVersion != string(jsonData) {
			if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Printf("failed to write message: %s", err)
			}

			// Update the last session version
			lastSessionVersion = string(jsonData)
		}

		time.Sleep(5 * time.Second)
	}
}

func parseRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *RequestHandler) extractSession(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(s.sessionConfig.Name)
}

func (s *RequestHandler) logoutUser(ctx context.Context, cookie *http.Cookie) error {
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
	return s.authService.Logout(ctx, user) // TODO include sessionID in event body
}

func (s *RequestHandler) logoutUsers(ctx context.Context, cookie *http.Cookie) error {
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
	return s.authService.LogoutAll(ctx, user) // TODO include sessionIDs in event body
}

func (s *RequestHandler) invalidateSessions(ctx context.Context, w http.ResponseWriter, cookie *http.Cookie) error {
	cookie, err := s.sessionService.RemoveAll(ctx, cookie)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s *RequestHandler) invalidateSession(ctx context.Context, w http.ResponseWriter, cookie *http.Cookie) error {
	cookie, err := s.sessionService.Remove(ctx, cookie)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s *RequestHandler) createSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	cookie, err := s.sessionService.Create(ctx, user.ID)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s *RequestHandler) close() model.Errors {
	errors := make(model.Errors, 0)
	errors = append(errors, s.userRepository.Close())
	errors = append(errors, s.eventRepository.Close())
	return errors
}

// TODO return model.Errors instead of error
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
