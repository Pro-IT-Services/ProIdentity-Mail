package webmail

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildRFC822IncludesHTMLAndAttachments(t *testing.T) {
	raw := string(buildRFC822(OutboundMessage{
		From:     "marko@example.com",
		To:       []string{"ada@example.net"},
		Subject:  "Report",
		Body:     "Plain report",
		BodyHTML: "<p><strong>Plain</strong> report</p>",
		Attachments: []OutboundAttachment{{
			Filename:    "report.txt",
			ContentType: "text/plain; charset=utf-8",
			Data:        []byte("hello attachment"),
		}},
	}))

	for _, want := range []string{
		"Content-Type: multipart/mixed;",
		"Content-Type: multipart/alternative;",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Disposition: attachment; filename=\"report.txt\"",
		"Content-Transfer-Encoding: base64",
		"aGVsbG8gYXR0YWNobWVudA==",
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("message missing %q:\n%s", want, raw)
		}
	}
}

func TestParseOutboundRequestRejectsRecipientHeaderInjection(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/send", bytes.NewBufferString(`{"to":["ada@example.net\r\nBcc: attacker@example.net"],"subject":"Hello","body":"Hi"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	_, err := parseOutboundRequest(rec, req)

	if err == nil || !strings.Contains(err.Error(), "valid recipient") {
		t.Fatalf("error = %v, want valid recipient rejection", err)
	}
}

func TestBuildRFC822StripsHeaderNewlines(t *testing.T) {
	raw := string(buildRFC822(OutboundMessage{
		From:    "marko@example.com\r\nX-Injected: yes",
		To:      []string{"ada@example.net"},
		Subject: "Hello\r\nBcc: attacker@example.net",
		Body:    "Body",
	}))

	if strings.Contains(raw, "X-Injected: yes") || strings.Contains(raw, "Bcc: attacker@example.net") {
		t.Fatalf("message contains injected header:\n%s", raw)
	}
}

func TestSMTPStartTLSConfigNeverDisablesVerification(t *testing.T) {
	remote := smtpTLSConfigForHost("smtp.example.com")
	if remote == nil {
		t.Fatal("remote smtp host should use TLS verification")
	}
	if remote.InsecureSkipVerify {
		t.Fatal("remote smtp tls config must not disable verification")
	}
	if remote.MinVersion < 0x0303 {
		t.Fatalf("MinVersion = %x, want TLS 1.2+", remote.MinVersion)
	}
	if loopback := smtpTLSConfigForHost("127.0.0.1"); loopback != nil {
		t.Fatalf("loopback smtp config = %+v, want nil to avoid pointless local STARTTLS", loopback)
	}
}
