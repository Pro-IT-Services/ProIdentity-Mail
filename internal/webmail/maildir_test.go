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

func TestMaildirStoreReportsUnreadCounts(t *testing.T) {
	root := t.TempDir()
	newDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	curDir := filepath.Join(root, "example.com", "marko", "Maildir", "cur")
	if err := os.MkdirAll(newDir, 0750); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	if err := os.MkdirAll(curDir, 0750); err != nil {
		t.Fatalf("mkdir cur: %v", err)
	}
	raw := []byte("From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Count\r\n\r\nbody")
	if err := os.WriteFile(filepath.Join(newDir, "unread"), raw, 0640); err != nil {
		t.Fatalf("write unread: %v", err)
	}
	if err := os.WriteFile(filepath.Join(curDir, "read-seen"), raw, 0640); err != nil {
		t.Fatalf("write read: %v", err)
	}

	store := MaildirStore{Root: root}
	folders, err := store.ListFolders(context.Background(), "marko@example.com")
	if err != nil {
		t.Fatalf("ListFolders returned error: %v", err)
	}
	if folders[0].ID != "inbox" || folders[0].Total != 2 || folders[0].Unread != 1 {
		t.Fatalf("inbox counts = %+v, want total 2 unread 1", folders[0])
	}
	messages, err := store.ListMessages(context.Background(), "marko@example.com", "inbox", 10)
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(messages))
	}
	foundUnread := false
	foundRead := false
	for _, message := range messages {
		if message.ID == "unread" && message.Unread {
			foundUnread = true
		}
		if message.ID == "read-seen" && !message.Unread {
			foundRead = true
		}
	}
	if !foundUnread || !foundRead {
		t.Fatalf("unexpected unread flags: %+v", messages)
	}
}

func TestMaildirStoreIncludesSentFolderAndSavesSentMessages(t *testing.T) {
	root := t.TempDir()
	store := MaildirStore{Root: root}
	message := OutboundMessage{
		From:    "marko@example.com",
		To:      []string{"ada@example.net"},
		Subject: "Sent copy",
		Body:    "hello",
	}
	if err := store.SaveSentMessage(context.Background(), message); err != nil {
		t.Fatalf("SaveSentMessage returned error: %v", err)
	}
	folders, err := store.ListFolders(context.Background(), "marko@example.com")
	if err != nil {
		t.Fatalf("ListFolders returned error: %v", err)
	}
	if folders[1].ID != "sent" || folders[1].Name != "Sent" || folders[1].Total != 1 || folders[1].Unread != 0 {
		t.Fatalf("sent folder = %+v, want sent total 1 unread 0", folders[1])
	}
	messages, err := store.ListMessages(context.Background(), "marko@example.com", "sent", 10)
	if err != nil {
		t.Fatalf("ListMessages sent returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].Subject != "Sent copy" || messages[0].Unread {
		t.Fatalf("sent messages = %+v, want one read sent copy", messages)
	}
}

func TestMaildirStoreRejectsUnsafeMessageID(t *testing.T) {
	store := MaildirStore{Root: t.TempDir()}
	if _, err := store.GetMessage(context.Background(), "marko@example.com", "../secret"); err == nil {
		t.Fatal("expected unsafe message ID to fail")
	}
}

func TestMaildirStoreMovesMessageToSpamFolder(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	messageID := "message-1"
	raw := "From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Spam\r\n\r\nspam body"
	if err := os.WriteFile(filepath.Join(messageDir, messageID), []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	if err := store.MoveMessage(context.Background(), "marko@example.com", messageID, "spam"); err != nil {
		t.Fatalf("MoveMessage returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(messageDir, messageID)); !os.IsNotExist(err) {
		t.Fatalf("original message still exists or stat failed unexpectedly: %v", err)
	}
	movedPath := filepath.Join(root, "example.com", "marko", "Maildir", ".Spam", "new", messageID)
	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("moved message missing: %v", err)
	}
	message, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage after move returned error: %v", err)
	}
	if message.Mailbox != ".Spam" {
		t.Fatalf("mailbox = %q, want .Spam", message.Mailbox)
	}
}
