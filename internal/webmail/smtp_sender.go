package webmail

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPSender struct {
	Addr string
}

func (s SMTPSender) Send(ctx context.Context, message OutboundMessage) error {
	addr := s.Addr
	if addr == "" {
		addr = "127.0.0.1:25"
	}
	data := buildRFC822(message)
	errCh := make(chan error, 1)
	go func() {
		errCh <- sendLocalSMTP(addr, message.From, message.To, data)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func sendLocalSMTP(addr, from string, to []string, data []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	host := smtpHost(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if tlsConfig := smtpTLSConfigForHost(host); tlsConfig != nil {
			if err := client.StartTLS(tlsConfig); err != nil {
				return err
			}
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func smtpHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		return host
	}
	return strings.Trim(addr, "[]")
}

func smtpTLSConfigForHost(host string) *tls.Config {
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" || strings.EqualFold(host, "localhost") {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}
	return &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
}

func buildRFC822(message OutboundMessage) []byte {
	var buf bytes.Buffer
	now := time.Now().UTC().Format(time.RFC1123Z)
	fmt.Fprintf(&buf, "From: %s\r\n", sanitizeHeader(message.From))
	fmt.Fprintf(&buf, "To: %s\r\n", sanitizeHeader(strings.Join(message.To, ", ")))
	fmt.Fprintf(&buf, "Subject: %s\r\n", sanitizeHeader(message.Subject))
	fmt.Fprintf(&buf, "Date: %s\r\n", now)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	if len(message.Attachments) > 0 {
		mixedBoundary := mimeBoundary("mixed", message)
		fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mixedBoundary)
		fmt.Fprintf(&buf, "\r\n--%s\r\n", mixedBoundary)
		writeMessageBodyPart(&buf, message)
		for _, attachment := range message.Attachments {
			writeAttachmentPart(&buf, mixedBoundary, attachment)
		}
		fmt.Fprintf(&buf, "--%s--\r\n", mixedBoundary)
		return buf.Bytes()
	}
	writeMessageBodyPart(&buf, message)
	return buf.Bytes()
}

func writeMessageBodyPart(buf *bytes.Buffer, message OutboundMessage) {
	if strings.TrimSpace(message.BodyHTML) == "" {
		fmt.Fprintf(buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(buf, "Content-Transfer-Encoding: 8bit\r\n")
		fmt.Fprintf(buf, "\r\n%s\r\n", normalizeCRLF(message.Body))
		return
	}
	alternativeBoundary := mimeBoundary("alternative", message)
	fmt.Fprintf(buf, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", alternativeBoundary)
	fmt.Fprintf(buf, "--%s\r\n", alternativeBoundary)
	fmt.Fprintf(buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(buf, "\r\n%s\r\n", normalizeCRLF(message.Body))
	fmt.Fprintf(buf, "--%s\r\n", alternativeBoundary)
	fmt.Fprintf(buf, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(buf, "\r\n%s\r\n", normalizeCRLF(message.BodyHTML))
	fmt.Fprintf(buf, "--%s--\r\n", alternativeBoundary)
}

func writeAttachmentPart(buf *bytes.Buffer, boundary string, attachment OutboundAttachment) {
	filename := sanitizeAttachmentFilename(attachment.Filename)
	contentType := strings.TrimSpace(attachment.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	fmt.Fprintf(buf, "Content-Type: %s\r\n", sanitizeHeader(contentType))
	fmt.Fprintf(buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", strings.ReplaceAll(mime.QEncoding.Encode("utf-8", filename), "\"", ""))
	fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n\r\n")
	writeBase64Lines(buf, attachment.Data)
	fmt.Fprintf(buf, "\r\n")
}

func writeBase64Lines(buf *bytes.Buffer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	for len(encoded) > 76 {
		fmt.Fprintf(buf, "%s\r\n", encoded[:76])
		encoded = encoded[76:]
	}
	if encoded != "" {
		fmt.Fprintf(buf, "%s\r\n", encoded)
	}
}

func mimeBoundary(prefix string, message OutboundMessage) string {
	seed := sanitizeHeader(message.From + "-" + message.Subject)
	seed = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, seed)
	if seed == "" {
		seed = "message"
	}
	if len(seed) > 24 {
		seed = seed[:24]
	}
	return "proidentity-" + prefix + "-" + seed
}

func normalizeCRLF(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", "\r\n")
}

func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	line, _, _ := strings.Cut(value, "\n")
	return strings.TrimSpace(line)
}
