package webmail

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestMaildirStoreMarksMessageReadWhenOpened(t *testing.T) {
	root := t.TempDir()
	newDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	curDir := filepath.Join(root, "example.com", "marko", "Maildir", "cur")
	if err := os.MkdirAll(newDir, 0750); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	messageID := "message-1"
	raw := "From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Read me\r\n\r\nbody"
	if err := os.WriteFile(filepath.Join(newDir, messageID), []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	before, err := store.ListMessages(context.Background(), "marko@example.com", "inbox", 10)
	if err != nil {
		t.Fatalf("ListMessages before returned error: %v", err)
	}
	if len(before) != 1 || !before[0].Unread {
		t.Fatalf("before messages = %+v, want unread message", before)
	}

	message, err := store.MarkMessageRead(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("MarkMessageRead returned error: %v", err)
	}
	if message.Unread {
		t.Fatalf("message.Unread = true, want false")
	}
	wantID := "message-1:2,S"
	if runtime.GOOS == "windows" {
		wantID = "message-1"
	}
	if message.ID != wantID {
		t.Fatalf("message.ID = %q, want seen Maildir id %q", message.ID, wantID)
	}
	if _, err := os.Stat(filepath.Join(newDir, messageID)); !os.IsNotExist(err) {
		t.Fatalf("new message path still exists or unexpected stat error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(curDir, wantID)); err != nil {
		t.Fatalf("seen message missing in cur: %v", err)
	}

	byOldID, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage by old id returned error: %v", err)
	}
	if byOldID.ID != wantID || byOldID.Unread {
		t.Fatalf("GetMessage by old id = %+v, want seen message", byOldID)
	}
	after, err := store.ListFolders(context.Background(), "marko@example.com")
	if err != nil {
		t.Fatalf("ListFolders after returned error: %v", err)
	}
	if after[0].Unread != 0 {
		t.Fatalf("inbox unread = %d, want 0", after[0].Unread)
	}
}

func TestMaildirStoreDecodesMultipartMessageAndBlocksExternalSources(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	messageID := "mime-message"
	raw := "From: =?UTF-8?Q?Marcel_=C4=8C?= <marcel@mojemedia.eu>\r\n" +
		"To: marko@example.com\r\n" +
		"Subject: =?UTF-8?Q?Test_=C4=8D=2E2?=\r\n" +
		"Date: Sat, 09 May 2026 23:59:00 +0000\r\n" +
		"Content-Type: multipart/alternative; boundary=\"_c8cad9002f8e7eafaa644c763bec8429\"\r\n" +
		"\r\n" +
		"--_c8cad9002f8e7eafaa644c763bec8429\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"Content-Type: text/plain; charset=US-ASCII; format=flowed\r\n" +
		"\r\n" +
		"Hello there\r\n" +
		"--_c8cad9002f8e7eafaa644c763bec8429\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		"<html><head><meta http-equiv=3D\"Content-Type\" content=3D\"text/html; charset=3DUTF-8\" /></head><body><p>Hello there</p><img src=3D\"https://tracker.example/p.png\"><img src=3D\"https://mail.example.com/logo.png\"><script>alert(1)</script></body></html>\r\n" +
		"\r\n" +
		"--_c8cad9002f8e7eafaa644c763bec8429--\r\n"
	if err := os.WriteFile(filepath.Join(messageDir, messageID), []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	messages, err := store.ListMessages(context.Background(), "marko@example.com", "inbox", 10)
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Subject != "Test č.2" {
		t.Fatalf("summary subject = %q, want decoded subject", messages[0].Subject)
	}
	if messages[0].From != "Marcel Č <marcel@mojemedia.eu>" {
		t.Fatalf("summary from = %q, want decoded from", messages[0].From)
	}
	if messages[0].Preview != "Hello there" {
		t.Fatalf("summary preview = %q, want decoded body preview", messages[0].Preview)
	}

	message, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage returned error: %v", err)
	}
	if message.Subject != "Test č.2" {
		t.Fatalf("detail subject = %q, want decoded subject", message.Subject)
	}
	if message.Body != "Hello there" {
		t.Fatalf("body = %q, want decoded text part", message.Body)
	}
	if !strings.Contains(message.HTML, "<p>Hello there</p>") {
		t.Fatalf("html = %q, want decoded html body", message.HTML)
	}
	if strings.Contains(strings.ToLower(message.HTML), "script") {
		t.Fatalf("html = %q, should strip scripts", message.HTML)
	}
	if !strings.Contains(message.HTML, `data-external-src="https://tracker.example/p.png"`) {
		t.Fatalf("html = %q, want remote image blocked", message.HTML)
	}
	if strings.Contains(message.HTML, `<img src="https://tracker.example/p.png"`) {
		t.Fatalf("html = %q, remote image should not load by default", message.HTML)
	}
	if !strings.Contains(message.HTML, `src="https://mail.example.com/logo.png"`) {
		t.Fatalf("html = %q, local domain image should be trusted", message.HTML)
	}
	if message.ExternalSourceCount != 1 || !message.ExternalSourcesBlocked {
		t.Fatalf("external source state = %d/%v, want 1/true", message.ExternalSourceCount, message.ExternalSourcesBlocked)
	}
}

func TestSanitizeMailHTMLDropsUnsafeLinksAndNonAllowlistedTags(t *testing.T) {
	input := `<p onclick="alert(1)">Hi <a href="java&#x73;cript:alert(1)">bad</a> <a href="https://example.com/report">safe</a></p><marquee>flash</marquee><img src="https://tracker.example/p.png" onerror="alert(1)">`

	safe, blocked := sanitizeMailHTML(input, []string{"example.com"})

	if strings.Contains(strings.ToLower(safe), "javascript:") {
		t.Fatalf("sanitized html kept javascript URL: %s", safe)
	}
	if strings.Contains(strings.ToLower(safe), "onclick") || strings.Contains(strings.ToLower(safe), "onerror") {
		t.Fatalf("sanitized html kept event handler: %s", safe)
	}
	if strings.Contains(strings.ToLower(safe), "marquee") {
		t.Fatalf("sanitized html kept non-allowlisted tag: %s", safe)
	}
	if !strings.Contains(safe, `href="https://example.com/report"`) || !strings.Contains(safe, `>safe</a>`) {
		t.Fatalf("sanitized html removed safe https link: %s", safe)
	}
	if !strings.Contains(safe, `data-external-src="https://tracker.example/p.png"`) || blocked != 1 {
		t.Fatalf("external image blocking = html:%s blocked:%d, want one blocked remote image", safe, blocked)
	}
}

func TestSanitizeMailHTMLOmitsUnsafeImageSchemesInsteadOfDeferringThem(t *testing.T) {
	input := `<img src="javascript:alert(1)"><img src="data:image/svg+xml,<svg onload=alert(1)>"><img src="https://remote.example/pixel.png">`

	safe, blocked := sanitizeMailHTML(input, []string{"example.com"})

	if strings.Contains(strings.ToLower(safe), "javascript:") || strings.Contains(strings.ToLower(safe), "svg") {
		t.Fatalf("sanitized html preserved unsafe image source: %s", safe)
	}
	if strings.Contains(safe, `data-external-src="javascript:`) || strings.Contains(safe, `data-external-src="data:image/svg`) {
		t.Fatalf("unsafe image source was deferred for later loading: %s", safe)
	}
	if !strings.Contains(safe, `data-external-src="https://remote.example/pixel.png"`) || blocked != 1 {
		t.Fatalf("remote image blocking = html:%s blocked:%d, want one blocked http image", safe, blocked)
	}
}

func TestMaildirStoreParsesAuthenticationResults(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	messageID := "auth-message"
	raw := "From: sender@example.net\r\n" +
		"To: marko@example.com\r\n" +
		"Subject: Authenticated\r\n" +
		"Authentication-Results: mail.example.com; spf=pass smtp.mailfrom=example.net; dkim=pass header.d=example.net; dmarc=pass header.from=example.net\r\n" +
		"Received: from mail.example.net (mail.example.net [192.0.2.10]) by mail.example.com with ESMTPS id abc123\r\n" +
		"\r\n" +
		"trusted body"
	if err := os.WriteFile(filepath.Join(messageDir, messageID), []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	message, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage returned error: %v", err)
	}
	if message.Auth == nil {
		t.Fatal("Auth = nil, want parsed authentication status")
	}
	if message.Auth.SPF != "pass" || message.Auth.DKIM != "pass" || message.Auth.DMARC != "pass" {
		t.Fatalf("Auth = %+v, want SPF/DKIM/DMARC pass", message.Auth)
	}
	if message.Auth.TLS != "encrypted" {
		t.Fatalf("Auth.TLS = %q, want encrypted", message.Auth.TLS)
	}
	if !message.Auth.Trusted {
		t.Fatalf("Auth.Trusted = false, want true")
	}
}

func TestMaildirStoreDecodesCentralEuropeanCharsets(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	messageID := "latin2-message"
	raw := []byte("From: sender@example.net\r\nTo: marko@example.com\r\nSubject: =?iso-8859-2?Q?P=F8=EDli=B9_=BElu=BBou=E8k=FD?=\r\nContent-Type: text/plain; charset=iso-8859-2\r\n\r\n")
	raw = append(raw, []byte{'P', 0xF8, 0xED, 'l', 'i', 0xB9, ' ', 0xBE, 'l', 'u', 0xBB, 'o', 'u', 0xE8, 'k', 0xFD}...)
	if err := os.WriteFile(filepath.Join(messageDir, messageID), raw, 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	message, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage returned error: %v", err)
	}
	if message.Subject != "Příliš žluťoučký" {
		t.Fatalf("subject = %q, want decoded latin2 subject", message.Subject)
	}
	if message.Body != "Příliš žluťoučký" {
		t.Fatalf("body = %q, want decoded latin2 body", message.Body)
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
	var sent MailFolder
	for _, folder := range folders {
		if folder.ID == "sent" {
			sent = folder
		}
	}
	if sent.ID != "sent" || sent.Name != "Sent" || sent.Total != 1 || sent.Unread != 0 {
		t.Fatalf("sent folder = %+v, want sent total 1 unread 0", sent)
	}
	messages, err := store.ListMessages(context.Background(), "marko@example.com", "sent", 10)
	if err != nil {
		t.Fatalf("ListMessages sent returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].Subject != "Sent copy" || messages[0].Unread {
		t.Fatalf("sent messages = %+v, want one read sent copy", messages)
	}
}

func TestMaildirStoreIncludesDraftsFolderAndSavesDraftMessages(t *testing.T) {
	root := t.TempDir()
	store := MaildirStore{Root: root}
	message := OutboundMessage{
		From:     "marko@example.com",
		To:       []string{"ada@example.net"},
		Subject:  "Draft copy",
		Body:     "draft body",
		BodyHTML: "<p>draft body</p>",
	}
	id, err := store.SaveDraftMessage(context.Background(), message)
	if err != nil {
		t.Fatalf("SaveDraftMessage returned error: %v", err)
	}
	if id == "" {
		t.Fatal("draft id is empty")
	}
	folders, err := store.ListFolders(context.Background(), "marko@example.com")
	if err != nil {
		t.Fatalf("ListFolders returned error: %v", err)
	}
	var drafts MailFolder
	for _, folder := range folders {
		if folder.ID == "drafts" {
			drafts = folder
		}
	}
	if drafts.ID != "drafts" || drafts.Name != "Drafts" || drafts.Total != 1 || drafts.Unread != 0 {
		t.Fatalf("drafts folder = %+v, want drafts total 1 unread 0", drafts)
	}
	messages, err := store.ListMessages(context.Background(), "marko@example.com", "drafts", 10)
	if err != nil {
		t.Fatalf("ListMessages drafts returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].Subject != "Draft copy" || messages[0].Unread {
		t.Fatalf("draft messages = %+v, want one read draft copy", messages)
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

func TestMaildirStoreAddsTrashOriginWhenMovingToTrash(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", ".Sent", "cur")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir sent: %v", err)
	}
	messageID := "message-1"
	raw := "From: marko@example.com\r\nTo: sender@example.net\r\nSubject: Sent\r\n\r\nsent body"
	if err := os.WriteFile(filepath.Join(messageDir, messageID), []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	if err := store.MoveMessage(context.Background(), "marko@example.com", messageID, "trash"); err != nil {
		t.Fatalf("MoveMessage returned error: %v", err)
	}
	messages, err := store.ListMessages(context.Background(), "marko@example.com", "trash", 10)
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].TrashOrigin != "sent" {
		t.Fatalf("trash origin = %+v, want sent", messages)
	}
	detail, err := store.GetMessage(context.Background(), "marko@example.com", messageID)
	if err != nil {
		t.Fatalf("GetMessage returned error: %v", err)
	}
	if detail.TrashOrigin != "sent" {
		t.Fatalf("detail trash origin = %q, want sent", detail.TrashOrigin)
	}
}

func TestMaildirStoreDeletesMessagePermanently(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", ".Trash", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir trash: %v", err)
	}
	messageID := "message-1"
	raw := "From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Trash\r\n\r\ntrash body"
	messagePath := filepath.Join(messageDir, messageID)
	if err := os.WriteFile(messagePath, []byte(raw), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}

	store := MaildirStore{Root: root}
	if err := store.DeleteMessage(context.Background(), "marko@example.com", messageID); err != nil {
		t.Fatalf("DeleteMessage returned error: %v", err)
	}
	if _, err := os.Stat(messagePath); !os.IsNotExist(err) {
		t.Fatalf("message still exists or stat failed unexpectedly: %v", err)
	}
}
