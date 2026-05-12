package quarantine

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStoreStoresMessageWithHashAndSafePath(t *testing.T) {
	store := FileStore{Root: t.TempDir(), MailRoot: t.TempDir()}
	stored, err := store.StoreMessage(context.Background(), StoreRequest{
		TenantID:  42,
		Recipient: "Marko@Example.COM",
		MessageID: "../bad",
		Reader:    strings.NewReader("Subject: held\r\n\r\nbody"),
	})
	if err != nil {
		t.Fatalf("StoreMessage returned error: %v", err)
	}
	if stored.SizeBytes != int64(len("Subject: held\r\n\r\nbody")) {
		t.Fatalf("unexpected size: %d", stored.SizeBytes)
	}
	if stored.SHA256 == "" {
		t.Fatal("expected sha256")
	}
	if strings.Contains(stored.StoragePath, "..") || filepath.IsAbs(stored.StoragePath) {
		t.Fatalf("storage path is unsafe: %q", stored.StoragePath)
	}
	if _, err := os.Stat(filepath.Join(store.Root, stored.StoragePath)); err != nil {
		t.Fatalf("stored message missing: %v", err)
	}
}

func TestFileStoreReleaseCopiesHeldMessageIntoRecipientInbox(t *testing.T) {
	quarantineRoot := t.TempDir()
	mailRoot := t.TempDir()
	store := FileStore{Root: quarantineRoot, MailRoot: mailRoot}
	stored, err := store.StoreMessage(context.Background(), StoreRequest{
		TenantID:  42,
		Recipient: "marko@example.com",
		MessageID: "message-1",
		Reader:    strings.NewReader("From: sender@example.net\r\nSubject: held\r\n\r\nbody"),
	})
	if err != nil {
		t.Fatalf("StoreMessage returned error: %v", err)
	}
	released, err := store.Release(context.Background(), ReleaseRequest{
		Recipient:    "marko@example.com",
		MessageID:    "message-1",
		StoragePath:  stored.StoragePath,
		QuarantineID: 9,
	})
	if err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	wantDir := filepath.Join(mailRoot, "example.com", "marko", "Maildir", "new")
	if !strings.HasPrefix(released.MailboxPath, wantDir) {
		t.Fatalf("released outside inbox: %q", released.MailboxPath)
	}
	data, err := os.ReadFile(released.MailboxPath)
	if err != nil {
		t.Fatalf("read released message: %v", err)
	}
	if !strings.Contains(string(data), "Subject: held") {
		t.Fatalf("unexpected released message: %q", string(data))
	}
}

func TestFileStoreRejectsUnsafeReleasePath(t *testing.T) {
	store := FileStore{Root: t.TempDir(), MailRoot: t.TempDir()}
	_, err := store.Release(context.Background(), ReleaseRequest{
		Recipient:    "marko@example.com",
		StoragePath:  "../escape.eml",
		QuarantineID: 9,
	})
	if err == nil {
		t.Fatal("expected unsafe storage path to be rejected")
	}
}

func TestFileStoreReleaseCanDeliverThroughSMTP(t *testing.T) {
	addr, received, stop := startSMTPServer(t)
	defer stop()
	store := FileStore{Root: t.TempDir(), MailRoot: t.TempDir(), DeliveryAddr: addr}
	stored, err := store.StoreMessage(context.Background(), StoreRequest{
		TenantID:  42,
		Recipient: "marko@example.com",
		MessageID: "message-1",
		Reader:    strings.NewReader("From: sender@example.net\r\nSubject: held\r\n\r\nbody"),
	})
	if err != nil {
		t.Fatalf("StoreMessage returned error: %v", err)
	}
	released, err := store.Release(context.Background(), ReleaseRequest{
		Recipient:    "marko@example.com",
		MessageID:    "message-1",
		StoragePath:  stored.StoragePath,
		QuarantineID: 9,
	})
	if err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	if released.MailboxPath != "" {
		t.Fatalf("smtp release should not write Maildir directly: %q", released.MailboxPath)
	}
	if got := <-received; !strings.Contains(got, "Subject: held") {
		t.Fatalf("unexpected smtp payload: %q", got)
	}
}

func startSMTPServer(t *testing.T) (string, <-chan string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen smtp: %v", err)
	}
	received := make(chan string, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		writeLine := func(line string) {
			_, _ = conn.Write([]byte(line + "\r\n"))
		}
		writeLine("220 test")
		var data strings.Builder
		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					received <- data.String()
					writeLine("250 queued")
					inData = false
					continue
				}
				data.WriteString(line)
				data.WriteString("\n")
				continue
			}
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-test")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "MAIL FROM:"), strings.HasPrefix(upper, "RCPT TO:"):
				writeLine("250 OK")
			case strings.HasPrefix(upper, "DATA"):
				writeLine("354 end")
				inData = true
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 bye")
				return
			default:
				writeLine("250 OK")
			}
		}
	}()
	return listener.Addr().String(), received, func() {
		_ = listener.Close()
		<-done
	}
}
