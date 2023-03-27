package main

import (
	"auth/src/server"
	"log"
)

func main() {
	log.Default().Println("Auth service listening on port 8080")
	srv := server.NewHttpServer(":8080")
	srv.ListenAndServe()
}
