package server

import (
	"auth/src/model"
	"auth/src/repository"
	"auth/src/service"
	"encoding/json"
	"net/http"
)

// create a new AuthService

type httpHandler struct {
	authService service.AuthService
}

func NewHttpHandler() *httpHandler {
	return &httpHandler{
		authService: service.NewAuthService(repository.NewDBClient()),
	}
}

func (s *httpHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *httpHandler) registration(w http.ResponseWriter, r *http.Request) {
	// unmarshal request body
	var user model.User
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ensure username and password are not empty
	if user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// ensure username is unique
	if s.authService.Exists(user.Username) {
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}
	// create user
	user, err := s.authService.Create(user.Username, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO return valid session token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *httpHandler) authentication(w http.ResponseWriter, r *http.Request) {
	// from request body
	// username := r.body.username
	// password := r.body.password

	// below handled by auth service
	// hash + salt password
	// verify username password combo exists
	// return valid session token
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
