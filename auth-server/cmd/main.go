package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgyurics/auth/auth-server/config"
	"github.com/dgyurics/auth/auth-server/server"
)

func main() {
	// Create new server
	config := config.New()
	server := server.NewHTTPServer(":" + config.ServerConfig.Port)

	// Setup graceful shutdown
	gracefulShutdown(server)

	// Start server
	log.Println("Auth service listening on port " + config.ServerConfig.Port)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}
}

// FIXME refactor
func gracefulShutdown(server *server.HTTPServer) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
		defer cancel()
		err := server.Close(ctx)
		if err != nil {
			log.Println(err)
		}
	}()
}
