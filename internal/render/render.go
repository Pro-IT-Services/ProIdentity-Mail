package render

import (
	"bytes"
	"strings"
	"text/template"
)

type PostfixMainData struct {
	Hostname    string
	TLSCertFile string
	TLSKeyFile  string
	SNIEnabled  bool
	SNIMapPath  string
}

type DovecotSQLData struct {
	Database string
	User     string
	Password string
}

type DovecotLocalData struct {
	TLSCertFile string
	TLSKeyFile  string
	SNIHosts    []MailServerSNIHost
	AuthPolicy  DovecotAuthPolicyData
}

type DovecotAuthPolicyData struct {
	Enabled   bool
	ServerURL string
	APIHeader string
	Nonce     string
}

type MailServerSNIHost struct {
	Hostname     string
	TLSChainFile string
	TLSCertFile  string
	TLSKeyFile   string
}

type PostfixMySQLData struct {
	Database string
	User     string
	Password string
}

type DKIMSigningDomain struct {
	Domain   string
	Selector string
	KeyPath  string
}

type RspamdDKIMSigningData struct {
	Domains []DKIMSigningDomain
}

type RspamdTenantPolicyDomain struct {
	Domain        string
	SpamAction    string
	MalwareAction string
}

type RspamdTenantPolicyData struct {
	Domains []RspamdTenantPolicyDomain
}

type NginxProxyData struct {
	TLSMode                 string
	AdminHostname           string
	WebmailHostname         string
	DAVHostname             string
	MailHostname            string
	AutoconfigHostname      string
	AutodiscoverHostname    string
	ACMEWebroot             string
	CertPath                string
	KeyPath                 string
	ForceHTTPS              bool
	TrustProxyHeaders       bool
	TrustedProxyCIDRs       []string
	CloudflareRealIPEnabled bool
}

type CertbotScriptData struct {
	TLSMode                   string
	Hostnames                 []string
	ACMEWebroot               string
	CloudflareCredentialsFile string
	CloudflareCertDomain      string
	MailctlPath               string
	CloudflarePropagationSec  int
}

type rspamdTenantPolicyTemplateDomain struct {
	Domain          string
	RuleName        string
	MalwareRuleName string
	SymbolName      string
	SpamAction      string
	MalwareAction   string
}

type nginxProxyTemplateData struct {
	TLSMode              string
	AdminHostname        string
	WebmailHostname      string
	DAVHostname          string
	MailServerName       string
	DiscoveryServerNames string
	SeparateDAVHost      bool
	ACMEWebroot          string
	CertPath             string
	KeyPath              string
	ForceHTTPS           bool
	TrustProxyHeaders    bool
	TrustedProxyCIDRs    []string
	RealIPHeader         string
	TLSEnabled           bool
}

type certbotScriptTemplateData struct {
	TLSMode                   string
	Hostnames                 []string
	ACMEWebroot               string
	CloudflareCredentialsFile string
	CloudflareCertDomain      string
	MailctlPath               string
	CloudflarePropagationSec  int
}

var cloudflareRealIPCIDRs = []string{
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",
	"2400:cb00::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
}

type postfixMySQLTemplateData struct {
	Database string
	User     string
	Password string
	Query    string
}

func RenderPostfixMain(data PostfixMainData) ([]byte, error) {
	if strings.TrimSpace(data.TLSCertFile) == "" {
		data.TLSCertFile = "/etc/ssl/certs/ssl-cert-snakeoil.pem"
	}
	if strings.TrimSpace(data.TLSKeyFile) == "" {
		data.TLSKeyFile = "/etc/ssl/private/ssl-cert-snakeoil.key"
	}
	if data.SNIEnabled && strings.TrimSpace(data.SNIMapPath) == "" {
		data.SNIMapPath = "/etc/postfix/proidentity/tls-sni-map"
	}
	return renderTemplate("postfix-main", postfixMainTemplate, data)
}

func RenderPostfixMaster() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(postfixMasterTemplate), "\n")), nil
}

func RenderPostfixSNIMap(hosts []MailServerSNIHost) ([]byte, error) {
	var builder strings.Builder
	for _, host := range hosts {
		hostname := strings.ToLower(strings.TrimSpace(host.Hostname))
		chain := strings.TrimSpace(host.TLSChainFile)
		if hostname == "" || chain == "" {
			continue
		}
		builder.WriteString(hostname)
		builder.WriteByte('\t')
		builder.WriteString(chain)
		builder.WriteByte('\n')
	}
	return []byte(builder.String()), nil
}

func RenderPostfixVirtualMailboxDomains(data PostfixMySQLData) ([]byte, error) {
	return renderPostfixMySQL(data, "SELECT 1 FROM domains WHERE name='%s' AND status IN ('pending','active')")
}

func RenderPostfixVirtualMailboxMaps(data PostfixMySQLData) ([]byte, error) {
	return renderPostfixMySQL(data, "SELECT CONCAT(d.name, '/', u.local_part, '/Maildir/') FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE CONCAT(u.local_part, '@', d.name)='%s' AND u.status='active' AND d.status IN ('pending','active')")
}

func RenderPostfixVirtualAliasMaps(data PostfixMySQLData) ([]byte, error) {
	query := "SELECT destination FROM aliases a JOIN domains d ON d.id = a.domain_id WHERE CONCAT(a.source_local_part, '@', d.name)='%s' UNION SELECT c.destination FROM catch_all_routes c JOIN domains d ON d.id = c.domain_id WHERE d.name = SUBSTRING_INDEX('%s', '@', -1) AND c.status='active' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.primary_domain_id = d.id AND CONCAT(u.local_part, '@', d.name)='%s' AND u.status='active') AND NOT EXISTS (SELECT 1 FROM aliases a2 WHERE a2.domain_id = d.id AND CONCAT(a2.source_local_part, '@', d.name)='%s')"
	return renderPostfixMySQL(data, query)
}

func RenderPostfixSenderLoginMaps(data PostfixMySQLData) ([]byte, error) {
	query := "SELECT CONCAT(u.local_part, '@', d.name) FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE CONCAT(u.local_part, '@', d.name)='%s' AND u.mailbox_type='user' AND u.status='active' AND d.status IN ('pending','active') UNION SELECT a.destination FROM aliases a JOIN domains d ON d.id = a.domain_id WHERE CONCAT(a.source_local_part, '@', d.name)='%s' UNION SELECT CONCAT(owner.local_part, '@', owner_domain.name) FROM users shared JOIN domains shared_domain ON shared_domain.id = shared.primary_domain_id JOIN shared_mailbox_permissions p ON p.shared_mailbox_id = shared.id JOIN users owner ON owner.id = p.user_id JOIN domains owner_domain ON owner_domain.id = owner.primary_domain_id WHERE CONCAT(shared.local_part, '@', shared_domain.name)='%s' AND shared.mailbox_type='shared' AND shared.status='active' AND shared_domain.status IN ('pending','active') AND owner.mailbox_type='user' AND owner.status='active' AND owner_domain.status IN ('pending','active') AND p.can_send_as=1"
	return renderPostfixMySQL(data, query)
}

func RenderDovecotSQL(data DovecotSQLData) ([]byte, error) {
	return renderTemplate("dovecot-sql", dovecotSQLTemplate, data)
}

func RenderDovecotLocal(data ...DovecotLocalData) ([]byte, error) {
	value := DovecotLocalData{
		TLSCertFile: "/etc/ssl/certs/ssl-cert-snakeoil.pem",
		TLSKeyFile:  "/etc/ssl/private/ssl-cert-snakeoil.key",
	}
	if len(data) > 0 {
		value = data[0]
		if strings.TrimSpace(value.TLSCertFile) == "" {
			value.TLSCertFile = "/etc/ssl/certs/ssl-cert-snakeoil.pem"
		}
		if strings.TrimSpace(value.TLSKeyFile) == "" {
			value.TLSKeyFile = "/etc/ssl/private/ssl-cert-snakeoil.key"
		}
	}
	value.SNIHosts = uniqueSNIHosts(value.SNIHosts)
	if strings.TrimSpace(value.AuthPolicy.ServerURL) == "" || strings.TrimSpace(value.AuthPolicy.APIHeader) == "" || strings.TrimSpace(value.AuthPolicy.Nonce) == "" {
		value.AuthPolicy = DovecotAuthPolicyData{}
	} else {
		value.AuthPolicy.Enabled = true
	}
	return renderTemplate("dovecot-local", dovecotLocalTemplate, value)
}

func RenderRspamdLocal() ([]byte, error) {
	return []byte(rspamdLocalTemplate), nil
}

func RenderRspamdAntivirus() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(rspamdAntivirusTemplate), "\n")), nil
}

func RenderRspamdDKIMSigning(data RspamdDKIMSigningData) ([]byte, error) {
	return renderTemplate("rspamd-dkim-signing", rspamdDKIMSigningTemplate, data)
}

func RenderRspamdActions() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(rspamdActionsTemplate), "\n")), nil
}

func RenderRspamdMilterHeaders() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(rspamdMilterHeadersTemplate), "\n")), nil
}

func RenderRspamdTenantSettings(data RspamdTenantPolicyData) ([]byte, error) {
	return renderTemplate("rspamd-tenant-settings", rspamdTenantSettingsTemplate, map[string]any{"Domains": rspamdPolicyDomains(data)})
}

func RenderRspamdForceActions(data RspamdTenantPolicyData) ([]byte, error) {
	return renderTemplate("rspamd-force-actions", rspamdForceActionsTemplate, map[string]any{"Domains": rspamdPolicyDomains(data)})
}

func RenderNginxProxy(data NginxProxyData) ([]byte, error) {
	normalized := normalizeNginxProxyData(data)
	return renderTemplate("nginx-proxy", nginxProxyTemplate, normalized)
}

func RenderCertbotScript(data CertbotScriptData) ([]byte, error) {
	normalized := certbotScriptTemplateData{
		TLSMode:                   strings.ToLower(strings.TrimSpace(data.TLSMode)),
		Hostnames:                 uniqueNonEmpty(data.Hostnames),
		ACMEWebroot:               valueOrDefault(strings.TrimSpace(data.ACMEWebroot), "/var/lib/proidentity-mail/acme"),
		CloudflareCredentialsFile: strings.TrimSpace(data.CloudflareCredentialsFile),
		CloudflareCertDomain:      strings.TrimSpace(data.CloudflareCertDomain),
		MailctlPath:               valueOrDefault(strings.TrimSpace(data.MailctlPath), "/opt/proidentity-mail/bin/mailctl"),
		CloudflarePropagationSec:  data.CloudflarePropagationSec,
	}
	if normalized.CloudflareCredentialsFile == "" && normalized.CloudflareCertDomain != "" {
		normalized.CloudflareCredentialsFile = "/etc/proidentity-mail/cloudflare.ini"
	}
	if normalized.TLSMode == "" {
		normalized.TLSMode = "behind-proxy"
	}
	if normalized.CloudflarePropagationSec == 0 {
		normalized.CloudflarePropagationSec = 60
	}
	return renderTemplate("certbot-script", certbotScriptTemplate, normalized)
}

func RenderNginxProxyCommon() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(nginxProxyCommonTemplate), "\n")), nil
}

func renderPostfixMySQL(data PostfixMySQLData, query string) ([]byte, error) {
	return renderTemplate("postfix-mysql", postfixMySQLTemplate, postfixMySQLTemplateData{
		Database: data.Database,
		User:     data.User,
		Password: data.Password,
		Query:    query,
	})
}

func renderTemplate(name, text string, data any) ([]byte, error) {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return bytes.TrimLeft(buf.Bytes(), "\n"), nil
}

func rspamdPolicyDomains(data RspamdTenantPolicyData) []rspamdTenantPolicyTemplateDomain {
	out := make([]rspamdTenantPolicyTemplateDomain, 0, len(data.Domains))
	for _, domain := range data.Domains {
		name := rspamdSafeName(domain.Domain)
		out = append(out, rspamdTenantPolicyTemplateDomain{
			Domain:          strings.ToLower(strings.TrimSpace(domain.Domain)),
			RuleName:        "proidentity_" + name,
			MalwareRuleName: "PROIDENTITY_MALWARE_" + strings.ToUpper(name),
			SymbolName:      "RCPT_DOMAIN_" + strings.ToUpper(name),
			SpamAction:      actionOrDefault(domain.SpamAction, "mark"),
			MalwareAction:   actionOrDefault(domain.MalwareAction, "quarantine"),
		})
	}
	return out
}

func actionOrDefault(action, fallback string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "mark", "quarantine", "reject":
		return action
	default:
		return fallback
	}
}

func rspamdSafeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	return strings.Trim(builder.String(), "_")
}

func normalizeNginxProxyData(data NginxProxyData) nginxProxyTemplateData {
	mode := strings.ToLower(strings.TrimSpace(data.TLSMode))
	if mode == "" {
		mode = "behind-proxy"
	}
	adminHost := valueOrDefault(strings.TrimSpace(data.AdminHostname), "admin.local")
	webmailHost := valueOrDefault(strings.TrimSpace(data.WebmailHostname), adminHost)
	davHost := valueOrDefault(strings.TrimSpace(data.DAVHostname), webmailHost)
	discoveryHosts := discoveryHostnames(data, adminHost, webmailHost, davHost)
	mailHost := mailServiceHostname(data.MailHostname, adminHost, webmailHost, davHost, discoveryHosts)
	certPath := strings.TrimSpace(data.CertPath)
	keyPath := strings.TrimSpace(data.KeyPath)
	if certPath == "" {
		certPath = "/etc/letsencrypt/live/" + adminHost + "/fullchain.pem"
	}
	if keyPath == "" {
		keyPath = "/etc/letsencrypt/live/" + adminHost + "/privkey.pem"
	}
	realIPHeader := "X-Forwarded-For"
	trustProxyHeaders := data.TrustProxyHeaders
	trustedProxyCIDRs := uniqueNonEmpty(data.TrustedProxyCIDRs)
	if data.CloudflareRealIPEnabled {
		realIPHeader = "CF-Connecting-IP"
		trustProxyHeaders = true
		trustedProxyCIDRs = append([]string(nil), cloudflareRealIPCIDRs...)
	}
	return nginxProxyTemplateData{
		TLSMode:              mode,
		AdminHostname:        adminHost,
		WebmailHostname:      webmailHost,
		DAVHostname:          davHost,
		MailServerName:       mailHost,
		DiscoveryServerNames: strings.Join(discoveryHosts, " "),
		SeparateDAVHost:      !strings.EqualFold(davHost, webmailHost),
		ACMEWebroot:          valueOrDefault(strings.TrimSpace(data.ACMEWebroot), "/var/lib/proidentity-mail/acme"),
		CertPath:             certPath,
		KeyPath:              keyPath,
		ForceHTTPS:           data.ForceHTTPS,
		TrustProxyHeaders:    trustProxyHeaders,
		TrustedProxyCIDRs:    trustedProxyCIDRs,
		RealIPHeader:         realIPHeader,
		TLSEnabled:           mode == "letsencrypt-http" || mode == "letsencrypt-dns-cloudflare" || mode == "custom-cert",
	}
}

func mailServiceHostname(hostname, adminHost, webmailHost, davHost string, discoveryHosts []string) string {
	normalized := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))
	if normalized == "" || !strings.Contains(normalized, ".") || strings.HasSuffix(normalized, ".local") {
		return ""
	}
	reserved := map[string]bool{
		strings.ToLower(strings.TrimSpace(adminHost)):   true,
		strings.ToLower(strings.TrimSpace(webmailHost)): true,
		strings.ToLower(strings.TrimSpace(davHost)):     true,
	}
	for _, host := range discoveryHosts {
		reserved[strings.ToLower(strings.TrimSpace(host))] = true
	}
	if reserved[normalized] {
		return ""
	}
	return normalized
}

func discoveryHostnames(data NginxProxyData, adminHost, webmailHost, davHost string) []string {
	reserved := map[string]bool{
		strings.ToLower(strings.TrimSpace(adminHost)):   true,
		strings.ToLower(strings.TrimSpace(webmailHost)): true,
		strings.ToLower(strings.TrimSpace(davHost)):     true,
	}
	seen := map[string]bool{}
	out := make([]string, 0, 2)
	for _, host := range []string{data.AutoconfigHostname, data.AutodiscoverHostname} {
		normalized := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(host, ".")))
		if normalized == "" || !strings.Contains(normalized, ".") || strings.HasSuffix(normalized, ".local") {
			continue
		}
		if reserved[normalized] || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func uniqueSNIHosts(hosts []MailServerSNIHost) []MailServerSNIHost {
	seen := make(map[string]bool)
	out := make([]MailServerSNIHost, 0, len(hosts))
	for _, host := range hosts {
		host.Hostname = strings.ToLower(strings.TrimSpace(host.Hostname))
		host.TLSCertFile = strings.TrimSpace(host.TLSCertFile)
		host.TLSKeyFile = strings.TrimSpace(host.TLSKeyFile)
		host.TLSChainFile = strings.TrimSpace(host.TLSChainFile)
		if host.Hostname == "" || seen[host.Hostname] {
			continue
		}
		seen[host.Hostname] = true
		out = append(out, host)
	}
	return out
}
