package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "")
	t.Setenv("PROIDENTITY_HTTP_ADDR", "")
	t.Setenv("PROIDENTITY_CONFIG_DIR", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("HTTPAddr = %q, want default", cfg.HTTPAddr)
	}
	if cfg.ConfigDir != "/etc/proidentity-mail/generated" {
		t.Fatalf("ConfigDir = %q, want default", cfg.ConfigDir)
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
