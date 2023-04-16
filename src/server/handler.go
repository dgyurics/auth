package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgyurics/auth/src/cache"
	"github.com/dgyurics/auth/src/config"
	"github.com/dgyurics/auth/src/model"
	"github.com/dgyurics/auth/src/repository"
	"github.com/dgyurics/auth/src/service"
)

const SessionCookieName = "X-Session-ID"

type httpHandler struct {
	authService    service.AuthService
	sessionService service.SessionService
}

func NewHttpHandler() *httpHandler {
	config := config.New()
	redisClient := cache.NewClient(config.Redis)
	return &httpHandler{
		authService:    service.NewAuthService(repository.NewUserRepository()),
		sessionService: service.NewSessionService(redisClient),
	}
}

func (s *httpHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *httpHandler) registration(w http.ResponseWriter, r *http.Request) {
	// unmarshal request body
	var user *model.User
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO verify username is alphanumeric: ref https://stackoverflow.com/a/38554480/714618
	// Strings are UTF-8 encoded, this means each charcter aka rune can be of 1 to 4 bytes long
	if user.Username == "" || len(user.Username) > 50 || len(user.Password) < 1 || len(user.Password) > 72 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ensure username is unique
	ctx := r.Context()
	if s.authService.Exists(ctx, user) {
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}
	// create user
	if err := s.authService.Create(ctx, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionId, err := s.sessionService.Create(ctx, user.Id.String())
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, createCookie(SessionCookieName, sessionId))
	w.WriteHeader(http.StatusCreated)
}

func (s *httpHandler) login(w http.ResponseWriter, r *http.Request) {
	// unmarshal request body
	var user *model.User
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO verify username is alphanumeric: ref https://stackoverflow.com/a/38554480/714618
	if user.Username == "" || len(user.Username) > 50 || len(user.Password) < 1 || len(user.Password) > 72 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.authService.Login(ctx, user); err != nil {
		log.Default().Printf("login failed: username: %s, err: %s", user.Username, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionId, err := s.sessionService.Create(ctx, user.Id.String())
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, createCookie(SessionCookieName, sessionId))
	w.WriteHeader(http.StatusOK)
}

// FIXME should invalidate ALL user sessions,
// currently only invalidates the session cookie in the request
func (s *httpHandler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.SetCookie(w, expireCookie(cookie))

	// generate logout event
	userId, err := s.sessionService.Fetch(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = s.authService.Logout(r.Context(), &model.User{Id: userId}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// remove session from redis
	s.sessionService.Remove(r.Context(), cookie.Value)

	w.WriteHeader(http.StatusOK)
}

// secure endpoint which retrieves user information
func (s *httpHandler) user(w http.ResponseWriter, r *http.Request) {
	// ensure session is a valid 128+ bits long
	// https://owasp.org/www-community/attacks/Session_hijacking_attack

	// extract session from cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// verify session is valid and fetch user id
	userId, err := s.sessionService.Fetch(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user := &model.User{Id: userId}
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

// FIXME make configurable
// TODO Validate contents of cookie to ensure it has not been modified/tampered with.
// This can be done by adding a message authentication code (MAC) to the cookie,
// which can be used to verify the integrity of the cookie's contents.
func createCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		MaxAge:   1800, // 30 minutes
		// Domain: "",
		// Path: "",
		// Secure: true,
		// SameSite: ,
	}
}
