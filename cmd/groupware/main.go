package main

import (
	"log"
	"net/http"
	"time"

	"proidentity-mail/internal/app"
	"proidentity-mail/internal/groupware"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	server := http.Server{
		Addr:              cfg.GroupwareAddr,
		Handler:           groupware.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("groupware listening on %s", cfg.GroupwareAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("groupware stopped: %v", err)
	}
}
