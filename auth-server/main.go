package main

import (
	"log"

	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/server"
)

func main() {
	config := config.New()
	log.Println("Auth service listening on port " + config.ServerConfig.Port)
	srv := server.NewHTTPServer(":" + config.ServerConfig.Port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
