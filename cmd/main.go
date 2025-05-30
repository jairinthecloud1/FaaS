package main

import (
	"faas-api/platform/authenticator"
	"faas-api/platform/router"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(auth)

	log.Print("Server listening on http://localhost:9080/")
	if err := http.ListenAndServe("0.0.0.0:9080", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
