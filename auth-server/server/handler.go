package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgyurics/auth/auth-server/cache"
	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/model"
	"github.com/dgyurics/auth/auth-server/repository"
	"github.com/dgyurics/auth/auth-server/service"
)

var env = config.New()

// HTTPHandler is a struct that contains all the dependencies necessary
// to handle HTTP requests.
type HTTPHandler struct {
	authService    service.AuthService
	sessionService service.SessionService
}

// NewHTTPHandler returns a new instance of httpHandler
// FIXME refactor by returning interface rather than struct
func NewHTTPHandler() *HTTPHandler {
	redisClient := cache.NewClient(env.Redis)
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	sessionService := service.NewSessionService(redisClient)

	return &HTTPHandler{
		authService:    authService,
		sessionService: sessionService,
	}
}

func (s *HTTPHandler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *HTTPHandler) registration(w http.ResponseWriter, r *http.Request) {
	// Parse request body into a new user object
	var user *model.User
	if err := s.parseRequestBody(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate user input
	if err := s.authService.ValidateUserInput(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ensure username is unique
	if s.authService.Exists(r.Context(), user) {
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}

	// Create user and session
	if err := s.createUserAndSession(r.Context(), w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
}

func (s *HTTPHandler) login(w http.ResponseWriter, r *http.Request) {
	// Parse request body into a new user object
	var user *model.User
	if err := s.parseRequestBody(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate user input
	if err := s.authService.ValidateUserInput(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Login user and create session
	if err := s.loginUserAndCreateSession(r.Context(), w, user); err != nil {
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
}

// FIXME should invalidate ALL user sessions,
// currently only invalidates the session cookie in the request
func (s *HTTPHandler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(env.Session.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.SetCookie(w, expireCookie(cookie))

	// generate logout event
	userID, err := s.sessionService.Fetch(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = s.authService.Logout(r.Context(), &model.User{ID: userID}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// remove session from redis
	if err = s.sessionService.Remove(r.Context(), cookie.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// secure endpoint which retrieves user information
func (s *HTTPHandler) user(w http.ResponseWriter, r *http.Request) {
	// ensure session is a valid 128+ bits long
	// https://owasp.org/www-community/attacks/Session_hijacking_attack

	// extract session from cookie
	cookie, err := r.Cookie(env.Session.Name)
	if err != nil {
		log.Printf("failed to fetch session cookie: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// verify session is valid and fetch user id
	userID, err := s.sessionService.Fetch(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("invalid session: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user := &model.User{ID: userID}
	if err = s.authService.Fetch(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(repository.OmitPassword(user)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func expireCookie(cookie *http.Cookie) *http.Cookie {
	if cookie == nil {
		return nil
	}
	cookie.Expires = time.Now().AddDate(0, 0, -1)
	return cookie
}

// TODO Validate contents of cookie to ensure it has not been modified/tampered with.
// This can be done by adding a message authentication code (MAC) to the cookie,
// which can be used to verify the integrity of the cookie's contents.
func createCookie(session config.Session, value string) *http.Cookie {
	var sameSite http.SameSite
	switch session.SameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}
	return &http.Cookie{
		Value:    value,
		Name:     session.Name,
		Domain:   session.Domain,
		Path:     session.Path,
		MaxAge:   session.MaxAge,
		Expires:  time.Now().Add(time.Duration(session.MaxAge) * time.Second),
		Secure:   session.Secure,
		HttpOnly: session.HTTPOnly,
		SameSite: sameSite,
	}
}

func (s *HTTPHandler) parseRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *HTTPHandler) loginUserAndCreateSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	// login user
	if err := s.authService.Login(ctx, user); err != nil {
		log.Printf("login failed: username: %s, err: %s", user.Username, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return err
	}

	// create session
	sessionID, err := s.sessionService.Create(ctx, user.ID.String())
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return err
	}
	// set session cookie
	http.SetCookie(w, createCookie(env.Session, sessionID))

	return nil
}

func (s *HTTPHandler) createUserAndSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	if err := s.authService.Create(ctx, user); err != nil {
		return err
	}
	sessionID, err := s.sessionService.Create(ctx, user.ID.String())
	if err != nil {
		return err
	}
	http.SetCookie(w, createCookie(env.Session, sessionID))
	return nil
}
