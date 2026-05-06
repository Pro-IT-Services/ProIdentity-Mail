package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "")
	t.Setenv("PROIDENTITY_DB_NAME", "")
	t.Setenv("PROIDENTITY_DB_USER", "")
	t.Setenv("PROIDENTITY_DB_PASSWORD", "")
	t.Setenv("PROIDENTITY_HTTP_ADDR", "")
	t.Setenv("PROIDENTITY_CONFIG_DIR", "")
	t.Setenv("PROIDENTITY_MAIL_HOSTNAME", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != "0.0.0.0:8080" {
		t.Fatalf("HTTPAddr = %q, want default", cfg.HTTPAddr)
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
}

func TestLoadConfigRequiresDSNForDatabaseUse(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "mail:secret@tcp(127.0.0.1:3306)/proidentity_mail?parseTime=true")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.DBDSN == "" {
		t.Fatal("DBDSN is empty")
	}
}
