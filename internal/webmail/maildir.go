package webmail

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type MessageSummary struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
	Preview   string    `json:"preview"`
	Mailbox   string    `json:"mailbox"`
	SizeBytes int64     `json:"size_bytes"`
	Unread    bool      `json:"unread"`
}

type MessageDetail struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
	Body      string    `json:"body"`
	Mailbox   string    `json:"mailbox"`
	SizeBytes int64     `json:"size_bytes"`
	Unread    bool      `json:"unread"`
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
	case "spam":
		return []string{filepath.Join(".Spam", "new"), filepath.Join(".Spam", "cur")}
	case "trash":
		return []string{filepath.Join(".Trash", "new"), filepath.Join(".Trash", "cur")}
	case "archive":
		return []string{filepath.Join(".Archive", "new"), filepath.Join(".Archive", "cur")}
	case "sent":
		return []string{filepath.Join(".Sent", "new"), filepath.Join(".Sent", "cur")}
	case "all":
		return []string{"new", "cur", filepath.Join(".Sent", "new"), filepath.Join(".Sent", "cur"), filepath.Join(".Spam", "new"), filepath.Join(".Spam", "cur"), filepath.Join(".Trash", "new"), filepath.Join(".Trash", "cur"), filepath.Join(".Archive", "new"), filepath.Join(".Archive", "cur")}
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
	sort.SliceStable(folders[5:], func(i, j int) bool {
		return strings.ToLower(folders[5+i].Name) < strings.ToLower(folders[5+j].Name)
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
	date, _ := msg.Header.Date()
	rel, _ := filepath.Rel(maildirRoot, path)
	mailbox := strings.Split(filepath.ToSlash(rel), "/")[0]
	return MessageSummary{
		ID:        filepath.Base(path),
		From:      msg.Header.Get("From"),
		To:        msg.Header.Get("To"),
		Subject:   msg.Header.Get("Subject"),
		Date:      date,
		Preview:   preview(msg.Body),
		Mailbox:   mailbox,
		SizeBytes: info.Size(),
		Unread:    isUnreadPath(path),
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
	return os.Rename(source, filepath.Join(destinationDir, filepath.Base(source)))
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
	mailboxes := []string{"new", "cur", ".Sent/new", ".Sent/cur", ".Spam/new", ".Spam/cur", ".Trash/new", ".Trash/cur", ".Archive/new", ".Archive/cur"}
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
	return "", os.ErrNotExist
}

func destinationMailbox(folder string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(folder))
	switch name {
	case "inbox", "":
		return "new", nil
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
	info, err := os.Stat(path)
	if err != nil {
		return MessageDetail{}, err
	}
	date, _ := msg.Header.Date()
	rel, _ := filepath.Rel(maildirRoot, path)
	mailbox := strings.Split(filepath.ToSlash(rel), "/")[0]
	return MessageDetail{
		ID:        filepath.Base(path),
		From:      msg.Header.Get("From"),
		To:        msg.Header.Get("To"),
		Subject:   msg.Header.Get("Subject"),
		Date:      date,
		Body:      string(body),
		Mailbox:   mailbox,
		SizeBytes: info.Size(),
		Unread:    isUnreadPath(path),
	}, nil
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
