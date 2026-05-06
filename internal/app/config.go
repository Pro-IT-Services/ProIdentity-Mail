package app

import "os"

type Config struct {
	HTTPAddr      string
	GroupwareAddr string
	WebmailAddr   string
	DBDSN         string
	DBName        string
	DBUser        string
	DBPassword    string
	ConfigDir     string
	MailHostname  string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:      valueOrDefault(os.Getenv("PROIDENTITY_HTTP_ADDR"), "0.0.0.0:8080"),
		GroupwareAddr: valueOrDefault(os.Getenv("PROIDENTITY_GROUPWARE_ADDR"), "0.0.0.0:8081"),
		WebmailAddr:   valueOrDefault(os.Getenv("PROIDENTITY_WEBMAIL_ADDR"), "0.0.0.0:8082"),
		DBDSN:         os.Getenv("PROIDENTITY_DB_DSN"),
		DBName:        valueOrDefault(os.Getenv("PROIDENTITY_DB_NAME"), "proidentity_mail"),
		DBUser:        valueOrDefault(os.Getenv("PROIDENTITY_DB_USER"), "proidentity_mail"),
		DBPassword:    os.Getenv("PROIDENTITY_DB_PASSWORD"),
		ConfigDir:     valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_DIR"), "/etc/proidentity-mail/generated"),
		MailHostname:  valueOrDefault(os.Getenv("PROIDENTITY_MAIL_HOSTNAME"), "mail.local"),
	}
	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
