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
		w.Header().Set("Access-Control-Allow-Credentials", cfg.Cors.AllowCredentials)
		w.Header().Set("Access-Control-Allow-Origin", cfg.Cors.AllowOrigin)
		w.Header().Set("Access-Control-Allow-Headers", cfg.Cors.AllowHeaders)
		w.Header().Set("Access-Control-Allow-Methods", cfg.Cors.AllowMethods)
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
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
	h := NewHTTPHandler(config.New())

	defaultGroup := r.Group(nil)
	defaultGroup.Use(middleware.Logger)
	defaultGroup.Use(cors)
	defaultGroup.Use(middleware.Timeout(time.Duration(cfg.RequestTimeout) * time.Second))

	defaultGroup.Get("/health", h.healthCheck)
	defaultGroup.Get("/user", h.user)
	defaultGroup.Get("/sessions", h.sessions)
	defaultGroup.Post("/login", h.login)
	defaultGroup.Post("/logout", h.logout)
	defaultGroup.Post("/logout-all", h.logoutAll)
	defaultGroup.Post("/register", h.registration)

	wsGroup := r.Group(nil)
	wsGroup.HandleFunc("/ws", h.websocket)

	return &HTTPServer{
		Server: http.Server{
			Addr:    addr,
			Handler: r,
		},
		handler: h,
	}
}
