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
}
