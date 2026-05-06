package webmail

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMaildirStoreListsRecentMessages(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	oldPath := filepath.Join(messageDir, "old")
	newPath := filepath.Join(messageDir, "new")
	if err := os.WriteFile(oldPath, []byte("From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Old\r\nDate: Wed, 06 May 2026 10:00:00 +0000\r\n\r\nold body"), 0640); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("From: sender@example.net\r\nTo: marko@example.com\r\nSubject: New\r\nDate: Wed, 06 May 2026 11:00:00 +0000\r\n\r\nnew body"), 0640); err != nil {
		t.Fatalf("write new: %v", err)
	}
	oldTime := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 5, 6, 11, 0, 0, 0, time.UTC)
	_ = os.Chtimes(oldPath, oldTime, oldTime)
	_ = os.Chtimes(newPath, newTime, newTime)

	store := MaildirStore{Root: root}
	messages, err := store.ListRecentMessages(context.Background(), "marko@example.com", 2)
	if err != nil {
		t.Fatalf("ListRecentMessages returned error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(messages))
	}
	if messages[0].Subject != "New" || messages[1].Subject != "Old" {
		t.Fatalf("messages not newest first: %+v", messages)
	}
	if messages[0].Preview != "new body" {
		t.Fatalf("preview = %q, want body preview", messages[0].Preview)
	}
}

func TestMaildirStoreGetsMessageBodyByID(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	messageID := "message-1"
	raw := "From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Full Body\r\nDate: Wed, 06 May 2026 11:00:00 +0000\r\n\r\nline one\r\nline two\r\n"
	if err := os.WriteFile(filepath.Join(messageDir, messageID), []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	message, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage returned error: %v", err)
	}
	if message.Subject != "Full Body" {
		t.Fatalf("subject = %q, want Full Body", message.Subject)
	}
	if message.Body != "line one\r\nline two\r\n" {
		t.Fatalf("body = %q, want full body", message.Body)
	}
}

func TestMaildirStoreRejectsUnsafeMessageID(t *testing.T) {
	store := MaildirStore{Root: t.TempDir()}
	if _, err := store.GetMessage(context.Background(), "marko@example.com", "../secret"); err == nil {
		t.Fatal("expected unsafe message ID to fail")
	}
}
