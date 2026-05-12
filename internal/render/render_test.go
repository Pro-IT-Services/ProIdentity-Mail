package render

import (
	"strings"
	"testing"
)

func TestRenderPostfixMainIncludesVirtualMailboxDomain(t *testing.T) {
	out, err := RenderPostfixMain(PostfixMainData{
		Hostname:    "mail.example.com",
		TLSCertFile: "/etc/letsencrypt/live/admin.example.com/fullchain.pem",
		TLSKeyFile:  "/etc/letsencrypt/live/admin.example.com/privkey.pem",
		SNIEnabled:  true,
		SNIMapPath:  "/etc/postfix/proidentity/tls-sni-map",
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
	if strings.Contains(text, "smtpd_sasl_auth_enable = yes") {
		t.Fatalf("port 25 must not inherit SMTP AUTH from main.cf: %s", text)
	}
	if !strings.Contains(text, "smtpd_relay_restrictions = permit_mynetworks,permit_sasl_authenticated,reject_unauth_destination") {
		t.Fatalf("rendered config missing safe relay restrictions: %s", text)
	}
	if !strings.Contains(text, "smtpd_sender_login_maps = mysql:/etc/postfix/proidentity/sender-login-maps.cf") {
		t.Fatalf("rendered config missing sender login map: %s", text)
	}
	if !strings.Contains(text, "smtpd_client_restrictions = permit_mynetworks,reject_rbl_client zen.spamhaus.org") {
		t.Fatalf("rendered config must apply DNSBL checks on inbound SMTP: %s", text)
	}
	if !strings.Contains(text, "disable_vrfy_command = yes") {
		t.Fatalf("rendered config must disable VRFY: %s", text)
	}
	if !strings.Contains(text, "smtpd_tls_protocols = !SSLv2,!SSLv3,!TLSv1,!TLSv1.1") {
		t.Fatalf("rendered config must disable legacy TLS protocols: %s", text)
	}
	if !strings.Contains(text, "smtpd_tls_cert_file = /etc/letsencrypt/live/admin.example.com/fullchain.pem") {
		t.Fatalf("rendered config missing TLS cert path: %s", text)
	}
	if !strings.Contains(text, "smtpd_tls_key_file = /etc/letsencrypt/live/admin.example.com/privkey.pem") {
		t.Fatalf("rendered config missing TLS key path: %s", text)
	}
	if !strings.Contains(text, "tls_server_sni_maps = hash:/etc/postfix/proidentity/tls-sni-map") {
		t.Fatalf("rendered config missing SNI map: %s", text)
	}
	if !strings.Contains(text, "milter_mail_macros = i {auth_type} {auth_authen} {auth_author} {mail_addr}") {
		t.Fatalf("rendered config missing auth milter macros for DKIM signing: %s", text)
	}
	for _, want := range []string{
		"anvil_rate_time_unit = 60s",
		"smtpd_client_connection_count_limit = 20",
		"smtpd_client_connection_rate_limit = 30",
		"smtpd_client_auth_rate_limit = 10",
		"smtpd_client_new_tls_session_rate_limit = 20",
		"smtpd_client_event_limit_exceptions = $mynetworks",
		"smtpd_timeout = 60s",
		"smtpd_starttls_timeout = 60s",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered config missing SMTP abuse limit %q: %s", want, text)
		}
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
	if !strings.Contains(text, "smtps     inet  n       -       y       -       -       smtpd") {
		t.Fatalf("rendered master missing SMTPS service: %s", text)
	}
	if !strings.Contains(text, "-o smtpd_sender_restrictions=reject_authenticated_sender_login_mismatch,permit_sasl_authenticated,reject") {
		t.Fatalf("submission must prevent authenticated sender spoofing: %s", text)
	}
	for _, want := range []string{
		"-o smtpd_client_connection_count_limit=10",
		"-o smtpd_client_connection_rate_limit=20",
		"-o smtpd_client_auth_rate_limit=10",
		"-o smtpd_client_new_tls_session_rate_limit=20",
	} {
		if count := strings.Count(text, want); count != 2 {
			t.Fatalf("submission and smtps should both include %q exactly twice, got %d in %s", want, count, text)
		}
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

func TestRenderPostfixVirtualAliasMapIncludesCatchAllRoutes(t *testing.T) {
	out, err := RenderPostfixVirtualAliasMaps(PostfixMySQLData{
		Database: "maildb",
		User:     "mailuser",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("RenderPostfixVirtualAliasMaps returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{"FROM aliases", "catch_all_routes", "SUBSTRING_INDEX('%s', '@', -1)", "NOT EXISTS", "FROM users"} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered alias map missing %q: %s", want, text)
		}
	}
}

func TestRenderPostfixSenderLoginMapAllowsOwnedAliasesAndSharedSendAs(t *testing.T) {
	out, err := RenderPostfixSenderLoginMaps(PostfixMySQLData{
		Database: "maildb",
		User:     "mailuser",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("RenderPostfixSenderLoginMaps returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"query = SELECT CONCAT(u.local_part, '@', d.name)",
		"FROM aliases",
		"shared_mailbox_permissions",
		"p.can_send_as=1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered sender login map missing %q: %s", want, text)
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
		"u.mailbox_type = 'user'",
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

func TestRenderDovecotSQLSupportsAppPasswordsAndMFAPolicy(t *testing.T) {
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
		"user_app_passwords ap",
		"SHA2('%{password}', 256)",
		"NULL AS password",
		"'Y' AS nopassword",
		"mail_server_settings ms",
		"user_mfa_settings mfa",
		"COALESCE(ms.force_mailbox_mfa, 0) = 0",
		"COALESCE(mfa.totp_enabled, 0) = 0",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered dovecot SQL missing app-password/MFA clause %q: %s", want, text)
		}
	}
}

func TestRenderDovecotLocalIncludesMailStorageAndAuthSocket(t *testing.T) {
	out, err := RenderDovecotLocal(DovecotLocalData{
		TLSCertFile: "/etc/letsencrypt/live/admin.example.com/fullchain.pem",
		TLSKeyFile:  "/etc/letsencrypt/live/admin.example.com/privkey.pem",
		SNIHosts: []MailServerSNIHost{{
			Hostname:    "mail.customer.test",
			TLSCertFile: "/etc/letsencrypt/live/mail.customer.test/fullchain.pem",
			TLSKeyFile:  "/etc/letsencrypt/live/mail.customer.test/privkey.pem",
		}},
	})
	if err != nil {
		t.Fatalf("RenderDovecotLocal returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"mail_driver = maildir",
		"mail_path = /var/vmail/%{user|domain}/%{user|username}/Maildir",
		"mail_inbox_path = /var/vmail/%{user|domain}/%{user|username}/Maildir",
		"sieve_script personal {",
		"path = ~/sieve",
		"active_path = ~/.dovecot.sieve",
		"auth_failure_delay = 2 secs",
		"auth_internal_failure_delay = 2 secs",
		"auth_verbose = yes",
		"mail_max_userip_connections = 20",
		"ssl = required",
		"ssl_server_cert_file = /etc/letsencrypt/live/admin.example.com/fullchain.pem",
		"ssl_server_key_file = /etc/letsencrypt/live/admin.example.com/privkey.pem",
		"local_name mail.customer.test",
		"ssl_server_cert_file = /etc/letsencrypt/live/mail.customer.test/fullchain.pem",
		"ssl_server_key_file = /etc/letsencrypt/live/mail.customer.test/privkey.pem",
		"!include /etc/dovecot/proidentity-sql.conf.ext",
		"auth_username_format = %{user|lower}",
		"unix_listener /var/spool/postfix/private/auth",
		"unix_listener /var/spool/postfix/private/dovecot-lmtp",
		"service imap-login {",
		"restart_request_count = 1",
		"process_limit = 256",
		"service pop3-login {",
		"protocol lmtp {",
		"mail_plugins {",
		"sieve = yes",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered dovecot local config missing %q: %s", want, text)
		}
	}
}

func TestRenderDovecotLocalConfiguresAuthPolicyWhenProvided(t *testing.T) {
	out, err := RenderDovecotLocal(DovecotLocalData{
		AuthPolicy: DovecotAuthPolicyData{
			ServerURL: "http://127.0.0.1:8080/internal/dovecot/auth-policy",
			APIHeader: "X-ProIdentity-Auth-Policy: secret-token",
			Nonce:     "nonce-value",
		},
	})
	if err != nil {
		t.Fatalf("RenderDovecotLocal returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"auth_policy_server_url = http://127.0.0.1:8080/internal/dovecot/auth-policy",
		"auth_policy_server_api_header = X-ProIdentity-Auth-Policy: secret-token",
		"auth_policy_hash_nonce = nonce-value",
		"auth_policy_reject_on_fail = yes",
		"remote = %{remote_ip}",
		"success = %{success}",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered dovecot local config missing %q: %s", want, text)
		}
	}
}

func TestRenderPostfixSNIMapUsesCombinedChainFiles(t *testing.T) {
	out, err := RenderPostfixSNIMap([]MailServerSNIHost{{
		Hostname:     "mail.customer.test",
		TLSChainFile: "/etc/postfix/proidentity/tls-sni/mail.customer.test.pem",
		TLSCertFile:  "/etc/letsencrypt/live/mail.customer.test/fullchain.pem",
		TLSKeyFile:   "/etc/letsencrypt/live/mail.customer.test/privkey.pem",
	}})
	if err != nil {
		t.Fatalf("RenderPostfixSNIMap returned error: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "mail.customer.test\t/etc/postfix/proidentity/tls-sni/mail.customer.test.pem" {
		t.Fatalf("unexpected SNI map: %q", got)
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
		TLSMode:              "letsencrypt-http",
		AdminHostname:        "admin.example.com",
		WebmailHostname:      "mail.example.com",
		DAVHostname:          "dav.example.com",
		AutoconfigHostname:   "autoconfig.example.com",
		AutodiscoverHostname: "autodiscover.example.com",
		ACMEWebroot:          "/var/lib/proidentity-mail/acme",
		ForceHTTPS:           true,
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
		"server_name autoconfig.example.com autodiscover.example.com;",
		"location = /.well-known/autoconfig/mail/config-v1.1.xml",
		"location = /mail/config-v1.1.xml",
		"location = /autodiscover/autodiscover.xml",
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

func TestRenderNginxProxyDoesNotDuplicateSharedDAVHost(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:         "none",
		AdminHostname:   "admin.example.com",
		WebmailHostname: "webmail.example.com",
		DAVHostname:     "webmail.example.com",
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	text := string(out)
	if count := strings.Count(text, "server_name webmail.example.com;"); count != 1 {
		t.Fatalf("shared webmail/DAV hostname should render once, got %d in %s", count, text)
	}
	if !strings.Contains(text, "location /dav/") {
		t.Fatalf("shared webmail/DAV hostname should still expose DAV routes: %s", text)
	}
}

func TestRenderNginxProxyRoutesMailHostForDAVDiscovery(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:         "letsencrypt-dns-cloudflare",
		AdminHostname:   "madmin.example.com",
		WebmailHostname: "webmail.example.com",
		DAVHostname:     "webmail.example.com",
		MailHostname:    "mail.example.com",
		ForceHTTPS:      true,
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"server_name mail.example.com;",
		"location /dav/",
		"location /.well-known/caldav",
		"location /.well-known/carddav",
		"proxy_pass http://127.0.0.1:8081;",
		"proxy_pass http://127.0.0.1:8080;",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("mail host DAV discovery config missing %q: %s", want, text)
		}
	}
}

func TestRenderNginxProxyRoutesMailHostRootToWebmail(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:         "letsencrypt-dns-cloudflare",
		AdminHostname:   "madmin.example.com",
		WebmailHostname: "webmail.example.com",
		DAVHostname:     "webmail.example.com",
		MailHostname:    "mail.example.com",
		ForceHTTPS:      true,
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	block := serverBlockForNameContaining(t, string(out), "mail.example.com", "listen 443 ssl;")
	for _, want := range []string{
		"location / {",
		"proxy_pass http://127.0.0.1:8082;",
		"location /dav/",
		"location = /.well-known/autoconfig/mail/config-v1.1.xml",
	} {
		if !strings.Contains(block, want) {
			t.Fatalf("mail host server block missing %q: %s", want, block)
		}
	}
	if strings.Contains(block, "return 404;") {
		t.Fatalf("mail host root should serve webmail, not 404: %s", block)
	}
}

func TestRenderNginxProxyBlocksUnknownHosts(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:         "none",
		AdminHostname:   "admin.example.com",
		WebmailHostname: "webmail.example.com",
		DAVHostname:     "webmail.example.com",
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"listen 80 default_server;",
		"server_name _;",
		"return 444;",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("unknown-host default block missing %q: %s", want, text)
		}
	}
}

func TestRenderNginxProxyBlocksInternalAdminCallbacksPublicly(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:         "none",
		AdminHostname:   "madmin.example.com",
		WebmailHostname: "webmail.example.com",
		DAVHostname:     "webmail.example.com",
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	block := serverBlockForName(t, string(out), "madmin.example.com")
	if !strings.Contains(block, "location ^~ /internal/") || !strings.Contains(block, "return 404;") {
		t.Fatalf("admin public server block should hide internal callbacks: %s", block)
	}
}

func serverBlockForName(t *testing.T, text, serverName string) string {
	t.Helper()
	marker := "server_name " + serverName + ";"
	idx := strings.Index(text, marker)
	if idx < 0 {
		t.Fatalf("server_name %s not found in %s", serverName, text)
	}
	start := strings.LastIndex(text[:idx], "\nserver {")
	if start < 0 {
		start = strings.LastIndex(text[:idx], "server {")
	}
	if start < 0 {
		t.Fatalf("server block start for %s not found in %s", serverName, text)
	}
	next := strings.Index(text[idx:], "\nserver {")
	if next < 0 {
		return text[start:]
	}
	return text[start : idx+next]
}

func serverBlockForNameContaining(t *testing.T, text, serverName, contains string) string {
	t.Helper()
	searchFrom := 0
	marker := "server_name " + serverName + ";"
	for {
		idx := strings.Index(text[searchFrom:], marker)
		if idx < 0 {
			t.Fatalf("server_name %s with %q not found in %s", serverName, contains, text)
		}
		idx += searchFrom
		start := strings.LastIndex(text[:idx], "\nserver {")
		if start < 0 {
			start = strings.LastIndex(text[:idx], "server {")
		}
		if start < 0 {
			t.Fatalf("server block start for %s not found in %s", serverName, text)
		}
		next := strings.Index(text[idx:], "\nserver {")
		block := text[start:]
		if next >= 0 {
			block = text[start : idx+next]
		}
		if strings.Contains(block, contains) {
			return block
		}
		searchFrom = idx + len(marker)
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

func TestRenderNginxProxySupportsCloudflareRealIP(t *testing.T) {
	out, err := RenderNginxProxy(NginxProxyData{
		TLSMode:                 "behind-proxy",
		AdminHostname:           "madmin.example.com",
		WebmailHostname:         "webmail.example.com",
		CloudflareRealIPEnabled: true,
	})
	if err != nil {
		t.Fatalf("RenderNginxProxy returned error: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"set_real_ip_from 173.245.48.0/20;",
		"set_real_ip_from 2a06:98c0::/29;",
		"real_ip_header CF-Connecting-IP;",
		"real_ip_recursive on;",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("cloudflare real IP config missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, "real_ip_header X-Forwarded-For;") {
		t.Fatalf("cloudflare real IP config must not trust X-Forwarded-For: %s", text)
	}
}

func TestRenderCertbotScriptSupportsCloudflareDNS(t *testing.T) {
	out, err := RenderCertbotScript(CertbotScriptData{
		TLSMode:                   "letsencrypt-dns-cloudflare",
		Hostnames:                 []string{"admin.example.com", "mail.example.com"},
		CloudflareCredentialsFile: "/etc/proidentity-mail/cloudflare.ini",
		CloudflareCertDomain:      "example.com",
		MailctlPath:               "/opt/proidentity-mail/bin/mailctl",
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
		"/opt/proidentity-mail/bin/mailctl\" cloudflare-cert-credentials --domain \"example.com\" --output \"/etc/proidentity-mail/cloudflare.ini\"",
		"-d admin.example.com",
		"-d mail.example.com",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("certbot script missing %q: %s", want, text)
		}
	}
}
