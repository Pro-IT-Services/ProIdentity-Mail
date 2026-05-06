package app

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr        string
	GroupwareAddr   string
	WebmailAddr     string
	DBDSN           string
	DBName          string
	DBUser          string
	DBPassword      string
	ConfigDir       string
	QuarantineDir   string
	MailRoot        string
	ReleaseSMTPAddr string
	MailHostname    string
	AdminUsername   string
	AdminPassword   string
	SecureCookies   bool
}

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:        valueOrDefault(os.Getenv("PROIDENTITY_HTTP_ADDR"), "0.0.0.0:8080"),
		GroupwareAddr:   valueOrDefault(os.Getenv("PROIDENTITY_GROUPWARE_ADDR"), "0.0.0.0:8081"),
		WebmailAddr:     valueOrDefault(os.Getenv("PROIDENTITY_WEBMAIL_ADDR"), "0.0.0.0:8082"),
		DBDSN:           os.Getenv("PROIDENTITY_DB_DSN"),
		DBName:          valueOrDefault(os.Getenv("PROIDENTITY_DB_NAME"), "proidentity_mail"),
		DBUser:          valueOrDefault(os.Getenv("PROIDENTITY_DB_USER"), "proidentity_mail"),
		DBPassword:      os.Getenv("PROIDENTITY_DB_PASSWORD"),
		ConfigDir:       valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_DIR"), "/etc/proidentity-mail/generated"),
		QuarantineDir:   valueOrDefault(os.Getenv("PROIDENTITY_QUARANTINE_DIR"), "/var/lib/proidentity-mail/quarantine"),
		MailRoot:        valueOrDefault(os.Getenv("PROIDENTITY_MAIL_ROOT"), "/var/vmail"),
		ReleaseSMTPAddr: valueOrDefault(os.Getenv("PROIDENTITY_RELEASE_SMTP_ADDR"), "127.0.0.1:25"),
		MailHostname:    valueOrDefault(os.Getenv("PROIDENTITY_MAIL_HOSTNAME"), "mail.local"),
		AdminUsername:   os.Getenv("PROIDENTITY_ADMIN_USERNAME"),
		AdminPassword:   os.Getenv("PROIDENTITY_ADMIN_PASSWORD"),
		SecureCookies:   os.Getenv("PROIDENTITY_SECURE_COOKIES") == "1" || strings.EqualFold(os.Getenv("PROIDENTITY_SECURE_COOKIES"), "true"),
	}
	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
