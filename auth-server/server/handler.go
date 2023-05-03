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

// RequestHandler is an interface that defines the methods
// that are necessary to handle HTTP requests.
type RequestHandler interface {
	healthCheck(w http.ResponseWriter, r *http.Request)
	registration(w http.ResponseWriter, r *http.Request)
	login(w http.ResponseWriter, r *http.Request)
	logout(w http.ResponseWriter, r *http.Request)
	user(w http.ResponseWriter, r *http.Request)
}

// HTTPHandler is a struct that contains all the dependencies necessary
// to handle HTTP requests.
type HTTPHandler struct {
	authService    service.AuthService
	sessionService service.SessionService
}

// NewHTTPHandler returns a new instance of httpHandler
// FIXME refactor by returning interface rather than struct
func NewHTTPHandler() RequestHandler {
	redisClient := cache.NewClient(env.Redis)
	sessionService := service.NewSessionService(redisClient)

	sqlClient := repository.NewDBClient()
	sqlClient.Connect(config.New().PostgreSQL)
	userRepo := repository.NewUserRepository(sqlClient)
	eventRepo := repository.NewEventRepository(sqlClient)
	authService := service.NewAuthService(userRepo, eventRepo)

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
	// TODO check if user is already logged in

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
	if err != nil || cookie.Value == "" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.SetCookie(w, expireCookie(cookie))

	// generate logout event
	sessionID := cookie.Value
	userID, err := s.sessionService.Fetch(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fetch additional user information
	// so user logout event.body can be populated
	user := &model.User{ID: userID}
	if err = s.authService.Fetch(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = s.authService.Logout(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// remove session from redis
	if err = s.sessionService.Remove(r.Context(), sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// secure endpoint which retrieves user information
// used by api-gateway to verify user is authenticated

// FIXME create separate endpoint for api-gateway to use, which
// does not return user information (slightly more secure & performant)
func (s *HTTPHandler) user(w http.ResponseWriter, r *http.Request) {
	// extract session from cookie
	cookie, err := r.Cookie(env.Session.Name)
	if err != nil || cookie.Value == "" {
		log.Printf("failed to fetch session cookie: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// verify session is valid and fetch user id
	sessionID := cookie.Value
	userID, err := s.sessionService.Fetch(r.Context(), sessionID)
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

	if err := json.NewEncoder(w).Encode(model.OmitPassword(user)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// extend redis session TTL
	if err = s.sessionService.Extend(r.Context(), userID.String(), sessionID, expiration(env.Session.MaxAge)); err != nil {
		log.Printf("failed to extend session: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// extend session MaxAge and Expires
	http.SetCookie(w, updateCookie(env.Session, cookie))

	w.WriteHeader(http.StatusOK)
}

func expireCookie(cookie *http.Cookie) *http.Cookie {
	if cookie == nil {
		return nil
	}
	cookie.Expires = time.Now().AddDate(0, 0, -1)
	return cookie
}

func updateCookie(session config.Session, cookie *http.Cookie) *http.Cookie {
	if cookie == nil {
		return nil
	}
	cookie.MaxAge = session.MaxAge
	cookie.Expires = expireTime(session.MaxAge)
	return cookie
}

// TODO Validate contents of cookie to ensure it has not been modified/tampered with.
// This can be done by adding a message authentication code (MAC) to the cookie,
// which can be used to verify the integrity of the cookie's contents.
func newCookie(session config.Session, value string) *http.Cookie {
	return &http.Cookie{
		Value:    value,
		Name:     session.Name,
		Domain:   session.Domain,
		Path:     session.Path,
		MaxAge:   session.MaxAge,
		Expires:  expireTime(session.MaxAge),
		Secure:   session.Secure,
		HttpOnly: session.HTTPOnly,
		SameSite: mapSameSite(session.SameSite),
	}
}

func mapSameSite(value string) http.SameSite {
	switch value {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

// expiration used for cookie session
func expireTime(maxAge int) time.Time {
	return time.Now().Add(time.Duration(maxAge) * time.Second)
}

// expiration used for redis session
func expiration(maxAge int) time.Duration {
	return time.Duration(maxAge) * time.Second
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
	sessionID, err := s.sessionService.Create(ctx, user.ID.String(), expiration(env.Session.MaxAge))
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return err
	}
	// set session cookie
	http.SetCookie(w, newCookie(env.Session, sessionID))

	return nil
}

func (s *HTTPHandler) createUserAndSession(ctx context.Context, w http.ResponseWriter, user *model.User) error {
	if err := s.authService.Create(ctx, user); err != nil {
		return err
	}
	sessionID, err := s.sessionService.Create(ctx, user.ID.String(), expiration(env.Session.MaxAge))
	if err != nil {
		return err
	}
	http.SetCookie(w, newCookie(env.Session, sessionID))
	return nil
}
