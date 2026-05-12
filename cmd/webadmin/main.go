package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"proidentity-mail/internal/admin"
	"proidentity-mail/internal/app"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/quarantine"
	"proidentity-mail/internal/session"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	var store admin.Store
	var limiter session.Limiter = session.NewLoginLimiter(session.Options{Penalties: session.AdminPenaltySchedule()})
	var authPolicyLimiter session.Limiter = session.NewLoginLimiter(session.Options{})
	var discoveryLimiter session.Limiter = session.NewLoginLimiter(session.Options{MaxFailures: 60, Lockout: 5 * time.Minute, Window: time.Minute})
	if cfg.DBDSN != "" {
		conn, err := db.Open(context.Background(), cfg.DBDSN)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		defer conn.Close()
		sqlStore := admin.NewSQLStore(conn, quarantine.FileStore{Root: cfg.QuarantineDir, MailRoot: cfg.MailRoot, DeliveryAddr: cfg.ReleaseSMTPAddr}).WithDNSSettings(admin.DNSSettings{
			MailHostname:    cfg.MailHostname,
			AdminHostname:   cfg.AdminHostname,
			WebmailHostname: cfg.WebmailHostname,
			PublicIPv4:      cfg.PublicIPv4,
			PublicIPv6:      cfg.PublicIPv6,
		})
		store = sqlStore
		limiter = session.NewSQLLoginLimiter(conn, "admin", session.Options{Penalties: session.AdminPenaltySchedule()})
		authPolicyLimiter = session.NewSQLLoginLimiter(conn, "dovecot", session.Options{})
		discoveryLimiter = session.NewSQLLoginLimiter(conn, "discovery", session.Options{MaxFailures: 60, Lockout: 5 * time.Minute, Window: time.Minute})
	}
	sessions := session.NewManager(session.Options{CookieName: "proidentity_admin_session", TTL: 15 * time.Minute, SameSite: http.SameSiteStrictMode, Secure: cfg.SecureCookies})
	server := app.NewHTTPServer(cfg, store, admin.AuthConfig{
		Username:          cfg.AdminUsername,
		Password:          cfg.AdminPassword,
		Sessions:          sessions,
		Limiter:           limiter,
		AuthPolicyLimiter: authPolicyLimiter,
		DiscoveryLimiter:  discoveryLimiter,
		AuthPolicyToken:   cfg.AuthPolicyToken,
		System: admin.SystemConfig{
			MailctlPath:            cfg.MailctlPath,
			ConfigApplyRequestPath: cfg.ConfigApplyRequestPath,
		},
	})
	log.Printf("webadmin listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("webadmin stopped: %v", err)
	}
}
