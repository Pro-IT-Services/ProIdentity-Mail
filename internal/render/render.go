package render

import (
	"bytes"
	"strings"
	"text/template"
)

type PostfixMainData struct {
	Hostname string
}

type DovecotSQLData struct {
	Database string
	User     string
	Password string
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

type rspamdTenantPolicyTemplateDomain struct {
	Domain          string
	RuleName        string
	MalwareRuleName string
	SymbolName      string
	SpamAction      string
	MalwareAction   string
}

type postfixMySQLTemplateData struct {
	Database string
	User     string
	Password string
	Query    string
}

func RenderPostfixMain(data PostfixMainData) ([]byte, error) {
	return renderTemplate("postfix-main", postfixMainTemplate, data)
}

func RenderPostfixMaster() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(postfixMasterTemplate), "\n")), nil
}

func RenderPostfixVirtualMailboxDomains(data PostfixMySQLData) ([]byte, error) {
	return renderPostfixMySQL(data, "SELECT 1 FROM domains WHERE name='%s' AND status IN ('pending','active')")
}

func RenderPostfixVirtualMailboxMaps(data PostfixMySQLData) ([]byte, error) {
	return renderPostfixMySQL(data, "SELECT CONCAT(d.name, '/', u.local_part, '/Maildir/') FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE CONCAT(u.local_part, '@', d.name)='%s' AND u.status='active' AND d.status IN ('pending','active')")
}

func RenderPostfixVirtualAliasMaps(data PostfixMySQLData) ([]byte, error) {
	return renderPostfixMySQL(data, "SELECT destination FROM aliases a JOIN domains d ON d.id = a.domain_id WHERE CONCAT(a.source_local_part, '@', d.name)='%s'")
}

func RenderDovecotSQL(data DovecotSQLData) ([]byte, error) {
	return renderTemplate("dovecot-sql", dovecotSQLTemplate, data)
}

func RenderDovecotLocal() ([]byte, error) {
	return []byte(bytes.TrimLeft([]byte(dovecotLocalTemplate), "\n")), nil
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
