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
	PublicIPv4                string
	PublicIPv6                string
	ProxyMode                 string
	TLSMode                   string
	PublicHostname            string
	AdminHostname             string
	WebmailHostname           string
	DAVHostname               string
	AutoconfigHostname        string
	AutodiscoverHostname      string
	ACMEWebroot               string
	TLSCertPath               string
	TLSKeyPath                string
	MailTLSCertPath           string
	MailTLSKeyPath            string
	CloudflareCredentialsFile string
	CloudflareCertDomain      string
	MailctlPath               string
	ConfigApplyRequestPath    string
	TrustedProxyCIDRs         []string
	ForceHTTPS                bool
	TrustProxyHeaders         bool
	AdminUsername             string
	AdminPassword             string
	AuthPolicyToken           string
	AuthPolicyNonce           string
	SecureCookies             bool
}

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:                  valueOrDefault(os.Getenv("PROIDENTITY_HTTP_ADDR"), "127.0.0.1:8080"),
		GroupwareAddr:             valueOrDefault(os.Getenv("PROIDENTITY_GROUPWARE_ADDR"), "127.0.0.1:8081"),
		WebmailAddr:               valueOrDefault(os.Getenv("PROIDENTITY_WEBMAIL_ADDR"), "127.0.0.1:8082"),
		DBDSN:                     os.Getenv("PROIDENTITY_DB_DSN"),
		DBName:                    valueOrDefault(os.Getenv("PROIDENTITY_DB_NAME"), "proidentity_mail"),
		DBUser:                    valueOrDefault(os.Getenv("PROIDENTITY_DB_USER"), "proidentity_mail"),
		DBPassword:                os.Getenv("PROIDENTITY_DB_PASSWORD"),
		ConfigDir:                 valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_DIR"), "/etc/proidentity-mail/generated"),
		QuarantineDir:             valueOrDefault(os.Getenv("PROIDENTITY_QUARANTINE_DIR"), "/var/lib/proidentity-mail/quarantine"),
		MailRoot:                  valueOrDefault(os.Getenv("PROIDENTITY_MAIL_ROOT"), "/var/vmail"),
		ReleaseSMTPAddr:           valueOrDefault(os.Getenv("PROIDENTITY_RELEASE_SMTP_ADDR"), "127.0.0.1:25"),
		MailHostname:              valueOrDefault(os.Getenv("PROIDENTITY_MAIL_HOSTNAME"), "mail.local"),
		PublicIPv4:                os.Getenv("PROIDENTITY_PUBLIC_IPV4"),
		PublicIPv6:                os.Getenv("PROIDENTITY_PUBLIC_IPV6"),
		ProxyMode:                 valueOrDefault(os.Getenv("PROIDENTITY_PROXY_MODE"), "internal-nginx"),
		TLSMode:                   valueOrDefault(os.Getenv("PROIDENTITY_TLS_MODE"), "behind-proxy"),
		PublicHostname:            os.Getenv("PROIDENTITY_PUBLIC_HOSTNAME"),
		AdminHostname:             os.Getenv("PROIDENTITY_ADMIN_HOSTNAME"),
		WebmailHostname:           os.Getenv("PROIDENTITY_WEBMAIL_HOSTNAME"),
		DAVHostname:               os.Getenv("PROIDENTITY_DAV_HOSTNAME"),
		AutoconfigHostname:        os.Getenv("PROIDENTITY_AUTOCONFIG_HOSTNAME"),
		AutodiscoverHostname:      os.Getenv("PROIDENTITY_AUTODISCOVER_HOSTNAME"),
		ACMEWebroot:               valueOrDefault(os.Getenv("PROIDENTITY_ACME_WEBROOT"), "/var/lib/proidentity-mail/acme"),
		TLSCertPath:               os.Getenv("PROIDENTITY_TLS_CERT_PATH"),
		TLSKeyPath:                os.Getenv("PROIDENTITY_TLS_KEY_PATH"),
		MailTLSCertPath:           os.Getenv("PROIDENTITY_MAIL_TLS_CERT_PATH"),
		MailTLSKeyPath:            os.Getenv("PROIDENTITY_MAIL_TLS_KEY_PATH"),
		CloudflareCredentialsFile: os.Getenv("PROIDENTITY_CLOUDFLARE_CREDENTIALS_FILE"),
		CloudflareCertDomain:      os.Getenv("PROIDENTITY_CLOUDFLARE_CERT_DOMAIN"),
		MailctlPath:               valueOrDefault(os.Getenv("PROIDENTITY_MAILCTL_PATH"), "/opt/proidentity-mail/bin/mailctl"),
		ConfigApplyRequestPath:    valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_APPLY_REQUEST_PATH"), "/etc/proidentity-mail/apply-request"),
		TrustedProxyCIDRs:         splitCSV(os.Getenv("PROIDENTITY_TRUSTED_PROXY_CIDRS")),
		ForceHTTPS:                boolEnv(os.Getenv("PROIDENTITY_FORCE_HTTPS"), true),
		TrustProxyHeaders:         boolEnv(os.Getenv("PROIDENTITY_TRUST_PROXY_HEADERS"), false),
		AdminUsername:             os.Getenv("PROIDENTITY_ADMIN_USERNAME"),
		AdminPassword:             os.Getenv("PROIDENTITY_ADMIN_PASSWORD"),
		AuthPolicyToken:           os.Getenv("PROIDENTITY_AUTH_POLICY_TOKEN"),
		AuthPolicyNonce:           os.Getenv("PROIDENTITY_AUTH_POLICY_NONCE"),
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
	applyDiscoveryHostnameDefaults(&cfg)
	if cfg.CloudflareCredentialsFile == "" && cfg.CloudflareCertDomain != "" {
		cfg.CloudflareCredentialsFile = "/etc/proidentity-mail/cloudflare.ini"
	}
	applyMailTLSDefaults(&cfg)
	return cfg, nil
}

func applyDiscoveryHostnameDefaults(cfg *Config) {
	base := strings.TrimSpace(strings.TrimSuffix(cfg.PublicHostname, "."))
	if base == "" {
		base = discoveryBaseFromMailHostname(cfg.MailHostname)
	}
	if base == "" {
		return
	}
	if cfg.AutoconfigHostname == "" {
		cfg.AutoconfigHostname = "autoconfig." + base
	}
	if cfg.AutodiscoverHostname == "" {
		cfg.AutodiscoverHostname = "autodiscover." + base
	}
}

func discoveryBaseFromMailHostname(hostname string) string {
	host := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))
	if strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") {
		return ""
	}
	if strings.HasPrefix(host, "mail.") {
		return strings.TrimPrefix(host, "mail.")
	}
	return ""
}

func applyMailTLSDefaults(cfg *Config) {
	if cfg.MailTLSCertPath == "" {
		cfg.MailTLSCertPath = cfg.TLSCertPath
	}
	if cfg.MailTLSKeyPath == "" {
		cfg.MailTLSKeyPath = cfg.TLSKeyPath
	}
	if cfg.MailTLSCertPath != "" && cfg.MailTLSKeyPath != "" {
		return
	}
	if isManagedTLSMode(cfg.TLSMode) && cfg.AdminHostname != "" {
		if cfg.MailTLSCertPath == "" {
			cfg.MailTLSCertPath = "/etc/letsencrypt/live/" + cfg.AdminHostname + "/fullchain.pem"
		}
		if cfg.MailTLSKeyPath == "" {
			cfg.MailTLSKeyPath = "/etc/letsencrypt/live/" + cfg.AdminHostname + "/privkey.pem"
		}
		return
	}
	if cfg.MailTLSCertPath == "" {
		cfg.MailTLSCertPath = "/etc/ssl/certs/ssl-cert-snakeoil.pem"
	}
	if cfg.MailTLSKeyPath == "" {
		cfg.MailTLSKeyPath = "/etc/ssl/private/ssl-cert-snakeoil.key"
	}
}

func isManagedTLSMode(mode string) bool {
	return strings.EqualFold(mode, "letsencrypt-http") || strings.EqualFold(mode, "letsencrypt-dns-cloudflare") || strings.EqualFold(mode, "custom-cert")
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
