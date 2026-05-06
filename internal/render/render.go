package render

import (
	"bytes"
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
