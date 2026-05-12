package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"proidentity-mail/internal/app"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/groupware"
	"proidentity-mail/internal/session"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	conn, err := db.Open(context.Background(), cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()
	store := groupware.NewSQLStore(conn)
	limiter := session.NewSQLLoginLimiter(conn, "dav", session.Options{})
	server := http.Server{
		Addr:              cfg.GroupwareAddr,
		Handler:           groupware.NewRouterWithLimiter(store, limiter),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("groupware listening on %s", cfg.GroupwareAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("groupware stopped: %v", err)
	}
}
