package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	Port = ":8080"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("echo"))
	})
	log.Println("Secure service listening on port " + Port)
	http.ListenAndServe(Port, r)
}
