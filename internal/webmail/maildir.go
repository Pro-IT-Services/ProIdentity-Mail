package webmail

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	stdhtml "html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	htmltoken "golang.org/x/net/html"
)

type MessageSummary struct {
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Subject     string    `json:"subject"`
	Date        time.Time `json:"date"`
	Preview     string    `json:"preview"`
	Mailbox     string    `json:"mailbox"`
	TrashOrigin string    `json:"trash_origin,omitempty"`
	SizeBytes   int64     `json:"size_bytes"`
	Unread      bool      `json:"unread"`
}

type MessageDetail struct {
	ID                     string             `json:"id"`
	From                   string             `json:"from"`
	To                     string             `json:"to"`
	Subject                string             `json:"subject"`
	Date                   time.Time          `json:"date"`
	Body                   string             `json:"body"`
	HTML                   string             `json:"html,omitempty"`
	ExternalSourceCount    int                `json:"external_source_count,omitempty"`
	ExternalSourcesBlocked bool               `json:"external_sources_blocked,omitempty"`
	Mailbox                string             `json:"mailbox"`
	TrashOrigin            string             `json:"trash_origin,omitempty"`
	SizeBytes              int64              `json:"size_bytes"`
	Unread                 bool               `json:"unread"`
	Auth                   *MessageAuthStatus `json:"auth,omitempty"`
}

type MessageAuthStatus struct {
	SPF     string `json:"spf,omitempty"`
	DKIM    string `json:"dkim,omitempty"`
	DMARC   string `json:"dmarc,omitempty"`
	TLS     string `json:"tls,omitempty"`
	Trusted bool   `json:"trusted"`
}

type MaildirStore struct {
	Root string
}

func (s MaildirStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	return s.ListMessages(ctx, email, "inbox", limit)
}

func (s MaildirStore) ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error) {
	if limit <= 0 || limit > 300 {
		limit = 300
	}
	local, domain, ok := strings.Cut(strings.ToLower(email), "@")
	if !ok || local == "" || domain == "" {
		return nil, errors.New("valid email is required")
	}
	root := filepath.Join(s.Root, domain, local, "Maildir")
	var paths []string
	for _, mailbox := range folderMailboxes(folder) {
		dir := filepath.Join(root, mailbox)
		entries, err := os.ReadDir(dir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Slice(paths, func(i, j int) bool {
		left, _ := os.Stat(paths[i])
		right, _ := os.Stat(paths[j])
		if left == nil || right == nil {
			return paths[i] > paths[j]
		}
		return left.ModTime().After(right.ModTime())
	})
	if len(paths) > limit {
		paths = paths[:limit]
	}
	messages := make([]MessageSummary, 0, len(paths))
	for _, path := range paths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		message, err := parseMessageSummary(path, root)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func folderMailboxes(folder string) []string {
	name := strings.ToLower(strings.TrimSpace(folder))
	switch name {
	case "drafts", "draft":
		return []string{filepath.Join(".Drafts", "new"), filepath.Join(".Drafts", "cur")}
	case "spam":
		return []string{filepath.Join(".Spam", "new"), filepath.Join(".Spam", "cur")}
	case "trash":
		return []string{filepath.Join(".Trash", "new"), filepath.Join(".Trash", "cur")}
	case "archive":
		return []string{filepath.Join(".Archive", "new"), filepath.Join(".Archive", "cur")}
	case "sent":
		return []string{filepath.Join(".Sent", "new"), filepath.Join(".Sent", "cur")}
	case "all":
		return []string{"new", "cur", filepath.Join(".Drafts", "new"), filepath.Join(".Drafts", "cur"), filepath.Join(".Sent", "new"), filepath.Join(".Sent", "cur"), filepath.Join(".Spam", "new"), filepath.Join(".Spam", "cur"), filepath.Join(".Trash", "new"), filepath.Join(".Trash", "cur"), filepath.Join(".Archive", "new"), filepath.Join(".Archive", "cur")}
	default:
		if mailbox, err := customMailboxName(folder); err == nil && mailbox != "" {
			return []string{filepath.Join(mailbox, "new"), filepath.Join(mailbox, "cur")}
		}
		return []string{"new", "cur"}
	}
}

func (s MaildirStore) ListFolders(ctx context.Context, email string) ([]MailFolder, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return nil, err
	}
	if err := ensureMaildir(root); err != nil {
		return nil, err
	}
	folders := []MailFolder{
		{ID: "inbox", Name: "Inbox", System: true},
		{ID: "drafts", Name: "Drafts", System: true},
		{ID: "sent", Name: "Sent", System: true},
		{ID: "archive", Name: "Archive", System: true},
		{ID: "spam", Name: "Spam", System: true},
		{ID: "trash", Name: "Trash", System: true},
	}
	for i := range folders {
		total, unread, err := countFolderMessages(root, folders[i].ID)
		if err != nil {
			return nil, err
		}
		folders[i].Total = total
		folders[i].Unread = unread
	}
	entries, err := os.ReadDir(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		id := strings.TrimPrefix(entry.Name(), ".")
		if _, ok := systemFolderID(id); ok {
			continue
		}
		total, unread, err := countFolderMessages(root, id)
		if err != nil {
			return nil, err
		}
		folders = append(folders, MailFolder{ID: id, Name: id, System: false, Total: total, Unread: unread})
	}
	systemCount := 6
	sort.SliceStable(folders[systemCount:], func(i, j int) bool {
		return strings.ToLower(folders[systemCount+i].Name) < strings.ToLower(folders[systemCount+j].Name)
	})
	return folders, nil
}

func (s MaildirStore) CreateFolder(ctx context.Context, email, name string) (MailFolder, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return MailFolder{}, err
	}
	mailbox, err := customMailboxName(name)
	if err != nil {
		return MailFolder{}, err
	}
	if err := os.MkdirAll(filepath.Join(root, mailbox, "cur"), 0750); err != nil {
		return MailFolder{}, err
	}
	if err := os.MkdirAll(filepath.Join(root, mailbox, "new"), 0750); err != nil {
		return MailFolder{}, err
	}
	if err := os.MkdirAll(filepath.Join(root, mailbox, "tmp"), 0750); err != nil {
		return MailFolder{}, err
	}
	id := strings.TrimPrefix(mailbox, ".")
	return MailFolder{ID: id, Name: id, System: false}, nil
}

func (s MaildirStore) DeleteFolder(ctx context.Context, email, name string) error {
	root, err := s.maildirRoot(email)
	if err != nil {
		return err
	}
	mailbox, err := customMailboxName(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(root, mailbox))
}

func parseMessageSummary(path, maildirRoot string) (MessageSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MessageSummary{}, err
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return MessageSummary{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return MessageSummary{}, err
	}
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return MessageSummary{}, err
	}
	content := parseMailContent(msg.Header, body, trustedDomainsForMaildirRoot(maildirRoot))
	date, _ := msg.Header.Date()
	rel, _ := filepath.Rel(maildirRoot, path)
	mailbox := strings.Split(filepath.ToSlash(rel), "/")[0]
	return MessageSummary{
		ID:          filepath.Base(path),
		From:        decodeMailHeader(msg.Header.Get("From")),
		To:          decodeMailHeader(msg.Header.Get("To")),
		Subject:     decodeMailHeader(msg.Header.Get("Subject")),
		Date:        date,
		Preview:     previewText(content.Text),
		Mailbox:     mailbox,
		TrashOrigin: normalizeTrashOrigin(msg.Header.Get("X-ProIdentity-Trash-Origin")),
		SizeBytes:   info.Size(),
		Unread:      isUnreadPath(path),
	}, nil
}

func (s MaildirStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return MessageDetail{}, err
	}
	path, err := s.messagePath(ctx, root, id)
	if err != nil {
		return MessageDetail{}, err
	}
	return parseMessageDetail(path, root)
}

func (s MaildirStore) MarkMessageRead(ctx context.Context, email, id string) (MessageDetail, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return MessageDetail{}, err
	}
	path, err := s.messagePath(ctx, root, id)
	if err != nil {
		return MessageDetail{}, err
	}
	path, err = markMessageSeen(path)
	if err != nil {
		return MessageDetail{}, err
	}
	return parseMessageDetail(path, root)
}

func (s MaildirStore) MoveMessage(ctx context.Context, email, id, folder string) error {
	root, err := s.maildirRoot(email)
	if err != nil {
		return err
	}
	source, err := s.messagePath(ctx, root, id)
	if err != nil {
		return err
	}
	destinationMailbox, err := destinationMailbox(folder)
	if err != nil {
		return err
	}
	destinationDir := filepath.Join(root, destinationMailbox, "new")
	if destinationMailbox == "new" {
		destinationDir = filepath.Join(root, "new")
	}
	if err := os.MkdirAll(destinationDir, 0750); err != nil {
		return err
	}
	if destinationMailbox == ".Trash" {
		origin := mailboxIDFromPath(root, source)
		if origin != "" && origin != "trash" {
			if err := addTrashOriginHeader(source, origin); err != nil {
				return err
			}
		}
	}
	return os.Rename(source, filepath.Join(destinationDir, filepath.Base(source)))
}

func (s MaildirStore) DeleteMessage(ctx context.Context, email, id string) error {
	root, err := s.maildirRoot(email)
	if err != nil {
		return err
	}
	source, err := s.messagePath(ctx, root, id)
	if err != nil {
		return err
	}
	return os.Remove(source)
}

func (s MaildirStore) SaveSentMessage(ctx context.Context, message OutboundMessage) error {
	root, err := s.maildirRoot(message.From)
	if err != nil {
		return err
	}
	destinationDir := filepath.Join(root, ".Sent", "cur")
	if err := os.MkdirAll(destinationDir, 0750); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	name := fmt.Sprintf("%d.%d.webmail.S", time.Now().UnixNano(), os.Getpid())
	return os.WriteFile(filepath.Join(destinationDir, name), buildRFC822(message), 0640)
}

func (s MaildirStore) SaveDraftMessage(ctx context.Context, message OutboundMessage) (string, error) {
	root, err := s.maildirRoot(message.From)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(message.DraftID) != "" {
		_ = s.DeleteMessage(ctx, message.From, message.DraftID)
	}
	destinationDir := filepath.Join(root, ".Drafts", "cur")
	if err := os.MkdirAll(destinationDir, 0750); err != nil {
		return "", err
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	name := maildirSeenName(fmt.Sprintf("%d.%d.webmail-draft", time.Now().UnixNano(), os.Getpid()))
	path := filepath.Join(destinationDir, name)
	if err := os.WriteFile(path, buildRFC822(message), 0640); err != nil {
		return "", err
	}
	return filepath.Base(path), nil
}

func (s MaildirStore) MessagePath(ctx context.Context, email, id string) (string, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return "", err
	}
	return s.messagePath(ctx, root, id)
}

func (s MaildirStore) maildirRoot(email string) (string, error) {
	local, domain, ok := strings.Cut(strings.ToLower(email), "@")
	if !ok || local == "" || domain == "" {
		return "", errors.New("valid email is required")
	}
	return filepath.Join(s.Root, domain, local, "Maildir"), nil
}

func (s MaildirStore) messagePath(ctx context.Context, root, id string) (string, error) {
	if strings.Contains(id, "/") || strings.Contains(id, `\`) || strings.Contains(id, "..") || id == "" {
		return "", errors.New("invalid message id")
	}
	mailboxes := []string{"new", "cur", ".Drafts/new", ".Drafts/cur", ".Sent/new", ".Sent/cur", ".Spam/new", ".Spam/cur", ".Trash/new", ".Trash/cur", ".Archive/new", ".Archive/cur"}
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
			mailboxes = append(mailboxes, filepath.ToSlash(filepath.Join(entry.Name(), "new")), filepath.ToSlash(filepath.Join(entry.Name(), "cur")))
		}
	}
	for _, mailbox := range mailboxes {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		path := filepath.Join(root, filepath.FromSlash(mailbox), id)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	if !strings.Contains(id, ":2,") {
		for _, mailbox := range mailboxes {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}
			dir := filepath.Join(root, filepath.FromSlash(mailbox))
			entries, err := os.ReadDir(dir)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return "", err
			}
			for _, entry := range entries {
				if !entry.Type().IsRegular() {
					continue
				}
				if maildirBaseName(entry.Name()) == id {
					return filepath.Join(dir, entry.Name()), nil
				}
			}
		}
	}
	return "", os.ErrNotExist
}

func destinationMailbox(folder string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(folder))
	switch name {
	case "inbox", "":
		return "new", nil
	case "drafts", "draft":
		return ".Drafts", nil
	case "spam":
		return ".Spam", nil
	case "trash":
		return ".Trash", nil
	case "archive":
		return ".Archive", nil
	case "sent":
		return ".Sent", nil
	default:
		return customMailboxName(folder)
	}
}

func ensureMaildir(root string) error {
	for _, dir := range []string{"cur", "new", "tmp"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0750); err != nil {
			return err
		}
	}
	return nil
}

func customMailboxName(name string) (string, error) {
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return "", errors.New("folder name is required")
	}
	if _, ok := systemFolderID(cleaned); ok {
		return "", errors.New("system folder already exists")
	}
	if strings.ContainsAny(cleaned, `/\:`) || strings.Contains(cleaned, "..") {
		return "", errors.New("folder name contains unsupported characters")
	}
	cleaned = strings.TrimPrefix(cleaned, ".")
	if cleaned == "" {
		return "", errors.New("folder name is required")
	}
	return "." + cleaned, nil
}

func systemFolderID(name string) (string, bool) {
	switch strings.ToLower(strings.TrimPrefix(strings.TrimSpace(name), ".")) {
	case "inbox":
		return "inbox", true
	case "drafts", "draft":
		return "drafts", true
	case "spam":
		return "spam", true
	case "sent":
		return "sent", true
	case "trash":
		return "trash", true
	case "archive":
		return "archive", true
	default:
		return "", false
	}
}

func countFolderMessages(root, folder string) (int, int, error) {
	count := 0
	unread := 0
	for _, mailbox := range folderMailboxes(folder) {
		entries, err := os.ReadDir(filepath.Join(root, mailbox))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return 0, 0, fmt.Errorf("count folder %s: %w", folder, err)
		}
		for _, entry := range entries {
			if entry.Type().IsRegular() {
				count++
				if isUnreadMaildirEntry(mailbox, entry.Name()) {
					unread++
				}
			}
		}
	}
	return count, unread, nil
}

func parseMessageDetail(path, maildirRoot string) (MessageDetail, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MessageDetail{}, err
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return MessageDetail{}, err
	}
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return MessageDetail{}, err
	}
	content := parseMailContent(msg.Header, body, trustedDomainsForMaildirRoot(maildirRoot))
	info, err := os.Stat(path)
	if err != nil {
		return MessageDetail{}, err
	}
	date, _ := msg.Header.Date()
	rel, _ := filepath.Rel(maildirRoot, path)
	mailbox := strings.Split(filepath.ToSlash(rel), "/")[0]
	return MessageDetail{
		ID:                     filepath.Base(path),
		From:                   decodeMailHeader(msg.Header.Get("From")),
		To:                     decodeMailHeader(msg.Header.Get("To")),
		Subject:                decodeMailHeader(msg.Header.Get("Subject")),
		Date:                   date,
		Body:                   content.Text,
		HTML:                   content.HTML,
		ExternalSourceCount:    content.ExternalSourceCount,
		ExternalSourcesBlocked: content.ExternalSourceCount > 0,
		Mailbox:                mailbox,
		TrashOrigin:            normalizeTrashOrigin(msg.Header.Get("X-ProIdentity-Trash-Origin")),
		SizeBytes:              info.Size(),
		Unread:                 isUnreadPath(path),
		Auth:                   parseMessageAuth(msg.Header),
	}, nil
}

func parseMessageAuth(header mail.Header) *MessageAuthStatus {
	auth := &MessageAuthStatus{}
	authResults := strings.ToLower(strings.Join(header["Authentication-Results"], " ; "))
	if authResults == "" {
		authResults = strings.ToLower(header.Get("Authentication-Results"))
	}
	auth.SPF = authResultValue(authResults, "spf")
	auth.DKIM = authResultValue(authResults, "dkim")
	auth.DMARC = authResultValue(authResults, "dmarc")
	if auth.SPF == "" {
		auth.SPF = firstTokenLower(header.Get("Received-SPF"))
	}
	auth.TLS = tlsStatusFromHeaders(header)
	auth.Trusted = auth.SPF == "pass" || auth.DKIM == "pass" || auth.DMARC == "pass"
	if auth.SPF == "" && auth.DKIM == "" && auth.DMARC == "" && auth.TLS == "" {
		return nil
	}
	return auth
}

func authResultValue(value, key string) string {
	if value == "" {
		return ""
	}
	re := regexp.MustCompile(`(?:^|[;\s])` + regexp.QuoteMeta(strings.ToLower(key)) + `\s*=\s*([a-z0-9_-]+)`)
	match := re.FindStringSubmatch(value)
	if len(match) < 2 {
		return ""
	}
	return normalizeAuthResult(match[1])
}

func firstTokenLower(value string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	if len(fields) == 0 {
		return ""
	}
	return normalizeAuthResult(strings.Trim(fields[0], ";"))
}

func normalizeAuthResult(value string) string {
	value = strings.Trim(strings.ToLower(value), " ;,.")
	switch value {
	case "pass", "fail", "softfail", "neutral", "none", "temperror", "permerror", "policy":
		return value
	default:
		return value
	}
}

func tlsStatusFromHeaders(header mail.Header) string {
	for _, name := range []string{"X-ProIdentity-TLS", "X-TLS", "X-Postfix-TLS"} {
		value := strings.ToLower(header.Get(name))
		if strings.Contains(value, "encrypt") || strings.Contains(value, "tls") {
			return "encrypted"
		}
	}
	for _, received := range header["Received"] {
		lower := strings.ToLower(received)
		if strings.Contains(lower, "with esmtps") || strings.Contains(lower, "with esmtpa") || strings.Contains(lower, "tls") {
			return "encrypted"
		}
	}
	return ""
}

func mailboxIDFromPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "new", "cur":
		return "inbox"
	case ".Trash":
		return "trash"
	case ".Sent":
		return "sent"
	case ".Spam":
		return "spam"
	case ".Archive":
		return "archive"
	default:
		return normalizeTrashOrigin(parts[0])
	}
}

func addTrashOriginHeader(path, origin string) error {
	origin = normalizeTrashOrigin(origin)
	if origin == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if bytes.Contains(bytes.ToLower(data), []byte("\nx-proidentity-trash-origin:")) || bytes.HasPrefix(bytes.ToLower(data), []byte("x-proidentity-trash-origin:")) {
		return nil
	}
	header := []byte("X-ProIdentity-Trash-Origin: " + origin + "\r\n")
	if index := bytes.Index(data, []byte("\r\n\r\n")); index >= 0 {
		out := make([]byte, 0, len(data)+len(header))
		out = append(out, data[:index+2]...)
		out = append(out, header...)
		out = append(out, data[index+2:]...)
		return os.WriteFile(path, out, 0640)
	}
	if index := bytes.Index(data, []byte("\n\n")); index >= 0 {
		out := make([]byte, 0, len(data)+len(header))
		out = append(out, data[:index+1]...)
		out = append(out, header...)
		out = append(out, data[index+1:]...)
		return os.WriteFile(path, out, 0640)
	}
	out := append(header, data...)
	return os.WriteFile(path, out, 0640)
}

func markMessageSeen(path string) (string, error) {
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "new" {
		return path, nil
	}
	parent := filepath.Dir(dir)
	targetDir := filepath.Join(parent, "cur")
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return "", err
	}
	target := filepath.Join(targetDir, maildirSeenName(filepath.Base(path)))
	if target == path {
		return path, nil
	}
	if err := os.Rename(path, target); err != nil {
		return "", err
	}
	return target, nil
}

func maildirSeenName(name string) string {
	if runtime.GOOS == "windows" {
		return name
	}
	base, flags, ok := strings.Cut(name, ":2,")
	if !ok {
		return name + ":2,S"
	}
	if strings.Contains(flags, "S") {
		return name
	}
	return base + ":2," + flags + "S"
}

func maildirBaseName(name string) string {
	base, _, ok := strings.Cut(name, ":2,")
	if !ok {
		return name
	}
	return base
}

func normalizeTrashOrigin(value string) string {
	cleaned := strings.TrimPrefix(strings.TrimSpace(value), ".")
	if cleaned == "" || strings.ContainsAny(cleaned, "\r\n/\\:") || strings.Contains(cleaned, "..") {
		return ""
	}
	switch strings.ToLower(cleaned) {
	case "new", "cur":
		return "inbox"
	case "sent":
		return "sent"
	case "trash":
		return "trash"
	case "spam":
		return "spam"
	case "archive":
		return "archive"
	default:
		return cleaned
	}
}

func isUnreadPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "new" {
			return true
		}
	}
	return false
}

func isUnreadMaildirEntry(mailbox, name string) bool {
	dir := filepath.Base(filepath.ToSlash(mailbox))
	if dir == "new" {
		return true
	}
	return false
}

func preview(reader io.Reader) string {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 4096), 64*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if len(line) > 160 {
				return line[:160]
			}
			return line
		}
	}
	return ""
}

func previewText(text string) string {
	return preview(strings.NewReader(strings.TrimSpace(text)))
}

type mailHeader interface {
	Get(string) string
}

type parsedMailContent struct {
	Text                string
	HTML                string
	ExternalSourceCount int
}

func parseMailContent(header mailHeader, body []byte, trustedDomains []string) parsedMailContent {
	var content parsedMailContent
	collectMailContent(header, body, trustedDomains, &content)
	if strings.TrimSpace(content.Text) == "" && strings.TrimSpace(content.HTML) != "" {
		content.Text = htmlToText(content.HTML)
	}
	if strings.TrimSpace(content.Text) == "" {
		content.Text = string(body)
	}
	return content
}

func collectMailContent(header mailHeader, body []byte, trustedDomains []string, content *parsedMailContent) {
	mediaType, params := parseContentType(header.Get("Content-Type"))
	if disposition, _, err := mime.ParseMediaType(header.Get("Content-Disposition")); err == nil && strings.EqualFold(disposition, "attachment") {
		return
	}
	if strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return
		}
		reader := multipart.NewReader(bytes.NewReader(body), boundary)
		for {
			part, err := reader.NextPart()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				return
			}
			partBody, err := io.ReadAll(part)
			_ = part.Close()
			if err != nil {
				continue
			}
			collectMailContent(part.Header, partBody, trustedDomains, content)
		}
	}

	decoded, err := decodeTransferEncoding(body, header.Get("Content-Transfer-Encoding"))
	if err != nil {
		decoded = body
	}
	text := decodeBytesCharset(decoded, params["charset"])
	switch strings.ToLower(mediaType) {
	case "text/html":
		safe, blocked := sanitizeMailHTML(text, trustedDomains)
		if strings.TrimSpace(content.HTML) == "" {
			content.HTML = safe
		}
		content.ExternalSourceCount += blocked
		if strings.TrimSpace(content.Text) == "" {
			content.Text = htmlToText(text)
		}
	case "text/plain", "":
		if strings.TrimSpace(content.Text) == "" {
			content.Text = text
		}
	}
}

func parseContentType(value string) (string, map[string]string) {
	mediaType, params, err := mime.ParseMediaType(value)
	if err != nil || mediaType == "" {
		return "text/plain", map[string]string{}
	}
	return strings.ToLower(mediaType), params
}

func decodeTransferEncoding(body []byte, encoding string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		return io.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(body)))
	case "quoted-printable":
		return io.ReadAll(quotedprintable.NewReader(bytes.NewReader(body)))
	default:
		return body, nil
	}
}

var wordDecoder = mime.WordDecoder{CharsetReader: charsetReader}

func decodeMailHeader(value string) string {
	decoded, err := wordDecoder.DecodeHeader(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(decoded)
}

func decodeBytesCharset(body []byte, charset string) string {
	charset = normalizeCharset(charset)
	if charset == "" || charset == "utf-8" || charset == "us-ascii" {
		if utf8.Valid(body) {
			return string(body)
		}
		return string([]rune(string(body)))
	}
	reader, err := charsetReader(charset, bytes.NewReader(body))
	if err != nil {
		return string(body)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return string(body)
	}
	return string(decoded)
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	switch normalizeCharset(charset) {
	case "", "utf-8", "us-ascii":
		return bytes.NewReader(data), nil
	case "iso-8859-1":
		return strings.NewReader(decodeSingleByte(data, iso88591Rune)), nil
	case "iso-8859-2":
		return strings.NewReader(decodeSingleByte(data, iso88592Rune)), nil
	case "windows-1250":
		return strings.NewReader(decodeSingleByte(data, windows1250Rune)), nil
	case "windows-1252":
		return strings.NewReader(decodeSingleByte(data, windows1252Rune)), nil
	default:
		return bytes.NewReader(data), nil
	}
}

func normalizeCharset(charset string) string {
	charset = strings.ToLower(strings.TrimSpace(strings.Trim(charset, `"`)))
	charset = strings.ReplaceAll(charset, "_", "-")
	switch charset {
	case "latin1", "latin-1", "iso8859-1":
		return "iso-8859-1"
	case "latin2", "latin-2", "iso8859-2":
		return "iso-8859-2"
	case "cp1250", "windows1250":
		return "windows-1250"
	case "cp1252", "windows1252":
		return "windows-1252"
	default:
		return charset
	}
}

func decodeSingleByte(data []byte, table func(byte) rune) string {
	var builder strings.Builder
	for _, b := range data {
		builder.WriteRune(table(b))
	}
	return builder.String()
}

func iso88591Rune(b byte) rune {
	return rune(b)
}

func iso88592Rune(b byte) rune {
	if b < 0xA0 {
		return rune(b)
	}
	table := [...]rune{
		0x00A0, 0x0104, 0x02D8, 0x0141, 0x00A4, 0x013D, 0x015A, 0x00A7,
		0x00A8, 0x0160, 0x015E, 0x0164, 0x0179, 0x00AD, 0x017D, 0x017B,
		0x00B0, 0x0105, 0x02DB, 0x0142, 0x00B4, 0x013E, 0x015B, 0x02C7,
		0x00B8, 0x0161, 0x015F, 0x0165, 0x017A, 0x02DD, 0x017E, 0x017C,
		0x0154, 0x00C1, 0x00C2, 0x0102, 0x00C4, 0x0139, 0x0106, 0x00C7,
		0x010C, 0x00C9, 0x0118, 0x00CB, 0x011A, 0x00CD, 0x00CE, 0x010E,
		0x0110, 0x0143, 0x0147, 0x00D3, 0x00D4, 0x0150, 0x00D6, 0x00D7,
		0x0158, 0x016E, 0x00DA, 0x0170, 0x00DC, 0x00DD, 0x0162, 0x00DF,
		0x0155, 0x00E1, 0x00E2, 0x0103, 0x00E4, 0x013A, 0x0107, 0x00E7,
		0x010D, 0x00E9, 0x0119, 0x00EB, 0x011B, 0x00ED, 0x00EE, 0x010F,
		0x0111, 0x0144, 0x0148, 0x00F3, 0x00F4, 0x0151, 0x00F6, 0x00F7,
		0x0159, 0x016F, 0x00FA, 0x0171, 0x00FC, 0x00FD, 0x0163, 0x02D9,
	}
	return table[b-0xA0]
}

func windows1250Rune(b byte) rune {
	if b < 0x80 {
		return rune(b)
	}
	table := [...]rune{
		0x20AC, 0x0081, 0x201A, 0x0083, 0x201E, 0x2026, 0x2020, 0x2021,
		0x0088, 0x2030, 0x0160, 0x2039, 0x015A, 0x0164, 0x017D, 0x0179,
		0x0090, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
		0x0098, 0x2122, 0x0161, 0x203A, 0x015B, 0x0165, 0x017E, 0x017A,
		0x00A0, 0x02C7, 0x02D8, 0x0141, 0x00A4, 0x0104, 0x00A6, 0x00A7,
		0x00A8, 0x00A9, 0x015E, 0x00AB, 0x00AC, 0x00AD, 0x00AE, 0x017B,
		0x00B0, 0x00B1, 0x02DB, 0x0142, 0x00B4, 0x00B5, 0x00B6, 0x00B7,
		0x00B8, 0x0105, 0x015F, 0x00BB, 0x013D, 0x02DD, 0x013E, 0x017C,
		0x0154, 0x00C1, 0x00C2, 0x0102, 0x00C4, 0x0139, 0x0106, 0x00C7,
		0x010C, 0x00C9, 0x0118, 0x00CB, 0x011A, 0x00CD, 0x00CE, 0x010E,
		0x0110, 0x0143, 0x0147, 0x00D3, 0x00D4, 0x0150, 0x00D6, 0x00D7,
		0x0158, 0x016E, 0x00DA, 0x0170, 0x00DC, 0x00DD, 0x0162, 0x00DF,
		0x0155, 0x00E1, 0x00E2, 0x0103, 0x00E4, 0x013A, 0x0107, 0x00E7,
		0x010D, 0x00E9, 0x0119, 0x00EB, 0x011B, 0x00ED, 0x00EE, 0x010F,
		0x0111, 0x0144, 0x0148, 0x00F3, 0x00F4, 0x0151, 0x00F6, 0x00F7,
		0x0159, 0x016F, 0x00FA, 0x0171, 0x00FC, 0x00FD, 0x0163, 0x02D9,
	}
	return table[b-0x80]
}

func windows1252Rune(b byte) rune {
	if b < 0x80 || b >= 0xA0 {
		return rune(b)
	}
	table := [...]rune{
		0x20AC, 0x0081, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
		0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x008D, 0x017D, 0x008F,
		0x0090, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
		0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x009D, 0x017E, 0x0178,
	}
	return table[b-0x80]
}

var (
	bodyTagRe         = regexp.MustCompile(`(?is)<\s*body[^>]*>(.*)<\s*/\s*body\s*>`)
	dangerousBlockRe  = regexp.MustCompile(`(?is)<\s*(script|style|iframe|object|embed|form|textarea|select|svg|math)[^>]*>.*?<\s*/\s*(script|style|iframe|object|embed|form|textarea|select|svg|math)\s*>`)
	dangerousSingleRe = regexp.MustCompile(`(?is)<\s*/?\s*(script|style|iframe|object|embed|link|meta|base|form|input|button|textarea|select|svg|math)[^>]*>`)
	eventAttributeRe  = regexp.MustCompile(`(?is)\s+on[a-z0-9_-]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	styleAttributeRe  = regexp.MustCompile(`(?is)\s+style\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	sourceAttributeRe = regexp.MustCompile(`(?is)\s(src|poster|background)\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	srcsetAttributeRe = regexp.MustCompile(`(?is)\ssrcset\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	breakTagRe        = regexp.MustCompile(`(?i)<\s*br\s*/?\s*>`)
	blockEndTagRe     = regexp.MustCompile(`(?i)</\s*(p|div|tr|li|h[1-6]|section|article|table)\s*>`)
	htmlTagRe         = regexp.MustCompile(`(?is)<[^>]+>`)
	whitespaceLineRe  = regexp.MustCompile(`[ \t]+`)
	blankLineRe       = regexp.MustCompile(`\n{3,}`)
)

func sanitizeMailHTML(markup string, trustedDomains []string) (string, int) {
	tokenizer := htmltoken.NewTokenizer(strings.NewReader(markup))
	var out strings.Builder
	blocked := 0
	skipDepth := 0
	for {
		kind := tokenizer.Next()
		switch kind {
		case htmltoken.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return strings.TrimSpace(out.String()), blocked
			}
			return strings.TrimSpace(out.String()), blocked
		case htmltoken.TextToken:
			if skipDepth == 0 {
				out.WriteString(stdhtml.EscapeString(string(tokenizer.Text())))
			}
		case htmltoken.StartTagToken, htmltoken.SelfClosingTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)
			if skipDepth > 0 {
				if skipContentTags[tag] {
					skipDepth++
				}
				continue
			}
			if skipContentTags[tag] {
				skipDepth = 1
				continue
			}
			if !allowedMailHTMLTags[tag] {
				continue
			}
			out.WriteByte('<')
			out.WriteString(tag)
			blocked += writeSanitizedAttributes(&out, tag, token.Attr, trustedDomains)
			if kind == htmltoken.SelfClosingTagToken && voidMailHTMLTags[tag] {
				out.WriteByte('>')
				continue
			}
			out.WriteByte('>')
		case htmltoken.EndTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)
			if skipDepth > 0 {
				if skipContentTags[tag] {
					skipDepth--
				}
				continue
			}
			if allowedMailHTMLTags[tag] && !voidMailHTMLTags[tag] {
				out.WriteString("</")
				out.WriteString(tag)
				out.WriteByte('>')
			}
		}
	}
}

var (
	allowedMailHTMLTags = map[string]bool{
		"a": true, "abbr": true, "b": true, "blockquote": true, "br": true, "code": true, "dd": true, "del": true,
		"div": true, "dl": true, "dt": true, "em": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true,
		"h6": true, "hr": true, "i": true, "img": true, "ins": true, "li": true, "ol": true, "p": true, "pre": true,
		"s": true, "small": true, "span": true, "strong": true, "sub": true, "sup": true, "table": true, "tbody": true,
		"td": true, "tfoot": true, "th": true, "thead": true, "tr": true, "u": true, "ul": true,
	}
	voidMailHTMLTags = map[string]bool{"br": true, "hr": true, "img": true}
	skipContentTags  = map[string]bool{
		"script": true, "style": true, "iframe": true, "object": true, "embed": true, "form": true, "textarea": true,
		"select": true, "svg": true, "math": true, "head": true,
	}
	globalMailHTMLAttrs = map[string]bool{"title": true, "dir": true, "lang": true}
)

func writeSanitizedAttributes(out *strings.Builder, tag string, attrs []htmltoken.Attribute, trustedDomains []string) int {
	blocked := 0
	wroteRel := false
	wroteTarget := false
	for _, attr := range attrs {
		name := strings.ToLower(strings.TrimSpace(attr.Key))
		value := stdhtml.UnescapeString(strings.TrimSpace(attr.Val))
		if name == "" || strings.HasPrefix(name, "on") || name == "style" || strings.Contains(name, ":") {
			continue
		}
		switch tag {
		case "a":
			switch name {
			case "href":
				if isSafeMailLink(value) {
					writeHTMLAttribute(out, "href", value)
				}
			case "target":
				if value == "_blank" {
					writeHTMLAttribute(out, "target", "_blank")
					wroteTarget = true
				}
			case "rel":
				wroteRel = true
				writeHTMLAttribute(out, "rel", "noopener noreferrer")
			default:
				if globalMailHTMLAttrs[name] {
					writeHTMLAttribute(out, name, value)
				}
			}
		case "img":
			switch name {
			case "src", "poster", "background":
				if isTrustedMailResource(value, trustedDomains) {
					writeHTMLAttribute(out, name, value)
				} else if isDeferrableExternalResource(value) {
					blocked++
					writeHTMLAttribute(out, "data-external-"+name, value)
				}
			case "srcset":
				if srcsetTrusted(value, trustedDomains) {
					writeHTMLAttribute(out, name, value)
				} else if srcsetDeferrableExternal(value, trustedDomains) {
					blocked++
					writeHTMLAttribute(out, "data-external-srcset", value)
				}
			case "alt", "width", "height", "title":
				if safePresentationAttribute(value) {
					writeHTMLAttribute(out, name, value)
				}
			}
		case "td", "th":
			if (name == "colspan" || name == "rowspan") && safeSmallIntegerAttribute(value) {
				writeHTMLAttribute(out, name, value)
			} else if globalMailHTMLAttrs[name] {
				writeHTMLAttribute(out, name, value)
			}
		default:
			if globalMailHTMLAttrs[name] {
				writeHTMLAttribute(out, name, value)
			}
		}
	}
	if tag == "a" {
		if !wroteTarget {
			writeHTMLAttribute(out, "target", "_blank")
		}
		if !wroteRel {
			writeHTMLAttribute(out, "rel", "noopener noreferrer")
		}
	}
	return blocked
}

func writeHTMLAttribute(out *strings.Builder, name, value string) {
	out.WriteByte(' ')
	out.WriteString(name)
	out.WriteString(`="`)
	out.WriteString(stdhtml.EscapeString(value))
	out.WriteByte('"')
}

func isSafeMailLink(raw string) bool {
	raw = strings.TrimSpace(stdhtml.UnescapeString(raw))
	if raw == "" || strings.HasPrefix(raw, "//") || strings.ContainsAny(raw, "\x00\r\n\t") {
		return false
	}
	if strings.HasPrefix(raw, "#") {
		return true
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return parsed.Hostname() != ""
	case "mailto", "tel":
		return true
	default:
		return false
	}
}

func safePresentationAttribute(value string) bool {
	return !strings.ContainsAny(value, "\x00\r\n")
}

func safeSmallIntegerAttribute(value string) bool {
	if value == "" || len(value) > 2 {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return value != "0"
}

func srcsetTrusted(value string, trustedDomains []string) bool {
	for _, candidate := range strings.Split(value, ",") {
		fields := strings.Fields(strings.TrimSpace(candidate))
		if len(fields) == 0 || !isTrustedMailResource(fields[0], trustedDomains) {
			return false
		}
	}
	return strings.TrimSpace(value) != ""
}

func srcsetDeferrableExternal(value string, trustedDomains []string) bool {
	hasExternal := false
	for _, candidate := range strings.Split(value, ",") {
		fields := strings.Fields(strings.TrimSpace(candidate))
		if len(fields) == 0 {
			return false
		}
		source := fields[0]
		if isTrustedMailResource(source, trustedDomains) {
			continue
		}
		if !isDeferrableExternalResource(source) {
			return false
		}
		hasExternal = true
	}
	return hasExternal
}

func isTrustedMailResource(raw string, trustedDomains []string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "cid:") || safeDataImage(lower) {
		return true
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return false
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	for _, domain := range trustedDomains {
		domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
		if domain != "" && (host == domain || strings.HasSuffix(host, "."+domain)) {
			return true
		}
	}
	return false
}

func isDeferrableExternalResource(raw string) bool {
	raw = strings.TrimSpace(stdhtml.UnescapeString(raw))
	if raw == "" || strings.HasPrefix(raw, "//") || strings.ContainsAny(raw, "\x00\r\n\t") {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return true
	default:
		return false
	}
}

func safeDataImage(value string) bool {
	return strings.HasPrefix(value, "data:image/png;") ||
		strings.HasPrefix(value, "data:image/jpeg;") ||
		strings.HasPrefix(value, "data:image/jpg;") ||
		strings.HasPrefix(value, "data:image/gif;") ||
		strings.HasPrefix(value, "data:image/webp;")
}

func trustedDomainsForMaildirRoot(root string) []string {
	localDir := filepath.Dir(root)
	domainDir := filepath.Dir(localDir)
	domain := filepath.Base(domainDir)
	if domain == "." || domain == string(filepath.Separator) || domain == "" {
		return nil
	}
	return []string{domain}
}

func htmlToText(markup string) string {
	markup = dangerousBlockRe.ReplaceAllString(markup, "")
	markup = breakTagRe.ReplaceAllString(markup, "\n")
	markup = blockEndTagRe.ReplaceAllString(markup, "\n")
	markup = htmlTagRe.ReplaceAllString(markup, "")
	markup = stdhtml.UnescapeString(markup)
	lines := strings.Split(strings.ReplaceAll(markup, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = whitespaceLineRe.ReplaceAllString(strings.TrimSpace(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}
	return blankLineRe.ReplaceAllString(strings.Join(out, "\n"), "\n\n")
}
