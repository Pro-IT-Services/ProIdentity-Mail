package webmail

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
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
	client, err := smtp.NewClient(conn, strings.Split(addr, ":")[0])
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: "mail.local", InsecureSkipVerify: true}); err != nil {
			return err
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

func buildRFC822(message OutboundMessage) []byte {
	var buf bytes.Buffer
	now := time.Now().UTC().Format(time.RFC1123Z)
	fmt.Fprintf(&buf, "From: %s\r\n", sanitizeHeader(message.From))
	fmt.Fprintf(&buf, "To: %s\r\n", sanitizeHeader(strings.Join(message.To, ", ")))
	fmt.Fprintf(&buf, "Subject: %s\r\n", sanitizeHeader(message.Subject))
	fmt.Fprintf(&buf, "Date: %s\r\n", now)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(&buf, "\r\n%s\r\n", strings.ReplaceAll(message.Body, "\n", "\r\n"))
	return buf.Bytes()
}

func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}
