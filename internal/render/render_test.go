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
	if !strings.Contains(text, "milter_mail_macros = i {auth_type} {auth_authen} {auth_author} {mail_addr}") {
		t.Fatalf("rendered config missing auth milter macros for DKIM signing: %s", text)
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

func TestRenderRspamdAntivirusEnablesClamAVReject(t *testing.T) {
	out, err := RenderRspamdAntivirus()
	if err != nil {
		t.Fatalf("RenderRspamdAntivirus returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"clamav {",
		`type = "clamav";`,
		`servers = "/run/clamav/clamd.ctl";`,
		`action = "reject";`,
		`symbol = "CLAM_VIRUS";`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("antivirus config missing %q: %s", want, text)
		}
	}
}

func TestRenderRspamdDKIMSigningUsesDomainKeyPath(t *testing.T) {
	out, err := RenderRspamdDKIMSigning(RspamdDKIMSigningData{
		Domains: []DKIMSigningDomain{
			{
				Domain:   "example.com",
				Selector: "mail",
				KeyPath:  "/var/lib/rspamd/dkim/example.com.mail.key",
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderRspamdDKIMSigning returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"enabled = true;",
		"example.com {",
		`path = "/var/lib/rspamd/dkim/example.com.mail.key";`,
		`selector = "mail";`,
		"sign_authenticated = true;",
		"sign_local = true;",
		"sign_inbound = false;",
		"use_esld = false;",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dkim signing config missing %q: %s", want, text)
		}
	}
}

func TestRenderRspamdActionsAndHeaders(t *testing.T) {
	actions, err := RenderRspamdActions()
	if err != nil {
		t.Fatalf("RenderRspamdActions returned error: %v", err)
	}
	headers, err := RenderRspamdMilterHeaders()
	if err != nil {
		t.Fatalf("RenderRspamdMilterHeaders returned error: %v", err)
	}
	if !strings.Contains(string(actions), "reject = 15;") || !strings.Contains(string(actions), "add_header = 6;") {
		t.Fatalf("actions config missing thresholds: %s", string(actions))
	}
	if !strings.Contains(string(headers), `"x-spamd-result"`) || !strings.Contains(string(headers), `"authentication-results"`) {
		t.Fatalf("milter headers config missing useful headers: %s", string(headers))
	}
}

func TestRenderRspamdTenantSettingsAppliesPerDomainActions(t *testing.T) {
	out, err := RenderRspamdTenantSettings(RspamdTenantPolicyData{
		Domains: []RspamdTenantPolicyDomain{
			{Domain: "example.com", SpamAction: "quarantine"},
			{Domain: "example.net", SpamAction: "reject"},
			{Domain: "example.org", SpamAction: "mark"},
		},
	})
	if err != nil {
		t.Fatalf("RenderRspamdTenantSettings returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		`proidentity_example_com {`,
		`rcpt = "@example.com";`,
		`"quarantine" = 6.0;`,
		`reject = 999.0;`,
		`proidentity_example_net {`,
		`reject = 6.0;`,
		`proidentity_example_org {`,
		`"add header" = 6.0;`,
		`subject = "[SPAM] %s";`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("tenant settings missing %q: %s", want, text)
		}
	}
}

func TestRenderRspamdForceActionsAppliesMalwarePolicy(t *testing.T) {
	out, err := RenderRspamdForceActions(RspamdTenantPolicyData{
		Domains: []RspamdTenantPolicyDomain{
			{Domain: "example.com", MalwareAction: "quarantine"},
			{Domain: "example.net", MalwareAction: "reject"},
		},
	})
	if err != nil {
		t.Fatalf("RenderRspamdForceActions returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		`PROIDENTITY_MALWARE_EXAMPLE_COM {`,
		`action = "quarantine";`,
		`expression = "CLAM_VIRUS & RCPT_DOMAIN_EXAMPLE_COM";`,
		`PROIDENTITY_MALWARE_EXAMPLE_NET {`,
		`action = "reject";`,
		`message = "Rejected due to malware policy";`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("force actions missing %q: %s", want, text)
		}
	}
}

func TestRenderNginxProxySupportsManagedHTTPChallengeTLS(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:         "letsencrypt-http",
		AdminHostname:   "admin.example.com",
		WebmailHostname: "mail.example.com",
		DAVHostname:     "dav.example.com",
		ACMEWebroot:     "/var/lib/proidentity-mail/acme",
		ForceHTTPS:      true,
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"server_name admin.example.com;",
		"proxy_pass http://127.0.0.1:8080;",
		"server_name mail.example.com;",
		"proxy_pass http://127.0.0.1:8082;",
		"server_name dav.example.com;",
		"proxy_pass http://127.0.0.1:8081;",
		"location ^~ /.well-known/acme-challenge/",
		"return 301 https://$host$request_uri;",
		"ssl_certificate /etc/letsencrypt/live/admin.example.com/fullchain.pem;",
		"add_header Strict-Transport-Security",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("nginx proxy config missing %q: %s", want, text)
		}
	}
}

func TestRenderNginxProxySupportsBehindProxyTrustedCIDRs(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:           "behind-proxy",
		AdminHostname:     "admin.internal",
		WebmailHostname:   "mail.internal",
		DAVHostname:       "dav.internal",
		TrustedProxyCIDRs: []string{"10.0.0.0/8", "192.168.0.0/16"},
		TrustProxyHeaders: true,
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"listen 80;",
		"set_real_ip_from 10.0.0.0/8;",
		"set_real_ip_from 192.168.0.0/16;",
		"real_ip_header X-Forwarded-For;",
		`"~^https?$" $http_x_forwarded_proto;`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("behind-proxy config missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, "listen 443 ssl") {
		t.Fatalf("behind-proxy mode should not force internal TLS: %s", text)
	}
}

func TestRenderCertbotScriptSupportsCloudflareDNS(t *testing.T) {
	out, err := RenderCertbotScript(CertbotScriptData{
		TLSMode:                   "letsencrypt-dns-cloudflare",
		Hostnames:                 []string{"admin.example.com", "mail.example.com"},
		CloudflareCredentialsFile: "/etc/proidentity-mail/cloudflare.ini",
		CloudflarePropagationSec:  60,
	})
	if err != nil {
		t.Fatalf("RenderCertbotScript returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"certbot certonly",
		"--dns-cloudflare",
		"--dns-cloudflare-credentials /etc/proidentity-mail/cloudflare.ini",
		"--dns-cloudflare-propagation-seconds 60",
		"-d admin.example.com",
		"-d mail.example.com",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("certbot script missing %q: %s", want, text)
		}
	}
}
