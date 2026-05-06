package app

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr                  string
	GroupwareAddr             string
	WebmailAddr               string
	DBDSN                     string
	DBName                    string
	DBUser                    string
	DBPassword                string
	ConfigDir                 string
	QuarantineDir             string
	MailRoot                  string
	ReleaseSMTPAddr           string
	MailHostname              string
	ProxyMode                 string
	TLSMode                   string
	PublicHostname            string
	AdminHostname             string
	WebmailHostname           string
	DAVHostname               string
	ACMEWebroot               string
	TLSCertPath               string
	TLSKeyPath                string
	CloudflareCredentialsFile string
	TrustedProxyCIDRs         []string
	ForceHTTPS                bool
	TrustProxyHeaders         bool
	AdminUsername             string
	AdminPassword             string
	SecureCookies             bool
}

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:                  valueOrDefault(os.Getenv("PROIDENTITY_HTTP_ADDR"), "0.0.0.0:8080"),
		GroupwareAddr:             valueOrDefault(os.Getenv("PROIDENTITY_GROUPWARE_ADDR"), "0.0.0.0:8081"),
		WebmailAddr:               valueOrDefault(os.Getenv("PROIDENTITY_WEBMAIL_ADDR"), "0.0.0.0:8082"),
		DBDSN:                     os.Getenv("PROIDENTITY_DB_DSN"),
		DBName:                    valueOrDefault(os.Getenv("PROIDENTITY_DB_NAME"), "proidentity_mail"),
		DBUser:                    valueOrDefault(os.Getenv("PROIDENTITY_DB_USER"), "proidentity_mail"),
		DBPassword:                os.Getenv("PROIDENTITY_DB_PASSWORD"),
		ConfigDir:                 valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_DIR"), "/etc/proidentity-mail/generated"),
		QuarantineDir:             valueOrDefault(os.Getenv("PROIDENTITY_QUARANTINE_DIR"), "/var/lib/proidentity-mail/quarantine"),
		MailRoot:                  valueOrDefault(os.Getenv("PROIDENTITY_MAIL_ROOT"), "/var/vmail"),
		ReleaseSMTPAddr:           valueOrDefault(os.Getenv("PROIDENTITY_RELEASE_SMTP_ADDR"), "127.0.0.1:25"),
		MailHostname:              valueOrDefault(os.Getenv("PROIDENTITY_MAIL_HOSTNAME"), "mail.local"),
		ProxyMode:                 valueOrDefault(os.Getenv("PROIDENTITY_PROXY_MODE"), "internal-nginx"),
		TLSMode:                   valueOrDefault(os.Getenv("PROIDENTITY_TLS_MODE"), "behind-proxy"),
		PublicHostname:            os.Getenv("PROIDENTITY_PUBLIC_HOSTNAME"),
		AdminHostname:             os.Getenv("PROIDENTITY_ADMIN_HOSTNAME"),
		WebmailHostname:           os.Getenv("PROIDENTITY_WEBMAIL_HOSTNAME"),
		DAVHostname:               os.Getenv("PROIDENTITY_DAV_HOSTNAME"),
		ACMEWebroot:               valueOrDefault(os.Getenv("PROIDENTITY_ACME_WEBROOT"), "/var/lib/proidentity-mail/acme"),
		TLSCertPath:               os.Getenv("PROIDENTITY_TLS_CERT_PATH"),
		TLSKeyPath:                os.Getenv("PROIDENTITY_TLS_KEY_PATH"),
		CloudflareCredentialsFile: os.Getenv("PROIDENTITY_CLOUDFLARE_CREDENTIALS_FILE"),
		TrustedProxyCIDRs:         splitCSV(os.Getenv("PROIDENTITY_TRUSTED_PROXY_CIDRS")),
		ForceHTTPS:                boolEnv(os.Getenv("PROIDENTITY_FORCE_HTTPS"), true),
		TrustProxyHeaders:         boolEnv(os.Getenv("PROIDENTITY_TRUST_PROXY_HEADERS"), false),
		AdminUsername:             os.Getenv("PROIDENTITY_ADMIN_USERNAME"),
		AdminPassword:             os.Getenv("PROIDENTITY_ADMIN_PASSWORD"),
		SecureCookies:             os.Getenv("PROIDENTITY_SECURE_COOKIES") == "1" || strings.EqualFold(os.Getenv("PROIDENTITY_SECURE_COOKIES"), "true"),
	}
	if cfg.PublicHostname != "" {
		if cfg.AdminHostname == "" {
			cfg.AdminHostname = "admin." + cfg.PublicHostname
		}
		if cfg.WebmailHostname == "" {
			cfg.WebmailHostname = "mail." + cfg.PublicHostname
		}
		if cfg.DAVHostname == "" {
			cfg.DAVHostname = "dav." + cfg.PublicHostname
		}
	}
	if cfg.AdminHostname == "" {
		cfg.AdminHostname = "admin." + cfg.MailHostname
	}
	if cfg.WebmailHostname == "" {
		cfg.WebmailHostname = cfg.MailHostname
	}
	if cfg.DAVHostname == "" {
		cfg.DAVHostname = "dav." + cfg.MailHostname
	}
	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func boolEnv(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}
	return value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
