package server

import (
	"net/http"
	"time"

	"github.com/dgyurics/auth/src/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func cors(next http.Handler) http.Handler {
	config := config.New().Cors
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", config.AllowOrigin)
		w.Header().Set("Allow-Control-Allow-Methods", config.AllowMethods)
		w.Header().Set("Allow-Control-Allow-Headers", config.AllowHeaders)
		w.Header().Set("Allow-Control-Allow-Credentials", config.AllowCredentials)
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func initMiddleware(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(cors)
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(30 * time.Second)) // FIXME make this configurable
}

func NewHttpServer(addr string) *http.Server {
	r := chi.NewRouter()
	initMiddleware(r)

	handler := NewHTTPHandler()
	setupRoutes(r, handler)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func setupRoutes(r chi.Router, h *httpHandler) {
	r.Get("/health", h.healthCheck)
	r.Get("/user", h.user)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	r.Post("/register", h.registration)
}
