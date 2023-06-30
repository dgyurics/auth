package server

import (
	"context"
	"net/http"
	"time"

	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/model"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var cfg config.Config

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

// HTTPServer is a wrapper around http.Server
// additionally exposing *RequestHandler for closing resources
type HTTPServer struct {
	http.Server
	handler *RequestHandler
}

// Close closes the server and all resources
func (s *HTTPServer) Close(ctx context.Context) model.Errors {
	errors := make(model.Errors, 0)
	errors = append(errors, s.Shutdown(ctx))
	errors = append(errors, s.handler.close())
	return errors
}

// NewHTTPServer returns a new http server
func NewHTTPServer(addr string) *HTTPServer {
	cfg = config.New()
	r := chi.NewRouter()
	initMiddleware(r)

	handler := NewHTTPHandler(config.New())
	setupRoutes(r, *handler)

	return &HTTPServer{
		Server: http.Server{
			Addr:    addr,
			Handler: r,
		},
		handler: handler,
	}
}

func setupRoutes(r chi.Router, h RequestHandler) {
	r.Get("/health", h.healthCheck)
	r.Get("/user", h.user)
	r.Get("/sessions", h.sessions)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	r.Post("/logout-all", h.logoutAll)
	r.Post("/register", h.registration)
}
