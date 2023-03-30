package server

import (
	"net/http"

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
}

func NewHttpServer(addr string) *http.Server {
	r := chi.NewRouter()
	initMiddleware(r)

	handler := NewHttpHandler()

	// TODO pass context in-order to fail fast
	r.Get("/health", handler.healthCheck)
	r.Get("/session", handler.session)
	r.Post("/login", handler.authentication)
	r.Post("/register", handler.registration)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
