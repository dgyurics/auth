// package main

// import (
// 	"log"

// 	"github.com/dgyurics/auth/auth-server/config"
// 	"github.com/dgyurics/auth/auth-server/server"
// )

// func main() {
// 	config := config.New()
// 	log.Println("Auth service listening on port " + config.ServerConfig.Port)
// 	srv := server.NewHTTPServer(":" + config.ServerConfig.Port)
// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("echo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("echo"))
	})
	http.ListenAndServe(":8080", r)
}
