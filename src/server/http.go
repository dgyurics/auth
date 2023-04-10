package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                  // FIXME make configurable
		w.Header().Set("Allow-Control-Allow-Methods", "GET, POST, OPTIONS") // FIXME make configurable
		w.Header().Set("Allow-Control-Allow-Headers", "*")                  // FIXME make configurable
		w.Header().Set("Allow-Control-Allow-Credentials", "true")           // FIXME make configurable
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

	handler := NewHttpHandler()

	r.Get("/health", handler.healthCheck)
	r.Get("/user", handler.session)
	r.Post("/login", handler.login)
	r.Post("/logout", handler.logout)
	r.Post("/register", handler.registration)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
