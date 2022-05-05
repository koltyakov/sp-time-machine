package main

import (
	"net/http"
	"os"

	"github.com/koltyakov/sp-time-machine/handlers"
	log "github.com/sirupsen/logrus"
)

func main() {
	port, exists := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if !exists {
		port = "8080"
	}

	h := handlers.NewHandlers()
	mux := http.NewServeMux()

	// Timer job(s)
	mux.HandleFunc("/timer", h.Sync)

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
