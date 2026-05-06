package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "")
	t.Setenv("PROIDENTITY_DB_NAME", "")
	t.Setenv("PROIDENTITY_DB_USER", "")
	t.Setenv("PROIDENTITY_DB_PASSWORD", "")
	t.Setenv("PROIDENTITY_HTTP_ADDR", "")
	t.Setenv("PROIDENTITY_GROUPWARE_ADDR", "")
	t.Setenv("PROIDENTITY_WEBMAIL_ADDR", "")
	t.Setenv("PROIDENTITY_ADMIN_USERNAME", "")
	t.Setenv("PROIDENTITY_ADMIN_PASSWORD", "")
	t.Setenv("PROIDENTITY_CONFIG_DIR", "")
	t.Setenv("PROIDENTITY_MAIL_HOSTNAME", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != "0.0.0.0:8080" {
		t.Fatalf("HTTPAddr = %q, want default", cfg.HTTPAddr)
	}
	if cfg.GroupwareAddr != "0.0.0.0:8081" {
		t.Fatalf("GroupwareAddr = %q, want default", cfg.GroupwareAddr)
	}
	if cfg.WebmailAddr != "0.0.0.0:8082" {
		t.Fatalf("WebmailAddr = %q, want default", cfg.WebmailAddr)
	}
	if cfg.AdminUsername != "" || cfg.AdminPassword != "" {
		t.Fatalf("admin credentials should default empty")
	}
	if cfg.ConfigDir != "/etc/proidentity-mail/generated" {
		t.Fatalf("ConfigDir = %q, want default", cfg.ConfigDir)
	}
	if cfg.DBName != "proidentity_mail" {
		t.Fatalf("DBName = %q, want default", cfg.DBName)
	}
	if cfg.DBUser != "proidentity_mail" {
		t.Fatalf("DBUser = %q, want default", cfg.DBUser)
	}
	if cfg.MailHostname != "mail.local" {
		t.Fatalf("MailHostname = %q, want default", cfg.MailHostname)
	}
	if cfg.ProxyMode != "internal-nginx" || cfg.TLSMode != "behind-proxy" {
		t.Fatalf("proxy defaults = %q/%q, want internal-nginx/behind-proxy", cfg.ProxyMode, cfg.TLSMode)
	}
	if cfg.ACMEWebroot != "/var/lib/proidentity-mail/acme" {
		t.Fatalf("ACMEWebroot = %q, want default", cfg.ACMEWebroot)
	}
}

func TestLoadConfigRequiresDSNForDatabaseUse(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "mail:secret@tcp(127.0.0.1:3306)/proidentity_mail?parseTime=true")
	t.Setenv("PROIDENTITY_ADMIN_USERNAME", "root")
	t.Setenv("PROIDENTITY_ADMIN_PASSWORD", "secret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.DBDSN == "" {
		t.Fatal("DBDSN is empty")
	}
	if cfg.AdminUsername != "root" || cfg.AdminPassword != "secret" {
		t.Fatalf("admin credentials not loaded")
	}
}

func TestLoadConfigDerivesProxyHostnamesFromPublicHostname(t *testing.T) {
	t.Setenv("PROIDENTITY_PUBLIC_HOSTNAME", "example.com")
	t.Setenv("PROIDENTITY_TLS_MODE", "letsencrypt-dns-cloudflare")
	t.Setenv("PROIDENTITY_CLOUDFLARE_CREDENTIALS_FILE", "/etc/proidentity-mail/cloudflare.ini")
	t.Setenv("PROIDENTITY_TRUSTED_PROXY_CIDRS", "10.0.0.0/8, 192.168.0.0/16")
	t.Setenv("PROIDENTITY_TRUST_PROXY_HEADERS", "true")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AdminHostname != "admin.example.com" || cfg.WebmailHostname != "mail.example.com" || cfg.DAVHostname != "dav.example.com" {
		t.Fatalf("derived hostnames = %q/%q/%q", cfg.AdminHostname, cfg.WebmailHostname, cfg.DAVHostname)
	}
	if cfg.TLSMode != "letsencrypt-dns-cloudflare" || cfg.CloudflareCredentialsFile == "" {
		t.Fatalf("tls config not loaded")
	}
	if len(cfg.TrustedProxyCIDRs) != 2 || !cfg.TrustProxyHeaders {
		t.Fatalf("trusted proxy config not loaded: %+v", cfg.TrustedProxyCIDRs)
	}
}
