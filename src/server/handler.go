package server

import (
	"auth/src/cache"
	"auth/src/config"
	"auth/src/model"
	"auth/src/repository"
	"auth/src/service"
	"encoding/json"
	"log"
	"net/http"
)

// create a new AuthService

type httpHandler struct {
	authService    service.AuthService
	sessionService service.SessionService
}

func NewHttpHandler() *httpHandler {
	config := config.New()
	dbClient := repository.NewDBClient()
	dbClient.Connect(config.PostgreSql)
	redisClient := cache.NewClient(config.Redis)
	return &httpHandler{
		authService:    service.NewAuthService(dbClient),
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

	// ensure username and password are not empty
	// TODO verify username is alphanumeric: ref https://stackoverflow.com/a/38554480/714618
	// Strings are UTF-8 encoded, this means each charcter aka rune can be of 1 to 4 bytes long
	password := []byte(user.Password)
	if user.Username == "" || len(user.Username) > 50 || len(password) < 1 || len(password) > 72 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ensure username is unique
	ctx := r.Context()
	if s.authService.Exists(ctx, user.Username) {
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}
	// create user
	user, err := s.authService.Create(ctx, user.Username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionId := s.sessionService.Create(ctx, user.Id.String())
	if sessionId == "" {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, createCookie("X-Session-ID", sessionId))
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

	// ensure username and password are not empty
	// TODO verify username is alphanumeric: ref https://stackoverflow.com/a/38554480/714618
	if user.Username == "" || len(user.Username) > 50 || len(user.Password) < 1 || len(user.Password) > 72 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userVrfd, err := s.authService.Login(ctx, user.Username, user.Password)
	if err != nil {
		log.Default().Printf("login failed: username: %s, err: %s", user.Username, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionId := s.sessionService.Create(ctx, userVrfd.Id.String())
	if sessionId == "" {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, createCookie("X-Session-ID", sessionId))
	w.WriteHeader(http.StatusOK)
}

func (s *httpHandler) logout(w http.ResponseWriter, r *http.Request) {
	// from request cookie session
	// session := r.cookie.session
	// or from request url param token

	// invalidate session
	w.WriteHeader(http.StatusOK)
}

func (s *httpHandler) session(w http.ResponseWriter, r *http.Request) {
	// from request cookie session
	// session := r.cookie.session
	// or from request url param token

	// ensure session is a valid 128+ bits long
	// https://owasp.org/www-community/attacks/Session_hijacking_attack
	w.WriteHeader(http.StatusOK)
}

// FIXME make configurable
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
