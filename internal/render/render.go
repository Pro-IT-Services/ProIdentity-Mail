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

func RenderPostfixMain(data PostfixMainData) ([]byte, error) {
	return renderTemplate("postfix-main", postfixMainTemplate, data)
}

func RenderDovecotSQL(data DovecotSQLData) ([]byte, error) {
	return renderTemplate("dovecot-sql", dovecotSQLTemplate, data)
}

func RenderRspamdLocal() ([]byte, error) {
	return []byte(rspamdLocalTemplate), nil
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
