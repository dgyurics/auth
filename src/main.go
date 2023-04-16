package main

import (
	"log"

	"github.com/dgyurics/auth/src/config"
	"github.com/dgyurics/auth/src/server"
)

func main() {
	config := config.New()
	log.Default().Println("Auth service listening on port " + config.ServerConfig.Port)
	srv := server.NewHttpServer(":" + config.ServerConfig.Port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
