package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/KongZ/phabrick/internal/config"
	"github.com/KongZ/phabrick/internal/handlers"
	log "github.com/sirupsen/logrus"
)

// Release a server version
const Release = "0.1"

func init() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to debug
	if ok {
		// parse string, this is built-in feature of logrus
		ll, err := log.ParseLevel(lvl)
		if err != nil {
			ll = log.InfoLevel
		}
		// set global log level
		log.SetLevel(ll)
	}
}

// How to try it: PORT=8000 go run main.go
func main() {
	log.Printf("Starting the server version %s ...", Release)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port is not set.")
	}

	r := handlers.Router(Release, config.GetConfig())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	log.Printf("The server is ready on port %s", port)

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The server is shutting down...")
	server.Shutdown(context.Background())
	log.Print("Done")
}
