package main

import (
	"log"

	"proidentity-mail/internal/app"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	server := app.NewHTTPServer(cfg)
	log.Printf("webadmin listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("webadmin stopped: %v", err)
	}
}
