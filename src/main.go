package main

import (
	"auth/src/config"
	"auth/src/server"
	"log"
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
