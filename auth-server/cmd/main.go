package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/server"
)

func main() {
	// Create new server
	config := config.New()
	server := server.NewHTTPServer(":" + config.ServerConfig.Port)

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	go func() {
		<-sig
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
		defer cancel()
		log.Println("Shutting down server...")
		err := server.Shutdown(ctx)
		if err != nil {
			log.Println(err)
		}
	}()

	// Start server
	log.Println("Auth service listening on port " + config.ServerConfig.Port)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}
}
