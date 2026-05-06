package render

import (
	"strings"
	"testing"
)

func TestRenderPostfixMainIncludesVirtualMailboxDomain(t *testing.T) {
	out, err := RenderPostfixMain(PostfixMainData{
		Hostname: "mail.example.com",
	})
	if err != nil {
		t.Fatalf("RenderPostfixMain returned error: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, "myhostname = mail.example.com") {
		t.Fatalf("rendered config missing hostname: %s", text)
	}
	if !strings.Contains(text, "smtpd_milters = inet:127.0.0.1:11332") {
		t.Fatalf("rendered config missing rspamd milter: %s", text)
	}
	if !strings.Contains(text, "virtual_mailbox_domains = mysql:/etc/postfix/proidentity/virtual-mailbox-domains.cf") {
		t.Fatalf("rendered config missing virtual domain map: %s", text)
	}
	if !strings.Contains(text, "smtpd_tls_auth_only = yes") {
		t.Fatalf("rendered config must require TLS for auth: %s", text)
	}
	if !strings.Contains(text, "smtpd_relay_restrictions = permit_mynetworks,permit_sasl_authenticated,defer_unauth_destination") {
		t.Fatalf("rendered config missing safe relay restrictions: %s", text)
	}
	if !strings.Contains(text, "smtpd_tls_cert_file = /etc/ssl/certs/ssl-cert-snakeoil.pem") {
		t.Fatalf("rendered config missing TLS cert path: %s", text)
	}
}

func TestRenderPostfixMasterEnablesSubmission(t *testing.T) {
	out, err := RenderPostfixMaster()
	if err != nil {
		t.Fatalf("RenderPostfixMaster returned error: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, "submission inet n       -       y       -       -       smtpd") {
		t.Fatalf("rendered master missing submission service: %s", text)
	}
	if !strings.Contains(text, "-o smtpd_tls_security_level=encrypt") {
		t.Fatalf("submission must require TLS: %s", text)
	}
}

func TestRenderPostfixMySQLMapUsesCredentials(t *testing.T) {
	out, err := RenderPostfixVirtualMailboxDomains(PostfixMySQLData{
		Database: "maildb",
		User:     "mailuser",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("RenderPostfixVirtualMailboxDomains returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"user = mailuser",
		"password = secret",
		"dbname = maildb",
		"query = SELECT 1 FROM domains WHERE name='%s' AND status IN ('pending','active')",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered map missing %q: %s", want, text)
		}
	}
}

func TestRenderDovecotSQLIncludesUserDBAndFullAddress(t *testing.T) {
	out, err := RenderDovecotSQL(DovecotSQLData{
		Database: "maildb",
		User:     "mailuser",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("RenderDovecotSQL returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"sql_driver = mysql",
		"default_password_scheme = BLF-CRYPT",
		"CONCAT(u.local_part, '@', d.name) AS user",
		"mail_driver",
		"mail_path",
		"5000 AS uid",
		"5000 AS gid",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered dovecot SQL missing %q: %s", want, text)
		}
	}
}

func TestRenderDovecotLocalIncludesMailStorageAndAuthSocket(t *testing.T) {
	out, err := RenderDovecotLocal()
	if err != nil {
		t.Fatalf("RenderDovecotLocal returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"mail_driver = maildir",
		"mail_path = /var/vmail/%{user|domain}/%{user|username}/Maildir",
		"mail_inbox_path = /var/vmail/%{user|domain}/%{user|username}/Maildir",
		"ssl = required",
		"!include /etc/dovecot/proidentity-sql.conf.ext",
		"auth_username_format = %{user|lower}",
		"unix_listener /var/spool/postfix/private/auth",
		"unix_listener /var/spool/postfix/private/dovecot-lmtp",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered dovecot local config missing %q: %s", want, text)
		}
	}
}

func TestRenderRspamdLocalIsLocalDFragment(t *testing.T) {
	out, err := RenderRspamdLocal()
	if err != nil {
		t.Fatalf("RenderRspamdLocal returned error: %v", err)
	}
	text := string(out)
	if strings.Contains(text, "redis {") {
		t.Fatalf("rspamd local.d redis fragment must not nest redis block: %s", text)
	}
	if !strings.Contains(text, `servers = "127.0.0.1";`) {
		t.Fatalf("rspamd redis fragment missing server: %s", text)
	}
}
