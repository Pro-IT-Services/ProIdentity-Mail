package app

import "os"

type Config struct {
	HTTPAddr  string
	DBDSN     string
	ConfigDir string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:  valueOrDefault(os.Getenv("PROIDENTITY_HTTP_ADDR"), "127.0.0.1:8080"),
		DBDSN:     os.Getenv("PROIDENTITY_DB_DSN"),
		ConfigDir: valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_DIR"), "/etc/proidentity-mail/generated"),
	}
	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
