package server

import (
	"net/http"
	"time"

	"github.com/dgyurics/auth/auth-server/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var cfg *config.Config

func init() {
	cfg = config.New()
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", cfg.Cors.AllowOrigin)
		w.Header().Set("Allow-Control-Allow-Methods", cfg.Cors.AllowMethods)
		w.Header().Set("Allow-Control-Allow-Headers", cfg.Cors.AllowHeaders)
		w.Header().Set("Allow-Control-Allow-Credentials", cfg.Cors.AllowCredentials)
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func initMiddleware(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(cors)
	r.Use(middleware.Timeout(time.Duration(cfg.RequestTimeout) * time.Second))
}

// NewHTTPServer returns a new http server
func NewHTTPServer(addr string) *http.Server {
	r := chi.NewRouter()
	initMiddleware(r)

	handler := NewHTTPHandler()
	setupRoutes(r, handler)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func setupRoutes(r chi.Router, h *HTTPHandler) {
	r.Get("/health", h.healthCheck)
	r.Get("/user", h.user)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	r.Post("/register", h.registration)
}
