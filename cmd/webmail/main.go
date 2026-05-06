package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"proidentity-mail/internal/app"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/session"
	"proidentity-mail/internal/webmail"
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
	store := webmail.CompositeStore{
		Auth:    webmail.NewSQLAuthStore(conn),
		Mailbox: webmail.MaildirStore{Root: "/var/vmail"},
		Sender:  webmail.SMTPSender{Addr: "127.0.0.1:25"},
		Learner: webmail.RspamdLearner{},
	}
	server := http.Server{
		Addr:              cfg.WebmailAddr,
		Handler:           webmail.NewRouter(store, session.NewManager(session.Options{CookieName: "proidentity_webmail_session", Secure: cfg.SecureCookies})),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("webmail listening on %s", cfg.WebmailAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("webmail stopped: %v", err)
	}
}
