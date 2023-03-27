package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func initMiddleware(r *chi.Mux) {
	r.Use(middleware.Logger)
}

func NewHttpServer(addr string) *http.Server {
	r := chi.NewRouter()
	initMiddleware(r)

	handler := NewHttpHandler()
	r.Get("/health", handler.healthCheck)
	r.Get("/session", handler.session)
	r.Post("/login", handler.authentication)
	r.Post("/register", handler.registration)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
