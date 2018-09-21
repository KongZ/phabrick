package handlers

import (
	"github.com/KongZ/phabrick/internal/config"
	"github.com/gorilla/mux"
)

// Router register necessary routes and returns an instance of a router.
func Router(release string, config *config.Config) *mux.Router {
	webhook := &Webhook{Config: config}
	r := mux.NewRouter()
	r.HandleFunc("/version", version(release))
	r.HandleFunc("/webhook", webhook.receiveNotify).Methods("POST")
	return r
}
