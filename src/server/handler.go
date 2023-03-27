package server

import (
	"net/http"
)

type httpHandler struct{}

func NewHttpHandler() *httpHandler {
	return &httpHandler{}
}

func (s *httpHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *httpHandler) registration(w http.ResponseWriter, r *http.Request) {
	// from request body
	// username := r.body.username
	// password := r.body.password

	// below handled by auth service
	// hash + salt password
	// verify username not exists
	// insert into db
	// return valid session token
	w.WriteHeader(http.StatusOK)
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
