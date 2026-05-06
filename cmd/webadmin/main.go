package main

import (
	"context"
	"log"

	"proidentity-mail/internal/admin"
	"proidentity-mail/internal/app"
	"proidentity-mail/internal/db"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	var store admin.Store
	if cfg.DBDSN != "" {
		conn, err := db.Open(context.Background(), cfg.DBDSN)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		defer conn.Close()
		sqlStore := admin.NewSQLStore(conn)
		store = sqlStore
	}
	server := app.NewHTTPServer(cfg, store, admin.AuthConfig{Username: cfg.AdminUsername, Password: cfg.AdminPassword})
	log.Printf("webadmin listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("webadmin stopped: %v", err)
	}
}
